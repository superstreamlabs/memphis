// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
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

func (s *Server) createProducerDirectCommon(c *client, pName, pType, pConnectionId string, pStationName StationName) (bool, bool, error) {
	name := strings.ToLower(pName)
	err := validateProducerName(name)
	if err != nil {
		serv.Warnf("createProducerDirectCommon: Producer " + pName + " at station " + pStationName.external + ": " + err.Error())
		return false, false, err
	}

	producerType := strings.ToLower(pType)
	err = validateProducerType(producerType)
	if err != nil {
		serv.Warnf("createProducerDirectCommon: Producer " + pName + " at station " + pStationName.external + ": " + err.Error())
		return false, false, err
	}

	connectionIdObj, err := primitive.ObjectIDFromHex(pConnectionId)
	if err != nil {
		serv.Warnf("createProducerDirectCommon: Producer " + pName + " at station " + pStationName.external + ": Connection ID " + pConnectionId + " is not valid")
		return false, false, err
	}
	exist, connection, err := IsConnectionExist(connectionIdObj)
	if err != nil {
		serv.Errorf("createProducerDirectCommon: Producer " + pName + " at station " + pStationName.external + ": " + err.Error())
		return false, false, err
	}
	if !exist {
		errMsg := "Connection ID " + pConnectionId + " was not found"
		serv.Warnf("createProducerDirectCommon: Producer " + pName + " at station " + pStationName.external + ": " + errMsg)
		return false, false, errors.New("memphis: " + errMsg)
	}
	if !connection.IsActive {
		errMsg := "Connection with ID " + pConnectionId + " is not active"
		serv.Warnf("createProducerDirectCommon: Producer " + pName + " at station " + pStationName.external + ": " + errMsg)
		return false, false, errors.New("memphis: " + errMsg)
	}

	exist, station, err := IsStationExist(pStationName)
	if err != nil {
		serv.Errorf("createProducerDirectCommon: Producer " + pName + " at station " + pStationName.external + ": " + err.Error())
		return false, false, err
	}
	if !exist {
		var created bool
		station, created, err = CreateDefaultStation(s, pStationName, connection.CreatedByUser)
		if err != nil {
			serv.Errorf("createProducerDirectCommon: creating default station error - producer " + pName + " at station " + pStationName.external + ": " + err.Error())
			return false, false, err
		}

		if created {
			message := "Station " + pStationName.Ext() + " has been created by user " + connection.CreatedByUser
			serv.Noticef(message)
			var auditLogs []interface{}
			newAuditLog := models.AuditLog{
				ID:            primitive.NewObjectID(),
				StationName:   pStationName.Ext(),
				Message:       message,
				CreatedByUser: connection.CreatedByUser,
				CreationDate:  time.Now(),
				UserType:      "application",
			}
			auditLogs = append(auditLogs, newAuditLog)
			err = CreateAuditLogs(auditLogs)
			if err != nil {
				serv.Errorf("createProducerDirectCommon: Producer " + pName + " at station " + pStationName.external + ": " + err.Error())
			}

			shouldSendAnalytics, _ := shouldSendAnalytics()
			if shouldSendAnalytics {
				param := analytics.EventParam{
					Name:  "station-name",
					Value: pStationName.Ext(),
				}
				analyticsParams := []analytics.EventParam{param}
				analytics.SendEventWithParams(connection.CreatedByUser, analyticsParams, "user-create-station-sdk")
			}
		}
	}

	exist, _, err = IsProducerExist(name, station.ID)
	if err != nil {
		serv.Errorf("createProducerDirectCommon: Producer " + pName + " at station " + pStationName.external + ": " + err.Error())
		return false, false, err
	}
	if exist {
		errMsg := "Producer name (" + pName + ") has to be unique per station (" + pStationName.external + ")"
		serv.Warnf("createProducerDirectCommon: " + errMsg)
		return false, false, errors.New("memphis: " + errMsg)
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
		serv.Errorf("createProducerDirectCommon: Producer " + pName + " at station " + pStationName.external + ": " + err.Error())
		return false, false, err
	}

	if updateResults.MatchedCount == 0 {
		message := "Producer " + name + " has been created by user " + connection.CreatedByUser
		serv.Noticef(message)
		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			ID:            primitive.NewObjectID(),
			StationName:   pStationName.Ext(),
			Message:       message,
			CreatedByUser: connection.CreatedByUser,
			CreationDate:  time.Now(),
			UserType:      "application",
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			serv.Errorf("createProducerDirectCommon: Producer " + pName + " at station " + pStationName.external + ": " + err.Error())
		}

		shouldSendAnalytics, _ := shouldSendAnalytics()
		if shouldSendAnalytics {
			param := analytics.EventParam{
				Name:  "producer-name",
				Value: newProducer.Name,
			}
			analyticsParams := []analytics.EventParam{param}
			analytics.SendEventWithParams(connection.CreatedByUser, analyticsParams, "user-create-producer-sdk")
		}
	}
	shouldSendNotifications, err := IsSlackEnabled()
	if err != nil {
		serv.Errorf("createProducerDirectCommon: Producer " + pName + " at station " + pStationName.external + ": " + err.Error())
	}

	return shouldSendNotifications, station.DlsConfiguration.Schemaverse, nil
}

func (s *Server) createProducerDirectV0(c *client, reply string, cpr createProducerRequestV0) {
	sn, err := StationNameFromStr(cpr.StationName)
	if err != nil {
		respondWithErr(s, reply, err)
		return
	}
	_, _, err = s.createProducerDirectCommon(c, cpr.Name,
		cpr.ProducerType, cpr.ConnectionId, sn)
	respondWithErr(s, reply, err)
}

func (s *Server) createProducerDirect(c *client, reply string, msg []byte) {
	var cpr createProducerRequestV1
	var resp createProducerResponse

	if err := json.Unmarshal(msg, &cpr); err != nil || cpr.RequestVersion < 1 {
		var cprV0 createProducerRequestV0
		if err := json.Unmarshal(msg, &cprV0); err != nil {
			s.Errorf("createProducerDirect: %v", err.Error())
			respondWithRespErr(s, reply, err, &resp)
			return
		}
		s.createProducerDirectV0(c, reply, cprV0)
		return
	}

	sn, err := StationNameFromStr(cpr.StationName)
	if err != nil {
		s.Errorf("createProducerDirect: Producer " + cpr.Name + " at station " + cpr.StationName + ": " + err.Error())
		respondWithRespErr(s, reply, err, &resp)
		return
	}

	clusterSendNotification, schemaVerseToDls, err := s.createProducerDirectCommon(c, cpr.Name, cpr.ProducerType, cpr.ConnectionId, sn)
	if err != nil {
		respondWithRespErr(s, reply, err, &resp)
		return
	}

	resp.SchemaVerseToDls = schemaVerseToDls
	resp.ClusterSendNotification = clusterSendNotification
	schemaUpdate, err := getSchemaUpdateInitFromStation(sn)
	if err == ErrNoSchema {
		respondWithResp(s, reply, &resp)
		return
	}
	if err != nil {
		s.Errorf("createProducerDirect: Producer " + cpr.Name + " at station " + cpr.StationName + ": " + err.Error())
		respondWithRespErr(s, reply, err, &resp)
		return
	}

	resp.SchemaUpdate = *schemaUpdate
	respondWithResp(s, reply, &resp)
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
		serv.Errorf("GetAllProducers: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &producers); err != nil {
		serv.Errorf("GetAllProducers: " + err.Error())
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

	stationName, err := StationNameFromStr(body.StationName)
	exist, station, err := IsStationExist(stationName)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("GetAllProducersByStation: Station " + body.StationName + " does not exist")
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
		serv.Errorf("GetAllProducersByStation: Station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &producers); err != nil {
		serv.Errorf("GetAllProducersByStation: Station " + body.StationName + ": " + err.Error())
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
		s.Errorf("destroyProducerDirect: %v", err.Error())
		respondWithErr(s, reply, err)
		return
	}

	stationName, err := StationNameFromStr(dpr.StationName)
	if err != nil {
		serv.Errorf("destroyProducerDirect: Producer " + dpr.ProducerName + "at station " + dpr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	name := strings.ToLower(dpr.ProducerName)
	_, station, err := IsStationExist(stationName)
	if err != nil {
		serv.Errorf("destroyProducerDirect: Producer " + dpr.ProducerName + "at station " + dpr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	var producer models.Producer
	err = producersCollection.FindOneAndUpdate(context.TODO(),
		bson.M{"name": name, "station_id": station.ID, "is_active": true},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	).Decode(&producer)

	if err == mongo.ErrNoDocuments {
		errMsg := "Producer " + name + " at station " + dpr.StationName + " does not exist"
		serv.Warnf("destroyProducerDirect: " + errMsg)
		respondWithErr(s, reply, errors.New(errMsg))
		return
	}
	if err != nil {
		serv.Errorf("destroyProducerDirect: Producer " + name + "at station " + dpr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	username := c.memphisInfo.username
	if username == "" {
		username = dpr.Username
	}

	message := "Producer " + name + " has been deleted by user " + username
	serv.Noticef(message)
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		ID:            primitive.NewObjectID(),
		StationName:   stationName.Ext(),
		Message:       message,
		CreatedByUser: username,
		CreationDate:  time.Now(),
		UserType:      "application",
	}
	auditLogs = append(auditLogs, newAuditLog)
	err = CreateAuditLogs(auditLogs)
	if err != nil {
		serv.Errorf("destroyProducerDirect: Producer " + name + "at station " + dpr.StationName + ": " + err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(username, "user-remove-producer-sdk")
	}

	respondWithErr(s, reply, nil)
}

func (ph ProducersHandler) ReliveProducers(connectionId primitive.ObjectID) error {
	_, err := producersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": connectionId, "is_deleted": false},
		bson.M{"$set": bson.M{"is_active": true}},
	)
	if err != nil {
		serv.Errorf("ReliveProducers: " + err.Error())
		return err
	}

	return nil
}
