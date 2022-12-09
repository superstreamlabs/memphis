// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server
package server

import (
	"encoding/hex"
	"encoding/json"
	"memphis-broker/models"
	"memphis-broker/notifications"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PoisonMessageTitle = "Poison message"

type PoisonMessagesHandler struct{ S *Server }

func (s *Server) ListenForPoisonMessages() {
	s.queueSubscribe("$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.>",
		"$memphis_poison_messages_listeners_group",
		createPoisonMessageHandler(s))
}

func createPoisonMessageHandler(s *Server) simplifiedMsgHandler {
	return func(_ *client, _, _ string, msg []byte) {
		go s.HandleNewMessage(copyBytes(msg))
	}
}

func (s *Server) HandleNewMessage(msg []byte) {
	var message map[string]interface{}
	err := json.Unmarshal(msg, &message)
	if err != nil {
		serv.Errorf("HandleNewMessage: Error while getting notified about a poison message: " + err.Error())
		return
	}

	streamName := message["stream"].(string)
	stationName := StationNameFromStreamName(streamName)
	cgName := message["consumer"].(string)
	cgName = revertDelimiters(cgName)
	messageSeq := message["stream_seq"].(float64)
	deliveriesCount := message["deliveries"].(float64)

	poisonMessageContent, err := s.memphisGetMessage(stationName.Intern(), uint64(messageSeq))
	if err != nil {
		serv.Errorf("HandleNewMessage: Error while getting notified about a poison message: " + err.Error())
		return
	}

	headersJson, err := DecodeHeader(poisonMessageContent.Header)

	if err != nil {
		serv.Errorf("HandleNewMessage: " + err.Error())
		return
	}
	connectionIdHeader := headersJson["$memphis_connectionId"]
	producedByHeader := headersJson["$memphis_producedBy"]

	//This check for backward compatability
	if connectionIdHeader == "" || producedByHeader == "" {
		connectionIdHeader = headersJson["connectionId"]
		producedByHeader = headersJson["producedBy"]
		if connectionIdHeader == "" || producedByHeader == "" {
			serv.Warnf("HandleNewMessage: Error while getting notified about a poison message: Missing mandatory message headers, please upgrade the SDK version you are using")
			return
		}
	}

	if producedByHeader == "$memphis_dlq" { // skip poison messages which have been resent
		return
	}

	connId, _ := primitive.ObjectIDFromHex(connectionIdHeader)
	_, conn, err := IsConnectionExist(connId)
	if err != nil {
		serv.Errorf("HandleNewMessage: Error while getting notified about a poison message: " + err.Error())
		return
	}

	producer := models.ProducerDetails{
		Name:          producedByHeader,
		ClientAddress: conn.ClientAddress,
		ConnectionId:  connId,
	}

	var headers []models.MsgHeader
	for key, value := range headersJson {
		header := models.MsgHeader{HeaderKey: key, HeaderValue: value}
		headers = append(headers, header)
	}

	messagePayload := models.MessagePayloadDb{
		TimeSent: poisonMessageContent.Time,
		Size:     len(poisonMessageContent.Subject) + len(poisonMessageContent.Data) + len(poisonMessageContent.Header),
		Data:     string(poisonMessageContent.Data),
		Headers:  headers,
	}
	poisonedCg := models.PoisonedCg{
		CgName:          cgName,
		PoisoningTime:   time.Now(),
		DeliveriesCount: int(deliveriesCount),
	}
	filter := bson.M{
		"station_name":      stationName.Ext(),
		"message_seq":       int(messageSeq),
		"producer.name":     producedByHeader,
		"message.time_sent": poisonMessageContent.Time,
	}
	var newID = primitive.NewObjectID()
	update := bson.M{
		"$setOnInsert": bson.M{"_id": newID},
		"$push":        bson.M{"poisoned_cgs": poisonedCg},
		"$set":         bson.M{"message": messagePayload, "producer": producer, "creation_date": time.Now()},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true)
	res := poisonMessagesCollection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	var poisonMsg models.PoisonMessage
	var idForUrl string
	err = res.Decode(&poisonMsg)
	if err == mongo.ErrNoDocuments {
		idForUrl = newID.Hex()
	} else if err != nil {
		serv.Errorf("HandleNewMessage: Error while getting notified about a poison message: " + err.Error())
		return
	} else {
		idForUrl = poisonMsg.ID.Hex()
	}
	if UI_url == "" {
		return
	}
	var msgUrl = UI_url + "/stations/" + stationName.Ext() + "/" + idForUrl
	err = notifications.SendNotification(PoisonMessageTitle, "Poison message has been identified, for more details head to: "+msgUrl, notifications.PoisonMAlert)
	if err != nil {
		serv.Warnf("HandleNewMessage: Error while sending a poison message notification: " + err.Error())
		return
	}
}

func (pmh PoisonMessagesHandler) GetPoisonMsgsByStation(station models.Station) ([]models.LightPoisonMessageResponse, error) {
	poisonMessages := make([]models.LightPoisonMessage, 0)
	poisonMessagesResponse := make([]models.LightPoisonMessageResponse, 0)

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"creation_date": -1})
	findOptions.SetLimit(1000) // fetch the last 1000
	cursor, err := poisonMessagesCollection.Find(context.TODO(), bson.M{
		"station_name": station.Name,
	}, findOptions)
	if err != nil {
		return []models.LightPoisonMessageResponse{}, err
	}

	if err = cursor.All(context.TODO(), &poisonMessages); err != nil {
		return []models.LightPoisonMessageResponse{}, err
	}

	for i, msg := range poisonMessages {
		msgData := hex.EncodeToString([]byte(msg.Message.Data))
		if len(msgData) > 40 {
			poisonMessages[i].Message.Data = msgData[0:40]
		} else {
			poisonMessages[i].Message.Data = msgData
		}

		msg := models.MessagePayload{
			TimeSent: poisonMessages[i].Message.TimeSent,
			Size:     poisonMessages[i].Message.Size,
			Data:     poisonMessages[i].Message.Data,
		}
		poisonMessagesResponse = append(poisonMessagesResponse, models.LightPoisonMessageResponse{
			ID:      poisonMessages[i].ID,
			Message: msg,
		})
	}
	return poisonMessagesResponse, nil
}

func (pmh PoisonMessagesHandler) GetTotalPoisonMsgsByStation(stationName string) (int, error) {

	count, err := poisonMessagesCollection.CountDocuments(context.TODO(), bson.M{
		"station_name": stationName,
	})
	if err != nil {
		return int(count), err
	}
	return int(count), nil
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

func RemovePoisonedCg(stationName StationName, cgName string) error {
	_, err := poisonMessagesCollection.UpdateMany(context.TODO(),
		bson.M{"station_name": stationName.Ext()},
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

func GetPoisonedCgsByMessage(stationNameExt string, message models.MessageDetails) ([]models.PoisonedCg, error) {
	var poisonMessage models.PoisonMessage
	err := poisonMessagesCollection.FindOne(context.TODO(), bson.M{
		"station_name":      stationNameExt,
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
