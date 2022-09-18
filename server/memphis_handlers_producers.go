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
	"encoding/json"
	"errors"
	"memphis-broker/analytics"
	"memphis-broker/models"
	"memphis-broker/utils"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/utils/strings/slices"
)

type ProducersHandler struct{ S *Server }

const (
	producerObjectName = "Producer"
)

func validateProducerName(name string) error {
	return validateName(name, producerObjectName)
}

func validateProducerType(producerType string) error {
	if producerType != "application" && producerType != "connector" {
		return errors.New("Producer type has to be one of the following application/connector")
	}
	return nil
}

func (s *Server) createProducerDirect(c *client, reply string, msg []byte) {
	var cpr createProducerRequest
	if err := json.Unmarshal(msg, &cpr); err != nil {
		s.Warnf("failed creating producer: %v", err.Error())
		respondWithErr(s, reply, err)
		return
	}
	name := strings.ToLower(cpr.Name)
	err := validateProducerName(name)
	if err != nil {
		serv.Warnf(err.Error())
		respondWithErr(s, reply, err)
		return
	}

	producerType := strings.ToLower(cpr.ProducerType)
	err = validateProducerType(producerType)
	if err != nil {
		serv.Warnf(err.Error())
		respondWithErr(s, reply, err)
		return
	}

	connectionIdObj, err := primitive.ObjectIDFromHex(cpr.ConnectionId)
	if err != nil {
		serv.Warnf("Connection id is not valid")
		respondWithErr(s, reply, err)
		return
	}
	exist, connection, err := IsConnectionExist(connectionIdObj)
	if err != nil {
		serv.Errorf("CreateProducer error: " + err.Error())
		respondWithErr(s, reply, err)
		return
	}
	if !exist {
		serv.Warnf("Connection id was not found")
		respondWithErr(s, reply, errors.New("memphis: connection id was not found"))
		return
	}
	if !connection.IsActive {
		serv.Warnf("Connection is not active")
		respondWithErr(s, reply, errors.New("memphis: connection id is not active"))
		return
	}

	stationName := strings.ToLower(cpr.StationName)
	exist, station, err := IsStationExist(stationName)
	if err != nil {
		serv.Errorf("CreateProducer error: " + err.Error())
		respondWithErr(s, reply, err)
		return
	}
	if !exist {
		var created bool
		station, created, err = CreateDefaultStation(s, stationName, connection.CreatedByUser)
		if err != nil {
			serv.Errorf("creating default station error: " + err.Error())
			respondWithErr(s, reply, err)
			return
		}

		if created {
			message := "Station " + stationName + " has been created"
			serv.Noticef(message)
			var auditLogs []interface{}
			newAuditLog := models.AuditLog{
				ID:            primitive.NewObjectID(),
				StationName:   stationName,
				Message:       message,
				CreatedByUser: c.memphisInfo.username,
				CreationDate:  time.Now(),
				UserType:      "application",
			}
			auditLogs = append(auditLogs, newAuditLog)
			err = CreateAuditLogs(auditLogs)
			if err != nil {
				serv.Errorf("CreateProducer error: " + err.Error())
			}

			shouldSendAnalytics, _ := shouldSendAnalytics()
			if shouldSendAnalytics {
				analytics.SendEvent(c.memphisInfo.username, "user-create-station")
			}
		}
	}

	exist, _, err = IsProducerExist(name, station.ID)
	if err != nil {
		serv.Errorf("CreateProducer error: " + err.Error())
		respondWithErr(s, reply, err)
		return
	}
	if exist {
		serv.Warnf("Producer name has to be unique per station")
		respondWithErr(s, reply, errors.New("memphis: producer name has to be unique per station"))
		return
	}

	newProducer := models.Producer{
		ID:            primitive.NewObjectID(),
		Name:          name,
		StationId:     station.ID,
		Type:          producerType,
		ConnectionId:  connectionIdObj,
		CreatedByUser: connection.CreatedByUser,
		IsActive:      true,
		CreationDate:  time.Now(),
		IsDeleted:     false,
	}

	filter := bson.M{"name": newProducer.Name, "station_id": station.ID, "is_active": true, "is_deleted": false}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":             newProducer.ID,
			"type":            newProducer.Type,
			"connection_id":   newProducer.ConnectionId,
			"created_by_user": newProducer.CreatedByUser,
			"creation_date":   newProducer.CreationDate,
		},
	}
	opts := options.Update().SetUpsert(true)
	updateResults, err := producersCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		serv.Errorf("CreateProducer error: " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	if updateResults.MatchedCount > 0 {
		message := "Producer " + name + " has been created"
		serv.Noticef(message)
		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			ID:            primitive.NewObjectID(),
			StationName:   stationName,
			Message:       message,
			CreatedByUser: c.memphisInfo.username,
			CreationDate:  time.Now(),
			UserType:      "application",
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			serv.Errorf("CreateProducer error: " + err.Error())
		}

		shouldSendAnalytics, _ := shouldSendAnalytics()
		if shouldSendAnalytics {
			analytics.SendEvent(c.memphisInfo.username, "user-create-producer")
		}
	}

	respondWithErr(s, reply, nil)
	return
}

func (ph ProducersHandler) GetAllProducers(c *gin.Context) {
	var producers []models.ExtendedProducer
	cursor, err := producersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	})

	if err != nil {
		serv.Errorf("GetAllProducers error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &producers); err != nil {
		serv.Errorf("GetAllProducers error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(producers) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, producers)
	}
}

func (ph ProducersHandler) GetProducersByStation(station models.Station) ([]models.ExtendedProducer, []models.ExtendedProducer, []models.ExtendedProducer, error) { // for socket io endpoint
	var producers []models.ExtendedProducer

	cursor, err := producersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"station_id", station.ID}}}},
		bson.D{{"$sort", bson.D{{"creation_date", -1}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	})

	if err != nil {
		return producers, producers, producers, err
	}

	if err = cursor.All(context.TODO(), &producers); err != nil {
		return producers, producers, producers, err
	}

	var connectedProducers []models.ExtendedProducer
	var disconnectedProducers []models.ExtendedProducer
	var deletedProducers []models.ExtendedProducer
	producersNames := []string{}

	for _, producer := range producers {
		if slices.Contains(producersNames, producer.Name) {
			continue
		}

		producersNames = append(producersNames, producer.Name)
		if producer.IsActive {
			connectedProducers = append(connectedProducers, producer)
		} else if !producer.IsDeleted && !producer.IsActive {
			disconnectedProducers = append(disconnectedProducers, producer)
		} else if producer.IsDeleted {
			deletedProducers = append(deletedProducers, producer)
		}
	}

	if len(connectedProducers) == 0 {
		connectedProducers = []models.ExtendedProducer{}
	}

	if len(disconnectedProducers) == 0 {
		disconnectedProducers = []models.ExtendedProducer{}
	}

	if len(deletedProducers) == 0 {
		deletedProducers = []models.ExtendedProducer{}
	}

	sort.Slice(connectedProducers, func(i, j int) bool {
		return connectedProducers[j].CreationDate.Before(connectedProducers[i].CreationDate)
	})
	sort.Slice(disconnectedProducers, func(i, j int) bool {
		return disconnectedProducers[j].CreationDate.Before(disconnectedProducers[i].CreationDate)
	})
	sort.Slice(deletedProducers, func(i, j int) bool {
		return deletedProducers[j].CreationDate.Before(deletedProducers[i].CreationDate)
	})
	return connectedProducers, disconnectedProducers, deletedProducers, nil
}

func (ph ProducersHandler) GetAllProducersByStation(c *gin.Context) { // for the REST endpoint
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
		serv.Warnf("Station does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station does not exist"})
		return
	}

	var producers []models.ExtendedProducer
	cursor, err := producersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"station_id", station.ID}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	})

	if err != nil {
		serv.Errorf("GetAllProducersByStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &producers); err != nil {
		serv.Errorf("GetAllProducersByStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(producers) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, producers)
	}
}

func (s *Server) destroyProducerDirect(c *client, reply string, msg []byte) {
	var dpr destroyProducerRequest
	if err := json.Unmarshal(msg, &dpr); err != nil {
		s.Warnf("failed destoying producer: %v", err.Error())
		respondWithErr(s, reply, err)
		return
	}
	stationName := strings.ToLower(dpr.StationName)
	name := strings.ToLower(dpr.ProducerName)
	_, station, err := IsStationExist(stationName)
	if err != nil {
		serv.Errorf("DestroyProducer error: " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	var producer models.Producer
	err = producersCollection.FindOneAndUpdate(context.TODO(),
		bson.M{"name": name, "station_id": station.ID, "is_active": true},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	).Decode(&producer)

	if err == mongo.ErrNoDocuments {
		serv.Warnf("Producer does not exist")
		respondWithErr(s, reply, errors.New("Producer does not exist"))
		return
	}
	if err != nil {
		serv.Errorf("DestroyProducer error: " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	message := "Producer " + name + " has been deleted"
	serv.Noticef(message)
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		ID:            primitive.NewObjectID(),
		StationName:   stationName,
		Message:       message,
		CreatedByUser: c.memphisInfo.username,
		CreationDate:  time.Now(),
		UserType:      "application",
	}
	auditLogs = append(auditLogs, newAuditLog)
	err = CreateAuditLogs(auditLogs)
	if err != nil {
		serv.Errorf("DestroyProducer error: " + err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(c.memphisInfo.username, "user-remove-producer")
	}

	respondWithErr(s, reply, nil)
	return
}

func (ph ProducersHandler) KillProducers(connectionId primitive.ObjectID) error {
	var producers []models.Producer
	var station models.Station

	cursor, err := producersCollection.Find(context.TODO(), bson.M{"connection_id": connectionId, "is_active": true})
	if err != nil {
		serv.Errorf("KillProducers error: " + err.Error())
	}
	if err = cursor.All(context.TODO(), &producers); err != nil {
		serv.Errorf("KillProducers error: " + err.Error())
	}

	if len(producers) > 0 {
		err = stationsCollection.FindOne(context.TODO(), bson.M{"_id": producers[0].StationId}).Decode(&station)
		if err != nil {
			serv.Errorf("KillProducers error: " + err.Error())
		}

		_, err = producersCollection.UpdateMany(context.TODO(),
			bson.M{"connection_id": connectionId},
			bson.M{"$set": bson.M{"is_active": false}},
		)
		if err != nil {
			serv.Errorf("KillProducers error: " + err.Error())
			return err
		}

		userType := "application"
		if producers[0].CreatedByUser == "root" {
			userType = "root"
		}

		var message string
		var auditLogs []interface{}
		var newAuditLog models.AuditLog
		for _, producer := range producers {
			message = "Producer " + producer.Name + " has been disconnected"
			newAuditLog = models.AuditLog{
				ID:            primitive.NewObjectID(),
				StationName:   station.Name,
				Message:       message,
				CreatedByUser: producers[0].CreatedByUser,
				CreationDate:  time.Now(),
				UserType:      userType,
			}
			auditLogs = append(auditLogs, newAuditLog)
		}
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			serv.Errorf("KillProducers error: " + err.Error())
		}
	}

	return nil
}

func (ph ProducersHandler) ReliveProducers(connectionId primitive.ObjectID) error {
	_, err := producersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": connectionId, "is_deleted": false},
		bson.M{"$set": bson.M{"is_active": true}},
	)
	if err != nil {
		serv.Errorf("ReliveProducers error: " + err.Error())
		return err
	}

	return nil
}
