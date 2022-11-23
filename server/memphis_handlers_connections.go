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
	"memphis-broker/analytics"
	"memphis-broker/models"
	"memphis-broker/notifications"

	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ConnectionsHandler struct{}

var connectionsHandler ConnectionsHandler
var producersHandler ProducersHandler
var consumersHandler ConsumersHandler

func handleConnectMessage(client *client) error {
	splittedMemphisInfo := strings.Split(client.opts.Name, "::")
	if len(splittedMemphisInfo) != 2 {
		client.Warnf("handleConnectMessage: missing username or connectionId")
		return errors.New("missing username or connectionId")
	}
	objIdString := splittedMemphisInfo[0]
	username := strings.ToLower(splittedMemphisInfo[1])

	exist, user, err := IsUserExist(username)
	if err != nil {
		client.Errorf("handleConnectMessage: " + err.Error())
		return err
	}
	if !exist {
		client.Warnf("handleConnectMessage: User is not exist")
		return errors.New("User is not exist")
	}
	if user.UserType != "root" && user.UserType != "application" {
		client.Warnf("handleConnectMessage: Please use a user of type Root/Application and not Management")
		return errors.New("Please use a user of type Root/Application and not Management")
	}

	objID, err := primitive.ObjectIDFromHex(objIdString)
	if err != nil {
		return err
	}

	exist, _, err = IsConnectionExist(objID)
	if err != nil {
		client.Errorf("handleConnectMessage: " + err.Error())
		return err
	}

	clientAddress := client.RemoteAddress().String()

	if exist {
		err = connectionsHandler.ReliveConnection(primitive.ObjectID(objID))
		if err != nil {
			client.Errorf("handleConnectMessage: " + err.Error())
			return err
		}
		err = producersHandler.ReliveProducers(primitive.ObjectID(objID))
		if err != nil {
			client.Errorf("handleConnectMessage: " + err.Error())
			return err
		}
		err = consumersHandler.ReliveConsumers(primitive.ObjectID(objID))
		if err != nil {
			client.Errorf("handleConnectMessage: " + err.Error())
			return err
		}
	} else {
		err := connectionsHandler.CreateConnection(username, clientAddress, objID)
		if err != nil {
			client.Errorf("handleConnectMessage: " + err.Error())
			return err
		}
	}
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if !shouldSendAnalytics {
		splitted := strings.Split(client.opts.Lang, ".")
		sdkName := splitted[len(splitted)-1]
		param := analytics.EventParam{
			Name:  "sdk",
			Value: sdkName,
		}
		analyticsParams := []analytics.EventParam{param}
		analytics.SendEventWithParams(user.Username, analyticsParams, "user-connect-sdk")
	}

	client.memphisInfo = memphisClientInfo{username: username, connectionId: objID}
	return nil
}

func (ch ConnectionsHandler) CreateConnection(username, clientAddress string, connectionId primitive.ObjectID) error {
	username = strings.ToLower(username)
	exist, _, err := IsUserExist(username)
	if err != nil {
		serv.Errorf("CreateConnection error: " + err.Error())
		return err
	}
	if !exist {
		return errors.New("User was not found")
	}

	newConnection := models.Connection{
		ID:            connectionId,
		CreatedByUser: username,
		IsActive:      true,
		CreationDate:  time.Now(),
		ClientAddress: clientAddress,
	}

	_, err = connectionsCollection.InsertOne(context.TODO(), newConnection)
	if err != nil {
		serv.Errorf("CreateConnection error: " + err.Error())
		return err
	}
	return nil
}

func (ch ConnectionsHandler) ReliveConnection(connectionId primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateOne(context.TODO(),
		bson.M{"_id": connectionId},
		bson.M{"$set": bson.M{"is_active": true}},
	)
	if err != nil {
		serv.Errorf("ReliveConnection error: " + err.Error())
		return err
	}

	return nil
}

func (mci *memphisClientInfo) updateDisconnection() error {
	if mci.connectionId.IsZero() {
		return nil
	}

	ctx := context.TODO()
	_, err := connectionsCollection.UpdateOne(ctx,
		bson.M{"_id": mci.connectionId},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		return err
	}
	var producers []models.ExtendedProducer
	cursor, err := producersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"connection_id", mci.connectionId}, {"is_active", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	})
	if err != nil {
		return err
	}

	if err = cursor.All(context.TODO(), &producers); err != nil {
		return err
	}
	var producerNames, consumerNames string
	if len(producers) > 0 {
		_, err = producersCollection.UpdateMany(ctx,
			bson.M{"connection_id": mci.connectionId},
			bson.M{"$set": bson.M{"is_active": false}},
		)
		if err != nil {
			return err
		}

		for i := 0; i < len(producers); i++ {
			producerNames = producerNames + "Producer: " + producers[i].Name + " Station: " + producers[i].StationName + "\n"
		}
	}

	var consumers []models.ExtendedConsumer
	cursor, err = consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"connection_id", mci.connectionId}, {"is_active", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}}})
	if err != nil {
		return err
	}
	if err = cursor.All(context.TODO(), &consumers); err != nil {
		return err
	}
	if len(consumers) > 0 {
		_, err = consumersCollection.UpdateMany(ctx,
			bson.M{"connection_id": mci.connectionId},
			bson.M{"$set": bson.M{"is_active": false}},
		)
		if err != nil {
			return err
		}

		for i := 0; i < len(consumers); i++ {
			consumerNames = consumerNames + "Consumer: " + consumers[i].Name + " Station: " + consumers[i].StationName + "\n"
		}
	}
	slackIntegration, ok := notifications.NotificationIntegrationsMap["slack"].(models.SlackIntegration)
	if ok {
		if slackIntegration.Properties["disconnection_events_alert"] {
			msg := ""
			if len(producerNames) > 0 {
				msg = producerNames
			}
			if len(consumerNames) > 0 {
				msg = msg + consumerNames
			}
			notifications.SendMessageToSlackChannel("Disconnection events", msg)
		}
	}

	serv.Noticef("Client has been disconnected from Memphis")
	return err
}
