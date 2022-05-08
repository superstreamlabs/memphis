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

package utils

import (
	"context"
	"fmt"
	"memphis-control-plane/db"
	"memphis-control-plane/logger"
	"memphis-control-plane/models"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var connectionsCollection *mongo.Collection = db.GetCollection("connections")
var producersCollection *mongo.Collection = db.GetCollection("producers")
var consumersCollection *mongo.Collection = db.GetCollection("consumers")

func killRelevantConnections() ([]models.Connection, error) {
	lastAllowedTime := time.Now().Add(time.Duration(-configuration.PING_INTERVAL_MS-5000) * time.Millisecond)

	fmt.Println("lastAllowedTime: ", lastAllowedTime)

	var connections []models.Connection // "is_active": true, 
	cursor, err := connectionsCollection.Find(context.TODO(), bson.M{"last_ping": bson.M{"$lt": lastAllowedTime}})
	if err != nil {
		logger.Error("killRelevantConnections error: " + err.Error())
		return connections, err
	}

	if err = cursor.All(context.TODO(), &connections); err != nil {
		logger.Error("killRelevantConnections error: " + err.Error())
		return connections, err
	}

	fmt.Println(connections)

	_, err = connectionsCollection.UpdateMany(context.TODO(),
		bson.M{"is_active": true, "last_ping": bson.M{"$lt": lastAllowedTime}},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		logger.Error("KillConnections error: " + err.Error())
		return connections, err
	}

	var connectionIds []primitive.ObjectID
	for _, con := range connections {
		connectionIds = append(connectionIds, con.ID)
	}

	return connections, nil
}

func killProducersByConnections(connectionIds []primitive.ObjectID) error {
	_, err := producersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": bson.M{"$in": connectionIds}},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		logger.Error("killProducersByConnections error: " + err.Error())
		return err
	}

	return nil
}

func killConsumersByConnections(connectionIds []primitive.ObjectID) error {
	_, err := consumersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": bson.M{"$in": connectionIds}},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		logger.Error("killConsumersByConnections error: " + err.Error())
		return err
	}

	return nil
}

func KillZombieConnections(wg *sync.WaitGroup) {
	for range time.Tick(time.Second * 30) {
		connections, err := killRelevantConnections()
		if err != nil {
			logger.Error("KillZombieConnections error: " + err.Error())
		} else if len(connections) > 0 {
			var connectionIds []primitive.ObjectID
			for _, con := range connections {
				connectionIds = append(connectionIds, con.ID)
			}

			err = killProducersByConnections(connectionIds)
			if err != nil {
				logger.Error("KillZombieConnections error: " + err.Error())
			}

			err = killConsumersByConnections(connectionIds)
			if err != nil {
				logger.Error("KillZombieConnections error: " + err.Error())
			}
		}
	}

	defer wg.Done()
}
