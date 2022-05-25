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
	"context"
	"memphis-control-plane/broker"
	"memphis-control-plane/config"
	"memphis-control-plane/db"
	"memphis-control-plane/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var usersCollection *mongo.Collection = db.GetCollection("users")
var tokensCollection *mongo.Collection = db.GetCollection("tokens")
var imagesCollection *mongo.Collection = db.GetCollection("images")
var factoriesCollection *mongo.Collection = db.GetCollection("factories")
var stationsCollection *mongo.Collection = db.GetCollection("stations")
var connectionsCollection *mongo.Collection = db.GetCollection("connections")
var producersCollection *mongo.Collection = db.GetCollection("producers")
var consumersCollection *mongo.Collection = db.GetCollection("consumers")
var systemKeysCollection *mongo.Collection = db.GetCollection("system_keys")
var configuration = config.GetConfig()

func getUserDetailsFromMiddleware(c *gin.Context) models.User {
	user, _ := c.Get("user")
	return user.(models.User)
}

func IsUserExist(username string) (bool, models.User, error) {
	filter := bson.M{"username": username}
	var user models.User
	err := usersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return false, user, nil
	} else if err != nil {
		return false, user, err
	}
	return true, user, nil
}

func IsFactoryExist(factoryName string) (bool, models.Factory, error) {
	filter := bson.M{"name": factoryName}
	var factory models.Factory
	err := factoriesCollection.FindOne(context.TODO(), filter).Decode(&factory)
	if err == mongo.ErrNoDocuments {
		return false, factory, nil
	} else if err != nil {
		return false, factory, err
	}
	return true, factory, nil
}

func IsStationExist(stationName string) (bool, models.Station, error) {
	filter := bson.M{"name": stationName}
	var station models.Station
	err := stationsCollection.FindOne(context.TODO(), filter).Decode(&station)
	if err == mongo.ErrNoDocuments {
		return false, station, nil
	} else if err != nil {
		return false, station, err
	}
	return true, station, nil
}

func IsConnectionExist(connectionId primitive.ObjectID) (bool, models.Connection, error) {
	filter := bson.M{"_id": connectionId}
	var connection models.Connection
	err := connectionsCollection.FindOne(context.TODO(), filter).Decode(&connection)
	if err == mongo.ErrNoDocuments {
		return false, connection, nil
	} else if err != nil {
		return false, connection, err
	}
	return true, connection, nil
}

func IsConsumerExist(consumerName string, stationId primitive.ObjectID) (bool, models.Consumer, error) {
	filter := bson.M{"name": consumerName, "station_id": stationId, "is_active": true}
	var consumer models.Consumer
	err := consumersCollection.FindOne(context.TODO(), filter).Decode(&consumer)
	if err == mongo.ErrNoDocuments {
		return false, consumer, nil
	} else if err != nil {
		return false, consumer, err
	}
	return true, consumer, nil
}

func IsProducerExist(producerName string, stationId primitive.ObjectID) (bool, models.Producer, error) {
	filter := bson.M{"name": producerName, "station_id": stationId, "is_active": true}
	var producer models.Producer
	err := producersCollection.FindOne(context.TODO(), filter).Decode(&producer)
	if err == mongo.ErrNoDocuments {
		return false, producer, nil
	} else if err != nil {
		return false, producer, err
	}
	return true, producer, nil
}

func CreateDefaultStation(stationName string, username string) (models.Station, error) {
	var newStation models.Station
	
	// create default factory
	var factoryId primitive.ObjectID
	exist, factory, err := IsFactoryExist("melvis")
	if err != nil {
		return newStation, err
	}
	if !exist {
		factoryId = primitive.NewObjectID()
		newFactory := models.Factory{
			ID:            factoryId,
			Name:          "melvis",
			Description:   "",
			CreatedByUser: username,
			CreationDate:  time.Now(),
		}

		_, err := factoriesCollection.InsertOne(context.TODO(), newFactory)
		if err != nil {
			return newStation, err
		}
	} else {
		factoryId = factory.ID
	}

	newStation = models.Station{
		ID:              primitive.NewObjectID(),
		Name:            stationName,
		FactoryId:       factoryId,
		RetentionType:   "message_age_sec",
		RetentionValue:  604800,
		StorageType:     "file",
		Replicas:        1,
		DedupEnabled:    false,
		DedupWindowInMs: 0,
		CreatedByUser:   username,
		CreationDate:    time.Now(),
		LastUpdate:      time.Now(),
		Functions:       []models.Function{},
	}

	err = broker.CreateStream(newStation)
	if err != nil {
		return newStation, err
	}

	_, err = stationsCollection.InsertOne(context.TODO(), newStation)
	if err != nil {
		return newStation, err
	}

	return newStation, nil
}

func shouldSendAnalytics() (bool, error) {
	filter := bson.M{"key": "analytics"}
	var systemKey models.SystemKey
	err := systemKeysCollection.FindOne(context.TODO(), filter).Decode(&systemKey)
	if err != nil {
		return false, err
	}

	if systemKey.Value == "true" {
		return true, nil
	} else {
		return false, nil
	}
}
