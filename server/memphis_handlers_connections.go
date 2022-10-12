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
	"memphis-broker/models"

	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (ch ConnectionsHandler) KillConnection(connectionId primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateOne(context.TODO(),
		bson.M{"_id": connectionId},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		serv.Errorf("KillConnection error: " + err.Error())
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
	_, err = producersCollection.UpdateMany(ctx,
		bson.M{"connection_id": mci.connectionId},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		return err
	}
	_, err = consumersCollection.UpdateMany(ctx,
		bson.M{"connection_id": mci.connectionId},
		bson.M{"$set": bson.M{"is_active": false}},
	)

	return err
}
