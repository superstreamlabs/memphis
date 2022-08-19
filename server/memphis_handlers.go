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
	"context"
	"memphis-broker/conf"
	"memphis-broker/db"
	"memphis-broker/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handlers struct {
	Producers  ProducersHandler
	Consumers  ConsumersHandler
	AuditLogs  AuditLogsHandler
	Stations   StationsHandler
	Factories  FactoriesHandler
	Monitoring MonitoringHandler
	PoisonMsgs PoisonMessagesHandler
}

var usersCollection *mongo.Collection
var imagesCollection *mongo.Collection
var factoriesCollection *mongo.Collection
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
	replySubjectActive map[string]bool
	replySubjectRespCh map[string]chan []byte
	dbClient           *mongo.Client
	dbCtx              context.Context
	dbCancel           context.CancelFunc
}

func (s *Server) InitializeMemphisHandlers(dbInstance db.DbInstance) {

	// initialise db
	serv = s
	serv.memphis.dbClient = dbInstance.Client
	serv.memphis.dbCtx = dbInstance.Ctx
	serv.memphis.dbCancel = dbInstance.Cancel
	serv.memphis.replySubjectActive = make(map[string]bool)
	serv.memphis.replySubjectRespCh = make(map[string]chan []byte)

	usersCollection = db.GetCollection("users", dbInstance.Client)
	imagesCollection = db.GetCollection("images", dbInstance.Client)
	factoriesCollection = db.GetCollection("factories", dbInstance.Client)
	stationsCollection = db.GetCollection("stations", dbInstance.Client)
	connectionsCollection = db.GetCollection("connections", dbInstance.Client)
	producersCollection = db.GetCollection("producers", dbInstance.Client)
	consumersCollection = db.GetCollection("consumers", dbInstance.Client)
	systemKeysCollection = db.GetCollection("system_keys", dbInstance.Client)
	auditLogsCollection = db.GetCollection("audit_logs", dbInstance.Client)
	poisonMessagesCollection = db.GetCollection("poison_messages", dbInstance.Client)
	s.initialiseSDKHandlers()
	listenForPoisonMessages(s)
	s.prepReplySubjSubscription(replySubjectStreamInfo)
}

func (s *Server) prepReplySubjSubscription(subject string) {
	s.memphis.replySubjectActive[subject] = false
	s.memphis.replySubjectRespCh[subject] = make(chan []byte)
	s.subscribeOnGlobalAcc(subject, subject+"_sid", createReplyHandler(s))
}

func createReplyHandler(s *Server) simplifiedMsgHandler {
	return func(subject, reply string, msg []byte) {
		active, ok := s.memphis.replySubjectActive[subject]
		if !ok {
			s.Fatalf("memphis reply subject does not exist")
		}

		if active {
			s.memphis.replySubjectRespCh[subject] <- msg
		}
	}
}

var poisonMessagesHandler PoisonMessagesHandler

func listenForPoisonMessages(s *Server) {
	s.queueSubscribe("$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.>",
		"$memphis_poison_messages_listeners_group",
		poisonMessagesHandler.HandleNewMessage)
}

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
	filter := bson.M{
		"name": factoryName,
		"$or": []interface{}{
			bson.M{"is_deleted": false},
			bson.M{"is_deleted": bson.M{"$exists": false}},
		},
	}
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

func CreateDefaultStation(s *Server, stationName string, username string) (models.Station, error) {
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

	err = s.CreateStream(newStation)
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
