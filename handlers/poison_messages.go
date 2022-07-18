// Copyright 2021-2022 The Memphis Authors
// Licensed under the GNU General Public License v3.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"encoding/json"
	"memphis-broker/broker"
	"memphis-broker/logger"
	"memphis-broker/models"

	"context"
	"time"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PoisonMessagesHandler struct{}

func (pmh PoisonMessagesHandler) HandleNewMessage(msg *nats.Msg) {
	var message map[string]interface{}
	err := json.Unmarshal(msg.Data, &message)
	if err != nil {
		logger.Error("Error while getting notified about a poison message: " + err.Error())
		return
	}

	stationName := message["stream"].(string)
	cgName := message["consumer"].(string)
	messageSeq := message["stream_seq"].(float64)
	deliveriesCount := message["deliveries"].(float64)

	poisonMessageContent, err := broker.GetMessage(stationName, uint64(messageSeq))
	if err != nil {
		logger.Error("Error while getting notified about a poison message: " + err.Error())
		return
	}

	connectionIdHeader := poisonMessageContent.Header.Get("connectionId")
	producedByHeader := poisonMessageContent.Header.Get("producedBy")

	if connectionIdHeader == "" || producedByHeader == "" {
		logger.Error("Error while getting notified about a poison message: Missing mandatory message headers, please upgrade the SDk version you are using")
		return
	}

	if producedByHeader == "$memphis_dlq" { // skip poison messages which have been resent
		return
	}

	connId, _ := primitive.ObjectIDFromHex(connectionIdHeader)
	_, conn, err := IsConnectionExist(connId)
	if err != nil {
		logger.Error("Error while getting notified about a poison message: " + err.Error())
		return
	}

	producer := models.ProducerDetails{
		Name:          producedByHeader,
		ClientAddress: conn.ClientAddress,
		ConnectionId:  connId,
	}

	messagePayload := models.MessagePayload{
		TimeSent: poisonMessageContent.Time,
		Size:     len(poisonMessageContent.Subject) + len(poisonMessageContent.Data) + broker.GetHeaderSizeInBytes(poisonMessageContent.Header),
		Data:     string(poisonMessageContent.Data),
	}
	poisonedCg := models.PoisonedCg{
		CgName:          cgName,
		PoisoningTime:   time.Now(),
		DeliveriesCount: int(deliveriesCount),
	}
	filter := bson.M{
		"station_name":      stationName,
		"message_seq":       int(messageSeq),
		"producer.name":     producedByHeader,
		"message.time_sent": poisonMessageContent.Time,
	}
	update := bson.M{
		"$push": bson.M{"poisoned_cgs": poisonedCg},
		"$set":  bson.M{"message": messagePayload, "producer": producer, "creation_date": time.Now()},
	}
	opts := options.Update().SetUpsert(true)
	_, err = poisonMessagesCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		logger.Error("Error while getting notified about a poison message: " + err.Error())
		return
	}
}

func (pmh PoisonMessagesHandler) GetPoisonMsgsByStation(station models.Station) ([]models.LightPoisonMessage, error) {
	poisonMessages := make([]models.LightPoisonMessage, 0)

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"creation_date": -1})
	findOptions.SetLimit(1000) // fetch the last 1000
	cursor, err := poisonMessagesCollection.Find(context.TODO(), bson.M{
		"station_name": station.Name,
	}, findOptions)
	if err != nil {
		return poisonMessages, err
	}

	if err = cursor.All(context.TODO(), &poisonMessages); err != nil {
		return poisonMessages, err
	}

	return poisonMessages, nil
}

func GetPoisonMsgById(messageId primitive.ObjectID) (models.PoisonMessage, error) {
	var poisonMessage models.PoisonMessage
	err := poisonMessagesCollection.FindOne(context.TODO(), bson.M{
		"_id": messageId,
	}).Decode(&poisonMessage)
	if err != nil {
		return poisonMessage, err
	}

	return poisonMessage, nil
}

func RemovePoisonMsgsByStation(stationName string) error {
	_, err := poisonMessagesCollection.DeleteMany(context.TODO(), bson.M{"station_name": stationName})
	if err != nil {
		return err
	}
	return nil
}

func RemovePoisonedCg(stationName, cgName string) error {
	_, err := poisonMessagesCollection.UpdateMany(context.TODO(),
		bson.M{"station_name": stationName},
		bson.M{"$pull": bson.M{"poisoned_cgs": bson.M{"cg_name": cgName}}},
	)
	if err != nil {
		return err
	}
	return nil
}

func GetTotalPoisonMsgsByCg(stationName, cgName string) (int, error) {
	count, err := poisonMessagesCollection.CountDocuments(context.TODO(), bson.M{
		"station_name":         stationName,
		"poisoned_cgs.cg_name": cgName,
	})
	if err != nil {
		return -1, err
	}

	return int(count), nil
}

func GetPoisonedCgsByMessage(stationName string, message models.MessageDetails) ([]models.PoisonedCg, error) {
	var poisonMessage models.PoisonMessage
	err := poisonMessagesCollection.FindOne(context.TODO(), bson.M{
		"station_name":      stationName,
		"message_seq":       message.MessageSeq,
		"producer.name":     message.ProducedBy,
		"message.time_sent": message.TimeSent,
	}).Decode(&poisonMessage)
	if err == mongo.ErrNoDocuments {
		return []models.PoisonedCg{}, nil
	}
	if err != nil {
		return []models.PoisonedCg{}, err
	}

	return poisonMessage.PoisonedCgs, nil
}
