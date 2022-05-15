// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"context"
	"errors"
	"memphis-control-plane/logger"
	"memphis-control-plane/models"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConnectionsHandler struct{}

func (umh ConnectionsHandler) CreateConnection(username string) (primitive.ObjectID, error) {
	connectionId := primitive.NewObjectID()

	username = strings.ToLower(username)
	exist, _, err := IsUserExist(username)
	if err != nil {
		logger.Error("CreateProducer error: " + err.Error())
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
	}

	_, err = connectionsCollection.InsertOne(context.TODO(), newConnection)
	if err != nil {
		logger.Error("CreateConnection error: " + err.Error())
		return connectionId, err
	}
	return connectionId, nil
}

func (umh ConnectionsHandler) KillConnection(connectionId primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateOne(context.TODO(),
		bson.M{"_id": connectionId},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		logger.Error("KillConnection error: " + err.Error())
		return err
	}

	return nil
}

func (umh ConnectionsHandler) ReliveConnection(connectionId primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateOne(context.TODO(),
		bson.M{"_id": connectionId},
		bson.M{"$set": bson.M{"is_active": true}},
	)
	if err != nil {
		logger.Error("ReliveConnection error: " + err.Error())
		return err
	}

	return nil
}

func (umh ConnectionsHandler) UpdatePingTime(connectionId primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateOne(context.TODO(),
		bson.M{"_id": connectionId},
		bson.M{"$set": bson.M{"last_ping": time.Now()}},
	)
	if err != nil {
		logger.Error("UpdatePingTime error: " + err.Error())
		return err
	}

	return nil
}
