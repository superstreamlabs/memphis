// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"encoding/json"
	"fmt"
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

const (
	connectItemSep                      = "::"
	connectConfigUpdatesSubjectTemplate = CONFIGURATIONS_UPDATES_SUBJ + ".init.%s"
)

func updateNewClientWithConfig(c *client, connId string) {
	// TODO more configurations logic here

	slackEnabled, err := notifications.IsSlackEnabled()
	if err != nil {
		c.Errorf("updateNewClientWithConfig: " + err.Error())
	}

	config := models.GlobalConfigurationsUpdate{
		Notifications: slackEnabled,
	}

	sendConnectUpdate(c, config, connId)
}

func sendConnectUpdate(c *client, ccu models.GlobalConfigurationsUpdate, connId string) {
	s := c.srv
	rawMsg, err := json.Marshal(ccu)
	if err != nil {
		s.Errorf(err.Error())
		return
	}
	subject := fmt.Sprintf(connectConfigUpdatesSubjectTemplate, connId)
	s.sendInternalAccountMsg(c.acc, subject, rawMsg)
}

func handleConnectMessage(client *client) error {
	splittedMemphisInfo := strings.Split(client.opts.Name, connectItemSep)

	var (
		isNativeMemphisClient bool
		username              string
		objID                 primitive.ObjectID
	)
	switch len(splittedMemphisInfo) {
	case 2:
		// normal Memphis SDK carry on to the rest of the function
		isNativeMemphisClient = true
		username = strings.ToLower(splittedMemphisInfo[1])
	case 1:
		// NATS SDK, means we extract username from the token field
		isNativeMemphisClient = false
		splittedToken := strings.Split(client.opts.Token, connectItemSep)
		if len(splittedToken) != 2 {
			client.Warnf("handleConnectMessage: missing username or token")
			return errors.New("missing username or token")
		}
		username = strings.ToLower(splittedToken[0])
	default:
		client.Warnf("handleConnectMessage: missing username or connectionId")
		return errors.New("missing username or connectionId")
	}

	exist, user, err := IsUserExist(username)
	if err != nil {
		errMsg := "User " + username + ": " + err.Error()
		client.Errorf("handleConnectMessage: " + errMsg)
		return err
	}
	if !exist {
		errMsg := "User " + username + " does not exist"
		client.Warnf("handleConnectMessage: " + errMsg)
		return errors.New(errMsg)
	}
	if user.UserType != "root" && user.UserType != "application" {
		client.Warnf("handleConnectMessage: Please use a user of type Root/Application and not Management")
		return errors.New("Please use a user of type Root/Application and not Management")
	}

	if isNativeMemphisClient {
		objIdString := splittedMemphisInfo[0]
		objID, err = primitive.ObjectIDFromHex(objIdString)
		if err != nil {
			errMsg := "User " + username + ": " + err.Error()
			client.Errorf("handleConnectMessage: " + errMsg)
			return err
		}

		exist, _, err = IsConnectionExist(objID)
		if err != nil {
			errMsg := "User " + username + ": " + err.Error()
			client.Errorf("handleConnectMessage: " + errMsg)
			return err
		}

		if exist {
			err = connectionsHandler.ReliveConnection(primitive.ObjectID(objID))
			if err != nil {
				errMsg := "User " + username + ": " + err.Error()
				client.Errorf("handleConnectMessage: " + errMsg)
				return err
			}
			err = producersHandler.ReliveProducers(primitive.ObjectID(objID))
			if err != nil {
				errMsg := "User " + username + ": " + err.Error()
				client.Errorf("handleConnectMessage: " + errMsg)
				return err
			}
			err = consumersHandler.ReliveConsumers(primitive.ObjectID(objID))
			if err != nil {

				errMsg := "User " + username + ": " + err.Error()
				client.Errorf("handleConnectMessage: " + errMsg)
				return err
			}
		} else {
			err := connectionsHandler.CreateConnection(username, client.RemoteAddress().String(), objID)
			if err != nil {
				errMsg := "User " + username + ": " + err.Error()
				client.Errorf("handleConnectMessage: " + errMsg)
				return err
			}
		}
		updateNewClientWithConfig(client, objIdString)
	}

	client.memphisInfo = memphisClientInfo{username: username, connectionId: objID, isNative: isNativeMemphisClient}
	if username == "" {
		client.Errorf(client.opts.Name, connectItemSep)
	}
 	shouldSendAnalytics, _ := shouldSendAnalytics()
	if !exist && shouldSendAnalytics { // exist indicates it is a reconnect
		splitted := strings.Split(client.opts.Lang, ".")
		sdkName := splitted[len(splitted)-1]
		param := analytics.EventParam{
			Name:  "sdk",
			Value: sdkName,
		}
		analyticsParams := []analytics.EventParam{param}
		event := "user-connect-sdk"
		if !isNativeMemphisClient {
			event = "user-connect-nats-sdk"
		}
		analytics.SendEventWithParams(username, analyticsParams, event)
	}

	return nil
}

func (ch ConnectionsHandler) CreateConnection(username, clientAddress string, connectionId primitive.ObjectID) error {
	username = strings.ToLower(username)
	exist, _, err := IsUserExist(username)
	if err != nil {
		errMsg := "User " + username + ": " + err.Error()
		serv.Errorf("CreateConnection error: " + errMsg)
		return err
	}
	if !exist {
		errMsg := "User " + username + " does not exist"
		return errors.New(errMsg)
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
		errMsg := "User " + username + ": " + err.Error()
		serv.Errorf("CreateConnection error: " + errMsg)
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
			consumerNames = consumerNames + "Consumer: " + consumers[i].Name + " | Station: " + consumers[i].StationName + "\n"
		}
	}
	msg := ""
	if len(producerNames) > 0 {
		msg = producerNames
	}
	if len(consumerNames) > 0 {
		msg = msg + consumerNames
	}
	err = notifications.SendNotification("Disconnection events", msg, notifications.DisconEAlert)
	if err != nil {
		return err
	}

	serv.Noticef("Client has been disconnected from Memphis")
	return err
}
