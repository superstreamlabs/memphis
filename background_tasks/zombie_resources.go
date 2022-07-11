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

package background_tasks

import (
	"context"
	"memphis-broker/config"
	"memphis-broker/db"
	"memphis-broker/logger"
	"memphis-broker/models"
	"strconv"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var configuration = config.GetConfig()

var connectionsCollection *mongo.Collection = db.GetCollection("connections")
var producersCollection *mongo.Collection = db.GetCollection("producers")
var consumersCollection *mongo.Collection = db.GetCollection("consumers")
var sysLogsCollection *mongo.Collection = db.GetCollection("system_logs")
var poisonMessagesCollection *mongo.Collection = db.GetCollection("poison_messages")

func killRelevantConnections() ([]models.Connection, error) {
	lastAllowedTime := time.Now().Add(time.Duration(-configuration.PING_INTERVAL_MS-5000) * time.Millisecond)

	var connections []models.Connection
	cursor, err := connectionsCollection.Find(context.TODO(), bson.M{"is_active": true, "last_ping": bson.M{"$lt": lastAllowedTime}})
	if err != nil {
		logger.Error("killRelevantConnections error: " + err.Error())
		return connections, err
	}

	if err = cursor.All(context.TODO(), &connections); err != nil {
		logger.Error("killRelevantConnections error: " + err.Error())
		return connections, err
	}

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

func removeOldLogs() error {
	retentionToInt, err := strconv.Atoi(configuration.LOGS_RETENTION_IN_DAYS)
	if err != nil {
		return err
	}
	retentionDaysToHours := 24 * retentionToInt
	filter := bson.M{"creation_date": bson.M{"$lte": (time.Now().Add(-(time.Hour * time.Duration(retentionDaysToHours))))}}
	_, err = sysLogsCollection.DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}

func removeOldPoisonMsgs() error {
	filter := bson.M{"creation_date": bson.M{"$lte": (time.Now().Add(-(time.Hour * time.Duration(configuration.POISON_MSGS_RETENTION_IN_HOURS))))}}
	_, err := poisonMessagesCollection.DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}

func KillZombieResources(wg *sync.WaitGroup) {
	for range time.Tick(time.Second * 30) {
		connections, err := killRelevantConnections()
		if err != nil {
			logger.Error("KillZombieResources error: " + err.Error())
		} else if len(connections) > 0 {
			var connectionIds []primitive.ObjectID
			for _, con := range connections {
				connectionIds = append(connectionIds, con.ID)
			}

			err = killProducersByConnections(connectionIds)
			if err != nil {
				logger.Error("KillZombieResources error: " + err.Error())
			}

			err = killConsumersByConnections(connectionIds)
			if err != nil {
				logger.Error("KillZombieResources error: " + err.Error())
			}
		}

		err = removeOldLogs()
		if err != nil {
			logger.Error("KillZombieResources error: " + err.Error())
		}

		err = removeOldPoisonMsgs()
		if err != nil {
			logger.Error("KillZombieResources error: " + err.Error())
		}
	}

	defer wg.Done()
}
