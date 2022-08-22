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
// limitations under the License.

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
		client.Errorf("handleConnectMessage: missing username or connectionId")
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
		return errors.New("User is not exist")
	}
	if user.UserType != "root" && user.UserType != "application" {
		return errors.New("Please use a user of type Root/Application and not Management")
	}

	var objID primitive.ObjectID

	if objIdString != "" {
		objID, err := primitive.ObjectIDFromHex(objIdString)
		if err != nil {
			return err
		}
		exist, _, err = IsConnectionExist(objID)
		if err != nil {
			client.Errorf("handleConnectMessage: " + err.Error())
			return err
		}
		client.memphisInfo.ConnectionId = objID
	} else {
		exist = false
	}

	clientAddress := client.host
	clientAddress = strings.Split(clientAddress, ":")[0]
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
		connectionId, err := connectionsHandler.CreateConnection(username, clientAddress)
		if err != nil {
			client.Errorf("handleConnectMessage: " + err.Error())
			return err
		}
		client.memphisInfo.ConnectionId = connectionId
	}

	return nil
}

func (ch ConnectionsHandler) CreateConnection(username string, clientAddress string) (primitive.ObjectID, error) {
	connectionId := primitive.NewObjectID()

	username = strings.ToLower(username)
	exist, _, err := IsUserExist(username)
	if err != nil {
		serv.Errorf("CreateConnection error: " + err.Error())
		return connectionId, err
	}
	if !exist {
		return connectionId, errors.New("User was not found")
	}

	newConnection := models.Connection{
		ID:            connectionId,
		CreatedByUser: username,
		IsActive:      true,
		CreationDate:  time.Now(),
		LastPing:      time.Now(),
		ClientAddress: clientAddress,
	}

	_, err = connectionsCollection.InsertOne(context.TODO(), newConnection)
	if err != nil {
		serv.Errorf("CreateConnection error: " + err.Error())
		return connectionId, err
	}
	return connectionId, nil
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

func (ch ConnectionsHandler) UpdatePingTime(connectionId primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateOne(context.TODO(),
		bson.M{"_id": connectionId},
		bson.M{"$set": bson.M{"last_ping": time.Now()}},
	)
	if err != nil {
		serv.Errorf("UpdatePingTime error: " + err.Error())
		return err
	}

	return nil
}
