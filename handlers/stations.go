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
	"memphis-broker/analytics"
	"memphis-broker/broker"
	"memphis-broker/logger"
	"memphis-broker/models"
	"memphis-broker/utils"
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
	if len(stationName) > 32 {
		return errors.New("station name should be under 32 characters")
	}

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

	_, err = producersCollection.UpdateMany(context.TODO(),
		bson.M{"station_id": station.ID},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	)
	if err != nil {
		return err
	}

	_, err = consumersCollection.UpdateMany(context.TODO(),
		bson.M{"station_id": station.ID},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	)
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
	err := stationsCollection.FindOne(context.TODO(), bson.M{
		"name": body.StationName,
		"$or": []interface{}{
			bson.M{"is_deleted": false},
			bson.M{"is_deleted": bson.M{"$exists": false}},
		},
	}).Decode(&station)
	if err == mongo.ErrNoDocuments {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station does not exist"})
		return
	} else if err != nil {
		logger.Error("GetStationById error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, station)
}

func (sh StationsHandler) GetAllStationsDetails() ([]models.ExtendedStation, error) {
	var stations []models.ExtendedStation
	cursor, err := stationsCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"$or", []interface{}{
			bson.D{{"is_deleted", false}},
			bson.D{{"is_deleted", bson.D{{"$exists", false}}}},
		}}}}},
		bson.D{{"$lookup", bson.D{{"from", "factories"}, {"localField", "factory_id"}, {"foreignField", "_id"}, {"as", "factory"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$factory"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"factory_id", 1}, {"retention_type", 1}, {"retention_value", 1}, {"storage_type", 1}, {"replicas", 1}, {"dedup_enabled", 1}, {"dedup_window_in_ms", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"last_update", 1}, {"functions", 1}, {"factory_name", "$factory.name"}}}},
		bson.D{{"$project", bson.D{{"factory", 0}}}},
	})

	if err != nil {
		return stations, err
	}

	if err = cursor.All(context.TODO(), &stations); err != nil {
		return stations, err
	}

	if len(stations) == 0 {
		return []models.ExtendedStation{}, nil
	} else {
		return stations, nil
	}
}

func (sh StationsHandler) GetAllStations(c *gin.Context) {
	stations, err := sh.GetAllStationsDetails()
	if err != nil {
		logger.Error("GetAllStations error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, stations)
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

	user := getUserDetailsFromMiddleware(c)
	factoryName := strings.ToLower(body.FactoryName)
	exist, factory, err := IsFactoryExist(factoryName)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist { // create this factory
		err := validateFactoryName(factoryName)
		if err != nil {
			logger.Warn(err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}

		factory = models.Factory{
			ID:            primitive.NewObjectID(),
			Name:          factoryName,
			Description:   "",
			CreatedByUser: user.Username,
			CreationDate:  time.Now(),
			IsDeleted:     false,
		}
		_, err = factoriesCollection.InsertOne(context.TODO(), factory)
		if err != nil {
			logger.Error("CreateStation error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
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
		IsDeleted:       false,
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

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.IncrementStationsCounter()
	}

	c.IndentedJSON(200, newStation)
}

func (sh StationsHandler) RemoveStation(c *gin.Context) {
	err := DenyForSandboxEnv()
	if err != nil {
		logger.Error("RemoveStation error: " + err.Error())
		c.AbortWithStatusJSON(666, gin.H{"message": err.Error()})
		return
	}
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

	_, err = stationsCollection.UpdateOne(context.TODO(),
		bson.M{
			"name": stationName,
			"$or": []interface{}{
				bson.M{"is_deleted": false},
				bson.M{"is_deleted": bson.M{"$exists": false}},
			},
		},
		bson.M{"$set": bson.M{"is_deleted": true}},
	)
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

func (sh StationsHandler) GetTotalMessagesAcrossAllStations() (int, error) {
	totalMessages, err := broker.GetTotalMessagesAcrossAllStations()
	return totalMessages, err
}

func (sh StationsHandler) GetAvgMsgSize(station models.Station) (int64, error) {
	avgMsgSize, err := broker.GetAvgMsgSizeInStation(station)
	return avgMsgSize, err
}

func (sh StationsHandler) GetMessages(station models.Station, messagesToFetch int) ([]models.Message, error) {
	messages, err := broker.GetMessages(station, messagesToFetch)
	if err != nil {
		return []models.Message{}, err
	}

	if len(messages) == 0 {
		return []models.Message{}, nil
	} else {
		return messages, nil
	}
}
