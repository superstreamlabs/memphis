// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package server

import (
	"encoding/json"
	"memphis-broker/models"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PoisonMessagesHandler struct{ S *Server }

func (s *Server) ListenForPoisonMessages() {
	s.queueSubscribe("$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.>",
		"$memphis_poison_messages_listeners_group",
		createPoisonMessageHandler(s))
}

func createPoisonMessageHandler(s *Server) simplifiedMsgHandler {
	return func(_ *client, _, _ string, msg []byte) {
		go s.HandleNewMessage(msg)
	}
}

func (s *Server) HandleNewMessage(msg []byte) {
	var message map[string]interface{}
	err := json.Unmarshal(msg, &message)
	if err != nil {
		serv.Errorf("Error while getting notified about a poison message: " + err.Error())
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
		serv.Errorf("Error while getting notified about a poison message: " + err.Error())
		return
	}

	headersJson, err := DecodeHeader(poisonMessageContent.Header)

	if err != nil {
		serv.Errorf(err.Error())
		return
	}
	connectionIdHeader := headersJson["$memphis_connectionId"]
	producedByHeader := headersJson["$memphis_producedBy"]

	//This check for backward compatability
	if connectionIdHeader == "" || producedByHeader == "" {
		connectionIdHeader = headersJson["connectionId"]
		producedByHeader = headersJson["producedBy"]
		if connectionIdHeader == "" || producedByHeader == "" {
			serv.Warnf("Error while getting notified about a poison message: Missing mandatory message headers, please upgrade the SDK version you are using")
			return
		}
	}

	if producedByHeader == "$memphis_dlq" { // skip poison messages which have been resent
		return
	}

	connId, _ := primitive.ObjectIDFromHex(connectionIdHeader)
	_, conn, err := IsConnectionExist(connId)
	if err != nil {
		serv.Errorf("Error while getting notified about a poison message: " + err.Error())
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
	update := bson.M{
		"$push": bson.M{"poisoned_cgs": poisonedCg},
		"$set":  bson.M{"message": messagePayload, "producer": producer, "creation_date": time.Now()},
	}
	opts := options.Update().SetUpsert(true)
	_, err = poisonMessagesCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		serv.Errorf("Error while getting notified about a poison message: " + err.Error())
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
		if len(msg.Message.Data) > 100 {
			poisonMessages[i].Message.Data = msg.Message.Data[0:100]
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
