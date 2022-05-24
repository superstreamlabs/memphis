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
	"memphis-control-plane/utils"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProducersHandler struct{}

type extendedProducer struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	Type          string             `json:"type" bson:"type"`
	ConnectionId  primitive.ObjectID `json:"connection_id" bson:"connection_id"`
	CreatedByUser string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
	StationName   string             `json:"station_name" bson:"station_name"`
	FactoryName   string             `json:"factory_name" bson:"factory_name"`
}

func validateProducerName(name string) error {
	re := regexp.MustCompile("^[a-z_]*$")

	validName := re.MatchString(name)
	if !validName {
		return errors.New("Producer name has to include only letters and _")
	}
	return nil
}

func validateProducerType(producerType string) error {
	if producerType != "application" && producerType != "connector" {
		return errors.New("Producer type has to be one of the following application/connector")
	}
	return nil
}

func (umh ProducersHandler) CreateProducer(c *gin.Context) {
	var body models.CreateProducerSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	name := strings.ToLower(body.Name)
	err := validateProducerName(name)
	if err != nil {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	producerType := strings.ToLower(body.ProducerType)
	err = validateProducerType(producerType)
	if err != nil {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	connectionId, err := primitive.ObjectIDFromHex(body.ConnectionId)
	if err != nil {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Connection id is not valid"})
		return
	}
	exist, connection, err := IsConnectionExist(connectionId)
	if err != nil {
		logger.Error("CreateProducer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Connection id was not found"})
		return
	}
	if !connection.IsActive {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Connection is not active"})
		return
	}

	stationName := strings.ToLower(body.StationName)
	exist, station, err := IsStationExist(stationName)
	if err != nil {
		logger.Error("CreateProducer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		station, err = CreateDefaultStation(stationName, connection.CreatedByUser)
		if err != nil {
			logger.Error("CreateProducer error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	exist, _, err = IsProducerExist(name, station.ID)
	if err != nil {
		logger.Error("CreateProducer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Producer name has to be unique in a station level"})
		return
	}

	producerId := primitive.NewObjectID()
	newProducer := models.Producer{
		ID:            producerId,
		Name:          name,
		StationId:     station.ID,
		FactoryId:     station.FactoryId,
		Type:          producerType,
		ConnectionId:  connectionId,
		CreatedByUser: connection.CreatedByUser,
		IsActive:      true,
		CreationDate:  time.Now(),
	}

	_, err = producersCollection.InsertOne(context.TODO(), newProducer)
	if err != nil {
		logger.Error("CreateProducer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	var user models.User
	err = usersCollection.FindOne(context.TODO(), bson.M{"username": connection.CreatedByUser}).Decode(&user)
	if err != nil {
		logger.Error("CreateProducer error: " + err.Error())
	}

	message := "Producer " + name + " has been created"
	logger.Info(message)
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		ID:              primitive.NewObjectID(),
		StationName:     stationName,
		Message:       	 message,
		CreatedByUser:   user.Username,
		CreationDate:    time.Now(),
		UserType: 		 user.UserType,
	}
	auditLogs = append(auditLogs, newAuditLog)
	CreateAuditLogs(auditLogs)
	c.IndentedJSON(200, gin.H{
		"producer_id": producerId,
	})
}

func (umh ProducersHandler) GetAllProducers(c *gin.Context) {
	var producers []extendedProducer
	cursor, err := producersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"is_active", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "factories"}, {"localField", "factory_id"}, {"foreignField", "_id"}, {"as", "factory"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$factory"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"station_name", "$station.name"}, {"factory_name", "$factory.name"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"factory", 0}}}},
	})

	if err != nil {
		logger.Error("GetAllProducers error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &producers); err != nil {
		logger.Error("GetAllProducers error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(producers) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, producers)
	}
}

func (umh ProducersHandler) GetAllProducersByStation(c *gin.Context) {
	var body models.GetAllProducersByStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	exist, station, err := IsStationExist(body.StationName)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station does not exist"})
		return
	}

	var producers []extendedProducer
	cursor, err := producersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"station_id", station.ID}, {"is_active", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "factories"}, {"localField", "factory_id"}, {"foreignField", "_id"}, {"as", "factory"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$factory"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"station_name", "$station.name"}, {"factory_name", "$factory.name"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"factory", 0}}}},
	})

	if err != nil {
		logger.Error("GetAllProducersByStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &producers); err != nil {
		logger.Error("GetAllProducersByStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(producers) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, producers)
	}
}

func (umh ProducersHandler) DestroyProducer(c *gin.Context) {
	var body models.DestroyProducerSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName := strings.ToLower(body.StationName)
	name := strings.ToLower(body.Name)
	_, station, err := IsStationExist(stationName)
	if err != nil {
		logger.Error("DestroyProducer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	var producer models.Producer
	err = producersCollection.FindOneAndUpdate(context.TODO(),
		bson.M{"name": name, "station_id": station.ID, "is_active": true},
		bson.M{"$set": bson.M{"is_active": false}},
	).Decode(&producer)
	if err != nil {
		logger.Error("DestroyProducer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if err == mongo.ErrNoDocuments {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "A producer with the given details was not found"})
		return
	}
	user := getUserDetailsFromMiddleware(c)
	message := "Producer " + name + " has been deleted"
	logger.Info(message)
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		ID:              primitive.NewObjectID(),
		StationName:     stationName,
		Message:       	 message,
		CreatedByUser:   user.Username,
		CreationDate:    time.Now(),
		UserType: 		 user.UserType,
	}
	auditLogs = append(auditLogs, newAuditLog)
	CreateAuditLogs(auditLogs)
	c.IndentedJSON(200, gin.H{})
}

func (umh ProducersHandler) GetAllProducersByConnection(connectionId primitive.ObjectID) ([]models.Producer, error) {
	var producers []models.Producer

	cursor, err := producersCollection.Find(context.TODO(), bson.M{"connection_id": connectionId})
	if err != nil {
		logger.Error("GetAllProducersByConnection error: " + err.Error())
		return producers, err
	}

	if err = cursor.All(context.TODO(), &producers); err != nil {
		logger.Error("GetAllProducersByConnection error: " + err.Error())
		return producers, err
	}

	return producers, nil
}

func (umh ProducersHandler) KillProducers(connectionId primitive.ObjectID) error {
	exist, connection, err := IsConnectionExist(connectionId)
	if err != nil {
		logger.Error("KillProducers error: " + err.Error())
		return err
	}
	if !exist {
		logger.Error("KillProducers error: connection does not exist")
		return errors.New("KillProducers error: connection does not exist")
	}
	if !connection.IsActive {
		logger.Error("KillProducers error: connection is not active")
		return errors.New("KillProducers error: connection is not active")
	}
	var producers []models.Producer
	cursor, err := producersCollection.Find(context.TODO(), bson.M{"connection_id": connectionId, "is_active": true})
	if err != nil {
		logger.Error("KillProducers error: " + err.Error())
		return err
	}

	if err = cursor.All(context.TODO(), &producers); err != nil {
		logger.Error("KillProducers error: " + err.Error())
		return err
	}
	_, err = producersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": connectionId},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		logger.Error("KillProducers error: " + err.Error())
		return err
	}
	var user models.User
	err = usersCollection.FindOne(context.TODO(), bson.M{"username": connection.CreatedByUser}).Decode(&user)
	if err != nil {
		logger.Error("KillProducers error: " + err.Error())
		return err
	}
	var station models.Station
	err = stationsCollection.FindOne(context.TODO(), bson.M{"_id": producers[0].StationId}).Decode(&station)
	if err != nil {
		logger.Error("KillProducers error: " + err.Error())
		return err
	}
	var message string
	var auditLogs []interface{}
	var newAuditLog models.AuditLog
	for _,producer := range producers{
		message = "Producer" + producer.Name + "disconnected"
		newAuditLog = models.AuditLog{
			ID:              primitive.NewObjectID(),
			StationName:     station.Name,
			Message:       	 message,
			CreatedByUser:   user.Username,
			CreationDate:    time.Now(),
			UserType: 		 user.UserType,
		}
		auditLogs = append(auditLogs, newAuditLog)
	}
	CreateAuditLogs(auditLogs)
	return nil
}

func (umh ProducersHandler) ReliveProducers(connectionId primitive.ObjectID) error {
	_, err := producersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": connectionId},
		bson.M{"$set": bson.M{"is_active": true}},
	)
	if err != nil {
		logger.Error("ReliveProducers error: " + err.Error())
		return err
	}

	return nil
}
