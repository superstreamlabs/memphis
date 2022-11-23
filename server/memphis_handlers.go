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
	"context"
	"errors"
	"fmt"
	"memphis-broker/conf"
	"memphis-broker/db"
	"memphis-broker/models"
	"regexp"
	"strings"
	"sync"
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
	Tags       TagsHandler
	Schemas    SchemasHandler
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
var tagsCollection *mongo.Collection
var schemasCollection *mongo.Collection
var schemaVersionCollection *mongo.Collection
var sandboxUsersCollection *mongo.Collection
var serv *Server
var configuration = conf.GetConfig()

type srvMemphis struct {
	serverID               string
	nuid                   *nuid.NUID
	dbClient               *mongo.Client
	dbCtx                  context.Context
	dbCancel               context.CancelFunc
	activateSysLogsPubFunc func()
	fallbackLogQ           *ipQueue
	mcrReported            bool
	mcr                    chan struct{} // memphis cluster ready
	jsApiMu                sync.Mutex
	ws                     memphisWS
}

type memphisWS struct {
	subscriptions map[string]memphisWSSubscription
	webSocketMu   sync.Mutex
	quitCh        chan struct{}
}

func (s *Server) InitializeMemphisHandlers(dbInstance db.DbInstance) {
	serv = s
	s.memphis.dbClient = dbInstance.Client
	s.memphis.dbCtx = dbInstance.Ctx
	s.memphis.dbCancel = dbInstance.Cancel
	s.memphis.nuid = nuid.New()
	s.memphis.serverID = configuration.SERVER_NAME
	s.memphis.mcrReported = false
	s.memphis.mcr = make(chan struct{})

	usersCollection = db.GetCollection("users", dbInstance.Client)
	imagesCollection = db.GetCollection("images", dbInstance.Client)
	stationsCollection = db.GetCollection("stations", dbInstance.Client)
	connectionsCollection = db.GetCollection("connections", dbInstance.Client)
	producersCollection = db.GetCollection("producers", dbInstance.Client)
	consumersCollection = db.GetCollection("consumers", dbInstance.Client)
	systemKeysCollection = db.GetCollection("system_keys", dbInstance.Client)
	auditLogsCollection = db.GetCollection("audit_logs", dbInstance.Client)
	poisonMessagesCollection = db.GetCollection("poison_messages", dbInstance.Client)
	tagsCollection = db.GetCollection("tags", dbInstance.Client)
	schemasCollection = db.GetCollection("schemas", dbInstance.Client)
	schemaVersionCollection = db.GetCollection("schema_versions", dbInstance.Client)
	sandboxUsersCollection = db.GetCollection("sandbox_users", serv.memphis.dbClient)

	poisonMessagesCollection.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{"creation_date": -1}, Options: nil,
	})

	s.initializeSDKHandlers()
	s.initWS()
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

func IsStationExist(sn StationName) (bool, models.Station, error) {
	stationName := sn.Ext()
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

func IsTagExist(tagName string) (bool, models.Tag, error) {
	filter := bson.M{
		"name": tagName,
	}
	var tag models.Tag
	err := tagsCollection.FindOne(context.TODO(), filter).Decode(&tag)
	if err == mongo.ErrNoDocuments {
		return false, tag, nil
	} else if err != nil {
		return false, tag, err
	}
	return true, tag, nil
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

func CreateDefaultStation(s *Server, sn StationName, username string) (models.Station, bool, error) {
	var newStation models.Station
	stationName := sn.Ext()
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

	err := s.CreateStream(sn, newStation)
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

func validateName(name, objectType string) error {
	emptyErrStr := fmt.Sprintf("%v name can not be empty", objectType)
	tooLongErrStr := fmt.Sprintf("%v should be under 32 characters", objectType)
	invalidCharErrStr := fmt.Sprintf("Only alphanumeric and the '_', '-', '.' characters are allowed in %v", objectType)
	firstLetterErrStr := fmt.Sprintf("%v name can not start or end with non alphanumeric character", objectType)

	emptyErr := errors.New(emptyErrStr)
	tooLongErr := errors.New(tooLongErrStr)
	invalidCharErr := errors.New(invalidCharErrStr)
	firstLetterErr := errors.New(firstLetterErrStr)

	if len(name) == 0 {
		return emptyErr
	}

	if len(name) > 32 {
		return tooLongErr
	}

	re := regexp.MustCompile("^[a-z0-9_.-]*$")

	validName := re.MatchString(name)
	if !validName {
		return invalidCharErr
	}

	if name[0:1] == "." || name[0:1] == "-" || name[0:1] == "_" || name[len(name)-1:] == "." || name[len(name)-1:] == "-" || name[len(name)-1:] == "_" {
		return firstLetterErr
	}

	return nil
}

const (
	delimiterToReplace   = "."
	delimiterReplacement = "#"
)

func replaceDelimiters(name string) string {
	return strings.Replace(name, delimiterToReplace, delimiterReplacement, -1)
}

func revertDelimiters(name string) string {
	return strings.Replace(name, delimiterReplacement, delimiterToReplace, -1)
}

func IsSchemaExist(schemaName string) (bool, models.Schema, error) {
	filter := bson.M{
		"name": schemaName}
	var schema models.Schema
	err := schemasCollection.FindOne(context.TODO(), filter).Decode(&schema)
	if err == mongo.ErrNoDocuments {
		return false, schema, nil
	} else if err != nil {
		return false, schema, err
	}
	return true, schema, nil
}

func isSchemaVersionExists(version int, schemaId primitive.ObjectID) (bool, models.SchemaVersion, error) {
	var schemaVersion models.SchemaVersion
	filter := bson.M{"schema_id": schemaId, "version_number": version}
	err := schemaVersionCollection.FindOne(context.TODO(), filter).Decode(&schemaVersion)

	if err == mongo.ErrNoDocuments {
		return false, schemaVersion, nil
	} else if err != nil {
		return false, schemaVersion, err
	}
	return true, schemaVersion, nil
}
