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
	"context"
	"errors"
	"memphis-broker/conf"
	"memphis-broker/db"
	"memphis-broker/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handlers struct {
	Producers  ProducersHandler
	Consumers  ConsumersHandler
	AuditLogs  AuditLogsHandler
	Stations   StationsHandler
	Monitoring MonitoringHandler
	PoisonMsgs PoisonMessagesHandler
}

var usersCollection *mongo.Collection
var imagesCollection *mongo.Collection
var stationsCollection *mongo.Collection
var connectionsCollection *mongo.Collection
var producersCollection *mongo.Collection
var consumersCollection *mongo.Collection
var systemKeysCollection *mongo.Collection
var auditLogsCollection *mongo.Collection
var poisonMessagesCollection *mongo.Collection
var serv *Server
var configuration = conf.GetConfig()

type srvMemphis struct {
	serverID               string
	nuid                   *nuid.NUID
	dbClient               *mongo.Client
	dbCtx                  context.Context
	dbCancel               context.CancelFunc
	activateSysLogsPubFunc func()
}

func (s *Server) InitializeMemphisHandlers(dbInstance db.DbInstance) {
	serv = s
	s.memphis.dbClient = dbInstance.Client
	s.memphis.dbCtx = dbInstance.Ctx
	s.memphis.dbCancel = dbInstance.Cancel
	s.memphis.nuid = nuid.New()
	s.memphis.serverID = configuration.SERVER_NAME

	usersCollection = db.GetCollection("users", dbInstance.Client)
	imagesCollection = db.GetCollection("images", dbInstance.Client)
	stationsCollection = db.GetCollection("stations", dbInstance.Client)
	connectionsCollection = db.GetCollection("connections", dbInstance.Client)
	producersCollection = db.GetCollection("producers", dbInstance.Client)
	consumersCollection = db.GetCollection("consumers", dbInstance.Client)
	systemKeysCollection = db.GetCollection("system_keys", dbInstance.Client)
	auditLogsCollection = db.GetCollection("audit_logs", dbInstance.Client)
	poisonMessagesCollection = db.GetCollection("poison_messages", dbInstance.Client)

	poisonMessagesCollection.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{"creation_date": -1}, Options: nil,
	})

	s.initializeSDKHandlers()
}

func getUserDetailsFromMiddleware(c *gin.Context) (models.User, error) {
	user, _ := c.Get("user")
	userModel := user.(models.User)
	if len(userModel.Username) == 0 {
		return userModel, errors.New("Username is empty")
	}
	return userModel, nil
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

func IsStationExist(stationName string) (bool, models.Station, error) {
	filter := bson.M{
		"name": stationName,
		"$or": []interface{}{
			bson.M{"is_deleted": false},
			bson.M{"is_deleted": bson.M{"$exists": false}},
		},
	}
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

func CreateDefaultStation(s *Server, stationName string, username string) (models.Station, bool, error) {
	var newStation models.Station
	newStation = models.Station{
		ID:              primitive.NewObjectID(),
		Name:            stationName,
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

	err := s.CreateStream(newStation)
	if err != nil {
		return newStation, false, err
	}

	filter := bson.M{"name": newStation.Name, "is_deleted": false}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":                newStation.ID,
			"retention_type":     newStation.RetentionType,
			"retention_value":    newStation.RetentionValue,
			"storage_type":       newStation.StorageType,
			"replicas":           newStation.Replicas,
			"dedup_enabled":      newStation.DedupEnabled,
			"dedup_window_in_ms": newStation.DedupWindowInMs,
			"created_by_user":    newStation.CreatedByUser,
			"creation_date":      newStation.CreationDate,
			"last_update":        newStation.LastUpdate,
			"functions":          newStation.Functions,
		},
	}
	opts := options.Update().SetUpsert(true)
	updateResults, err := stationsCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return newStation, false, err
	}
	if updateResults.MatchedCount > 0 {
		return newStation, false, nil
	}

	return newStation, true, nil
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
