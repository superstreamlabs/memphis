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
	"errors"
	"memphis-control-plane/broker"
	"memphis-control-plane/logger"
	"memphis-control-plane/models"
	"memphis-control-plane/utils"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type StationsHandler struct{}

func validateStationName(stationName string) error {
	re := regexp.MustCompile("^[a-z0-9_]*$")

	validName := re.MatchString(stationName)
	if !validName {
		return errors.New("station name has to include only letters, numbers and _")
	}
	return nil
}

func validateRetentionType(retentionType string) error {
	if retentionType != "message_age_sec" && retentionType != "messages" && retentionType != "bytes" {
		return errors.New("retention type can be one of the following message_age_sec/messages/bytes")
	}

	return nil
}

func validateStorageType(storageType string) error {
	if storageType != "file" && storageType != "memory" {
		return errors.New("storage type can be one of the following file/memory")
	}

	return nil
}

func validateReplicas(replicas int) error {
	if replicas > 5 {
		return errors.New("max replicas in a cluster is 5")
	}

	return nil
}

// TODO remove the station resources - functions, connectors
func removeStationResources(station models.Station) error {
	err := broker.RemoveStream(station.Name)
	if err != nil {
		return err
	}

	_, err = producersCollection.DeleteMany(context.TODO(), bson.M{"station_id": station.ID})
	if err != nil {
		return err
	}

	_, err = consumersCollection.DeleteMany(context.TODO(), bson.M{"station_id": station.ID})
	if err != nil {
		return err
	}

	err = RemoveAllAuditLogsByStation(station.Name)
	if err != nil {
		logger.Warn("removeStationResources error: " + err.Error())
	}

	return nil
}

func (sh StationsHandler) GetStation(c *gin.Context) {
	var body models.GetStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	var station models.Station
	err := stationsCollection.FindOne(context.TODO(), bson.M{"name": body.StationName}).Decode(&station)
	if err == mongo.ErrNoDocuments {
		c.AbortWithStatusJSON(404, gin.H{"message": "Station does not exist"})
		return
	} else if err != nil {
		logger.Error("GetStationById error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, station)
}

func (sh StationsHandler) GetAllStations(c *gin.Context) {
	type extendedStation struct {
		ID              primitive.ObjectID `json:"id" bson:"_id"`
		Name            string             `json:"name" bson:"name"`
		FactoryId       primitive.ObjectID `json:"factory_id" bson:"factory_id"`
		RetentionType   string             `json:"retention_type" bson:"retention_type"`
		RetentionValue  int                `json:"retention_value" bson:"retention_value"`
		StorageType     string             `json:"storage_type" bson:"storage_type"`
		Replicas        int                `json:"replicas" bson:"replicas"`
		DedupEnabled    bool               `json:"dedup_enabled" bson:"dedup_enabled"`
		DedupWindowInMs int                `json:"dedup_window_in_ms" bson:"dedup_window_in_ms"`
		CreatedByUser   string             `json:"created_by_user" bson:"created_by_user"`
		CreationDate    time.Time          `json:"creation_date" bson:"creation_date"`
		LastUpdate      time.Time          `json:"last_update" bson:"last_update"`
		Functions       []models.Function  `json:"functions" bson:"functions"`
		FactoryName     string             `json:"factory_name" bson:"factory_name"`
	}

	var stations []extendedStation
	cursor, err := stationsCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$lookup", bson.D{{"from", "factories"}, {"localField", "factory_id"}, {"foreignField", "_id"}, {"as", "factory"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$factory"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"factory_id", 1}, {"retention_type", 1}, {"retention_value", 1}, {"storage_type", 1}, {"replicas", 1}, {"dedup_enabled", 1}, {"dedup_window_in_ms", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"last_update", 1}, {"functions", 1}, {"factory_name", "$factory.name"}}}},
		bson.D{{"$project", bson.D{{"factory", 0}}}},
	})

	if err != nil {
		logger.Error("GetAllStations error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &stations); err != nil {
		logger.Error("GetAllStations error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(stations) == 0 {
		c.IndentedJSON(200, []models.Station{})
	} else {
		c.IndentedJSON(200, stations)
	}
}

func (sh StationsHandler) CreateStation(c *gin.Context) {
	var body models.CreateStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName := strings.ToLower(body.Name)
	err := validateStationName(stationName)
	if err != nil {
		logger.Warn(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, _, err := IsStationExist(stationName)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		logger.Warn("Station with the same name is already exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station with the same name is already exist"})
		return
	}

	factortyName := strings.ToLower(body.FactoryName)
	exist, factory, err := IsFactoryExist(factortyName)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		logger.Warn("Factory name does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Factory name does not exist"})
		return
	}

	var retentionType string
	if body.RetentionType != "" && body.RetentionValue > 0 {
		retentionType = strings.ToLower(body.RetentionType)
		err = validateRetentionType(retentionType)
		if err != nil {
			logger.Warn(err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	} else {
		retentionType = "message_age_sec"
		body.RetentionValue = 604800 // 1 week
	}

	var storageType string
	if body.StorageType != "" {
		storageType = strings.ToLower(body.StorageType)
		err = validateStorageType(storageType)
		if err != nil {
			logger.Warn(err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	} else {
		body.StorageType = "file"
	}

	if body.Replicas > 0 {
		err = validateReplicas(body.Replicas)
		if err != nil {
			logger.Warn(err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	} else {
		body.Replicas = 1
	}

	user := getUserDetailsFromMiddleware(c)
	newStation := models.Station{
		ID:              primitive.NewObjectID(),
		Name:            stationName,
		FactoryId:       factory.ID,
		RetentionType:   retentionType,
		RetentionValue:  body.RetentionValue,
		StorageType:     storageType,
		Replicas:        body.Replicas,
		DedupEnabled:    body.DedupEnabled,
		DedupWindowInMs: body.DedupWindowInMs,
		CreatedByUser:   user.Username,
		CreationDate:    time.Now(),
		LastUpdate:      time.Now(),
		Functions:       []models.Function{},
	}

	err = broker.CreateStream(newStation)
	if err != nil {
		logger.Warn(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	_, err = stationsCollection.InsertOne(context.TODO(), newStation)
	if err != nil {
		logger.Error("CreateStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	message := "Station " + stationName + " has been created"
	logger.Info(message)
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		ID:            primitive.NewObjectID(),
		StationName:   stationName,
		Message:       message,
		CreatedByUser: user.Username,
		CreationDate:  time.Now(),
		UserType:      user.UserType,
	}
	auditLogs = append(auditLogs, newAuditLog)
	err = CreateAuditLogs(auditLogs)
	if err != nil {
		logger.Warn("CreateStation error: " + err.Error())
	}
	c.IndentedJSON(200, newStation)
}

func (sh StationsHandler) RemoveStation(c *gin.Context) {
	var body models.RemoveStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName := strings.ToLower(body.StationName)
	exist, station, err := IsStationExist(stationName)
	if err != nil {
		logger.Error("RemoveStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		logger.Warn("Station does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station does not exist"})
		return
	}

	err = removeStationResources(station)
	if err != nil {
		logger.Error("RemoveStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	_, err = stationsCollection.DeleteOne(context.TODO(), bson.M{"name": stationName})
	if err != nil {
		logger.Error("RemoveStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	logger.Info("Station " + stationName + " has been deleted")
	c.IndentedJSON(200, gin.H{})
}

func (sh StationsHandler) GetTotalMessages(station models.Station) (int, error) {
	totalMessages, err := broker.GetTotalMessagesInStation(station)
	return totalMessages, err
}

func (sh StationsHandler) GetAvgMsgSize(station models.Station) (int64, error) {
	avgMsgSize, err := broker.GetAvgMsgSizeInStation(station)
	return avgMsgSize, err
}
