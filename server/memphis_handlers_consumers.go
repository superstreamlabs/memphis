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
	"sort"

	"memphis-broker/analytics"
	"memphis-broker/models"
	"memphis-broker/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/utils/strings/slices"
)

type ConsumersHandler struct{ S *Server }

const (
	consumerObjectName = "Consumer"
)

func validateConsumerName(consumerName string) error {
	return validateName(consumerName, consumerObjectName)
}

func validateConsumerType(consumerType string) error {
	if consumerType != "application" && consumerType != "connector" {
		return errors.New("Consumer type has to be one of the following application/connector")
	}
	return nil
}

func isConsumerGroupExist(consumerGroup string, stationId primitive.ObjectID) (bool, models.Consumer, error) {
	filter := bson.M{"consumers_group": consumerGroup, "station_id": stationId, "is_deleted": false}
	var consumer models.Consumer
	err := consumersCollection.FindOne(context.TODO(), filter).Decode(&consumer)
	if err == mongo.ErrNoDocuments {
		return false, models.Consumer{}, nil
	} else if err != nil {
		return false, models.Consumer{}, err
	}
	return true, consumer, nil
}

func GetConsumerGroupMembers(cgName string, station models.Station) ([]models.CgMember, error) {
	var consumers []models.CgMember

	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"consumers_group", cgName}, {"station_id", station.ID}}}},
		bson.D{{"$sort", bson.D{{"creation_date", -1}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"name", 1}, {"created_by_user", 1}, {"is_active", 1}, {"is_deleted", 1}, {"max_ack_time_ms", 1}, {"max_msg_deliveries", 1}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	})
	if err != nil {
		return consumers, err
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		return consumers, err
	}

	var dedupedConsumers []models.CgMember
	consumersNames := []string{}

	for _, consumer := range consumers {
		if slices.Contains(consumersNames, consumer.Name) {
			continue
		}
		consumersNames = append(consumersNames, consumer.Name)
		dedupedConsumers = append(dedupedConsumers, consumer)
	}

	return dedupedConsumers, nil
}

func (s *Server) createConsumerDirect(c *client, reply string, msg []byte) {
	var ccr createConsumerRequest
	if err := json.Unmarshal(msg, &ccr); err != nil {
		s.Errorf("createConsumerDirect: Failed creating consumer: %v\n%v", err.Error(), string(msg))
		respondWithErr(s, reply, err)
		return
	}
	name := strings.ToLower(ccr.Name)
	err := validateConsumerName(name)
	if err != nil {
		serv.Warnf("createConsumerDirect: Failed creating consumer " + ccr.Name + " at station " + ccr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	consumerGroup := strings.ToLower(ccr.ConsumerGroup)
	if consumerGroup != "" {
		err = validateConsumerName(consumerGroup)
		if err != nil {
			serv.Warnf("createConsumerDirect: Failed creating consumer " + ccr.Name + " at station " + ccr.StationName + ": " + err.Error())
			respondWithErr(s, reply, err)
			return
		}
	} else {
		consumerGroup = name
	}

	consumerType := strings.ToLower(ccr.ConsumerType)
	err = validateConsumerType(consumerType)
	if err != nil {
		serv.Warnf("createConsumerDirect: Failed creating consumer " + ccr.Name + " at station " + ccr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	connectionIdObj, err := primitive.ObjectIDFromHex(ccr.ConnectionId)
	if err != nil {
		serv.Warnf("createConsumerDirect: Failed creating consumer " + ccr.Name + " at station " + ccr.StationName + ": Connection ID is not valid")
		respondWithErr(s, reply, err)
		return
	}
	exist, connection, err := IsConnectionExist(connectionIdObj)
	if err != nil {
		errMsg := "Consumer " + ccr.Name + ": " + err.Error()
		serv.Errorf("createConsumerDirect: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}
	if !exist {
		errMsg := "Consumer " + ccr.Name + " at station " + ccr.StationName + ": Connection ID " + ccr.ConnectionId + " was not found"
		serv.Warnf("createConsumerDirect: " + errMsg)
		respondWithErr(s, reply, errors.New(errMsg))
		return
	}
	if !connection.IsActive {
		serv.Warnf("createConsumerDirect: Failed creating consumer " + ccr.Name + " at station " + ccr.StationName + ": Connection is not active")
		respondWithErr(s, reply, errors.New("connection is not active"))
		return
	}

	stationName, err := StationNameFromStr(ccr.StationName)
	if err != nil {
		errMsg := "Consumer " + ccr.Name + " at station " + ccr.StationName + ": " + err.Error()
		serv.Errorf("createConsumerDirect: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}

	exist, station, err := IsStationExist(stationName)
	if err != nil {
		errMsg := "Consumer " + ccr.Name + " at station " + ccr.StationName + ": " + err.Error()
		serv.Errorf("createConsumerDirect: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}
	if !exist {
		var created bool
		station, created, err = CreateDefaultStation(s, stationName, connection.CreatedByUser)
		if err != nil {
			errMsg := "creating default station error: Consumer " + ccr.Name + " at station " + ccr.StationName + ": " + err.Error()
			serv.Errorf("createConsumerDirect: " + errMsg)
			respondWithErr(s, reply, err)
			return
		}

		if created {
			message := "Station " + stationName.Ext() + " has been created by user " + connection.CreatedByUser
			serv.Noticef(message)
			var auditLogs []interface{}
			newAuditLog := models.AuditLog{
				ID:            primitive.NewObjectID(),
				StationName:   stationName.Ext(),
				Message:       message,
				CreatedByUser: connection.CreatedByUser,
				CreationDate:  time.Now(),
				UserType:      "application",
			}
			auditLogs = append(auditLogs, newAuditLog)
			err = CreateAuditLogs(auditLogs)
			if err != nil {
				errMsg := "Consumer " + ccr.Name + " at station " + ccr.StationName + ": " + err.Error()
				serv.Errorf("createConsumerDirect: " + errMsg)
			}

			shouldSendAnalytics, _ := shouldSendAnalytics()
			if shouldSendAnalytics {
				param := analytics.EventParam{
					Name:  "station-name",
					Value: stationName.Ext(),
				}
				analyticsParams := []analytics.EventParam{param}
				analytics.SendEventWithParams(connection.CreatedByUser, analyticsParams, "user-create-station")
			}
		}
	}

	exist, _, err = IsConsumerExist(name, station.ID)
	if err != nil {
		errMsg := "Consumer " + ccr.Name + " at station " + ccr.StationName + ": " + err.Error()
		serv.Errorf("createConsumerDirect: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}
	if exist {
		errMsg := "Consumer " + ccr.Name + " at station " + ccr.StationName + ": Consumer name has to be unique per station"
		serv.Warnf("createConsumerDirect: " + errMsg)
		respondWithErr(s, reply, errors.New("memphis: "+errMsg))
		return
	}

	consumerGroupExist, consumerFromGroup, err := isConsumerGroupExist(consumerGroup, station.ID)
	if err != nil {
		errMsg := "Consumer " + ccr.Name + " at station " + ccr.StationName + ": " + err.Error()
		serv.Errorf("createConsumerDirect: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}

	newConsumer := models.Consumer{
		ID:               primitive.NewObjectID(),
		Name:             name,
		StationId:        station.ID,
		Type:             consumerType,
		ConnectionId:     connectionIdObj,
		CreatedByUser:    connection.CreatedByUser,
		ConsumersGroup:   consumerGroup,
		IsActive:         true,
		CreationDate:     time.Now(),
		IsDeleted:        false,
		MaxAckTimeMs:     int64(ccr.MaxAckTimeMillis),
		MaxMsgDeliveries: ccr.MaxMsgDeliveries,
	}

	if consumerGroupExist {
		if newConsumer.MaxAckTimeMs != consumerFromGroup.MaxAckTimeMs || newConsumer.MaxMsgDeliveries != consumerFromGroup.MaxMsgDeliveries {
			err := s.CreateConsumer(newConsumer, station)
			if err != nil {
				errMsg := "Consumer " + ccr.Name + " at station " + ccr.StationName + ": " + err.Error()
				serv.Errorf("createConsumerDirect: " + errMsg)
				respondWithErr(s, reply, err)
				return
			}
		}
	} else {
		err := s.CreateConsumer(newConsumer, station)
		if err != nil {
			errMsg := "Consumer " + ccr.Name + " at station " + ccr.StationName + ": " + err.Error()
			serv.Errorf("createConsumerDirect: " + errMsg)
			respondWithErr(s, reply, err)
			return
		}
	}

	filter := bson.M{"name": newConsumer.Name, "station_id": station.ID, "is_active": true, "is_deleted": false}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":                newConsumer.ID,
			"type":               newConsumer.Type,
			"connection_id":      newConsumer.ConnectionId,
			"created_by_user":    newConsumer.CreatedByUser,
			"consumers_group":    newConsumer.ConsumersGroup,
			"creation_date":      newConsumer.CreationDate,
			"max_ack_time_ms":    newConsumer.MaxAckTimeMs,
			"max_msg_deliveries": newConsumer.MaxMsgDeliveries,
		},
	}
	opts := options.Update().SetUpsert(true)
	updateResults, err := consumersCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		errMsg := "Consumer " + ccr.Name + " at station " + ccr.StationName + ": " + err.Error()
		serv.Errorf("createConsumerDirect: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}

	if updateResults.MatchedCount == 0 {
		message := "Consumer " + name + " has been created by user " + connection.CreatedByUser
		serv.Noticef(message)
		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			ID:            primitive.NewObjectID(),
			StationName:   stationName.Ext(),
			Message:       message,
			CreatedByUser: connection.CreatedByUser,
			CreationDate:  time.Now(),
			UserType:      "application",
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			errMsg := "Consumer " + ccr.Name + " at station " + ccr.StationName + ": " + err.Error()
			serv.Errorf("createConsumerDirect: " + errMsg)
		}

		shouldSendAnalytics, _ := shouldSendAnalytics()
		if shouldSendAnalytics {
			param := analytics.EventParam{
				Name:  "consumer-name",
				Value: newConsumer.Name,
			}
			analyticsParams := []analytics.EventParam{param}
			analytics.SendEventWithParams(connection.CreatedByUser, analyticsParams, "user-create-consumer")
		}
	}

	respondWithErr(s, reply, nil)
	return
}

func (ch ConsumersHandler) GetAllConsumers(c *gin.Context) {
	var consumers []models.ExtendedConsumer
	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"consumers_group", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"max_ack_time_ms", 1}, {"max_msg_deliveries", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
	})
	if err != nil {
		serv.Errorf("GetAllConsumers: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		serv.Errorf("GetAllConsumers: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(consumers) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, consumers)
	}
}

func (ch ConsumersHandler) GetCgsByStation(stationName StationName, station models.Station, poisonedCgMap map[string]int) ([]models.Cg, []models.Cg, []models.Cg, error) { // for socket io endpoint
	var cgs []models.Cg
	var consumers []models.ExtendedConsumer

	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"station_id", station.ID}}}},
		bson.D{{"$sort", bson.D{{"creation_date", -1}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"name", 1}, {"created_by_user", 1}, {"consumers_group", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"max_ack_time_ms", 1}, {"max_msg_deliveries", 1}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"connection", 0}}}},
	})
	if err != nil {
		return cgs, cgs, cgs, err
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		return cgs, cgs, cgs, err
	}

	if len(consumers) == 0 {
		return []models.Cg{}, []models.Cg{}, []models.Cg{}, nil
	}

	m := make(map[string]*models.Cg)
	consumersNames := []string{}

	for _, consumer := range consumers {
		if slices.Contains(consumersNames, consumer.Name) {
			continue
		}
		consumersNames = append(consumersNames, consumer.Name)

		var cg *models.Cg
		if m[consumer.ConsumersGroup] == nil {
			cg = &models.Cg{
				Name:                  consumer.ConsumersGroup,
				MaxAckTimeMs:          consumer.MaxAckTimeMs,
				MaxMsgDeliveries:      consumer.MaxMsgDeliveries,
				ConnectedConsumers:    []models.ExtendedConsumer{},
				DisconnectedConsumers: []models.ExtendedConsumer{},
				DeletedConsumers:      []models.ExtendedConsumer{},
				IsActive:              consumer.IsActive,
				IsDeleted:             consumer.IsDeleted,
				LastStatusChangeDate:  consumer.CreationDate,
			}
			m[consumer.ConsumersGroup] = cg
		} else {
			cg = m[consumer.ConsumersGroup]
		}

		if consumer.IsActive {
			cg.ConnectedConsumers = append(cg.ConnectedConsumers, consumer)
		} else if !consumer.IsDeleted && !consumer.IsActive {
			cg.DisconnectedConsumers = append(cg.DisconnectedConsumers, consumer)
		} else if consumer.IsDeleted {
			cg.DeletedConsumers = append(cg.DeletedConsumers, consumer)
		}
	}

	var connectedCgs []models.Cg
	var disconnectedCgs []models.Cg
	var deletedCgs []models.Cg

	for _, cg := range m {
		if cg.IsDeleted {
			cg.IsActive = false
			cg.IsDeleted = true
		} else { // not deleted
			cgInfo, err := ch.S.GetCgInfo(stationName, cg.Name)
			if err != nil {
				return cgs, cgs, cgs, err
			}

			totalPoisonMsgs := 0
			if _, ok := poisonedCgMap[cg.Name]; ok {
				totalPoisonMsgs = poisonedCgMap[cg.Name]
			}

			cg.InProcessMessages = cgInfo.NumAckPending
			cg.UnprocessedMessages = int(cgInfo.NumPending)
			cg.PoisonMessages = totalPoisonMsgs
		}

		if len(cg.ConnectedConsumers) > 0 {
			connectedCgs = append(connectedCgs, *cg)
		} else if len(cg.DisconnectedConsumers) > 0 {
			disconnectedCgs = append(disconnectedCgs, *cg)
		} else {
			deletedCgs = append(deletedCgs, *cg)
		}
	}

	if len(connectedCgs) == 0 {
		connectedCgs = []models.Cg{}
	}

	if len(disconnectedCgs) == 0 {
		disconnectedCgs = []models.Cg{}
	}

	if len(deletedCgs) == 0 {
		deletedCgs = []models.Cg{}
	}

	sort.Slice(connectedCgs, func(i, j int) bool {
		return connectedCgs[j].LastStatusChangeDate.Before(connectedCgs[i].LastStatusChangeDate)
	})
	sort.Slice(disconnectedCgs, func(i, j int) bool {
		return disconnectedCgs[j].LastStatusChangeDate.Before(disconnectedCgs[i].LastStatusChangeDate)
	})
	sort.Slice(deletedCgs, func(i, j int) bool {
		return deletedCgs[j].LastStatusChangeDate.Before(deletedCgs[i].LastStatusChangeDate)
	})
	return connectedCgs, disconnectedCgs, deletedCgs, nil
}

// TODO fix it
func (ch ConsumersHandler) GetAllConsumersByStation(c *gin.Context) { // for REST endpoint
	var body models.GetAllConsumersByStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	sn, err := StationNameFromStr(body.StationName)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	exist, station, err := IsStationExist(sn)
	if err != nil {
		serv.Errorf("GetAllConsumersByStation: At station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("GetAllConsumersByStation: Station " + body.StationName + " does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station does not exist"})
		return
	}

	var consumers []models.ExtendedConsumer
	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"station_id", station.ID}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"consumers_group", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"max_ack_time_ms", 1}, {"max_msg_deliveries", 1}, {"station_name", "$station.name"}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"connection", 0}}}},
	})
	if err != nil {
		errMsg := "Station " + body.StationName + ": " + err.Error()
		serv.Errorf("GetAllConsumersByStation: " + errMsg)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		errMsg := "Station " + body.StationName + ": " + err.Error()
		serv.Errorf("GetAllConsumersByStation: " + errMsg)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(consumers) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, consumers)
	}
}

func (s *Server) destroyConsumerDirect(c *client, reply string, msg []byte) {
	var dcr destroyConsumerRequest
	if err := json.Unmarshal(msg, &dcr); err != nil {
		s.Errorf("destroyConsumerDirect: %v", err.Error())
		respondWithErr(s, reply, err)
		return
	}

	stationName, err := StationNameFromStr(dcr.StationName)
	if err != nil {
		errMsg := "Station " + dcr.StationName + ": " + err.Error()
		serv.Errorf("DestroyConsumer: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}

	name := strings.ToLower(dcr.ConsumerName)
	_, station, err := IsStationExist(stationName)
	if err != nil {
		errMsg := "Station " + dcr.StationName + ": " + err.Error()
		serv.Errorf("DestroyConsumer: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}

	var consumer models.Consumer
	err = consumersCollection.FindOneAndUpdate(context.TODO(),
		bson.M{"name": name, "station_id": station.ID, "is_active": true},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	).Decode(&consumer)
	if err == mongo.ErrNoDocuments {
		errMsg := "Consumer " + dcr.ConsumerName + " at station " + dcr.StationName + " does not exist"
		serv.Warnf("DestroyConsumer: " + errMsg)
		respondWithErr(s, reply, errors.New(errMsg))
		return
	}
	if err != nil {
		errMsg := "Consumer " + dcr.ConsumerName + " at station " + dcr.StationName + ": " + err.Error()
		serv.Errorf("DestroyConsumer: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}

	_, err = consumersCollection.UpdateMany(context.TODO(),
		bson.M{"name": name, "station_id": station.ID},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	)
	if err != nil {
		errMsg := "Consumer " + dcr.ConsumerName + " at station " + dcr.StationName + ": " + err.Error()
		serv.Errorf("DestroyConsumer: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}

	// ensure not part of an active consumer group
	count, err := consumersCollection.CountDocuments(context.TODO(), bson.M{"station_id": station.ID, "consumers_group": consumer.ConsumersGroup, "is_deleted": false})
	if err != nil {
		errMsg := "Consumer " + dcr.ConsumerName + " at station " + dcr.StationName + ": " + err.Error()
		serv.Errorf("DestroyConsumer: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}

	if count == 0 { // no other members in this group
		err = s.RemoveConsumer(stationName, consumer.ConsumersGroup)
		if err != nil && !IsNatsErr(err, JSConsumerNotFoundErr) {
			errMsg := "Consumer group " + consumer.ConsumersGroup + " at station " + dcr.StationName + ": " + err.Error()
			serv.Errorf("DestroyConsumer: " + errMsg)
			respondWithErr(s, reply, err)
			return
		}

		err = RemovePoisonedCg(stationName, consumer.ConsumersGroup)
		if err != nil {
			errMsg := "Consumer group " + consumer.ConsumersGroup + " at station " + dcr.StationName + ": " + err.Error()
			serv.Errorf("DestroyConsumer: " + errMsg)
			respondWithErr(s, reply, err)
			return
		}
	}

	username := c.memphisInfo.username
	if username == "" {
		username = dcr.Username
	}

	message := "Consumer " + name + " has been deleted by user " + username
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
		errMsg := "Consumer group " + consumer.ConsumersGroup + " at station " + dcr.StationName + ": " + err.Error()
		serv.Errorf("DestroyConsumer: " + errMsg)
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(username, "user-remove-consumer")
	}

	respondWithErr(s, reply, nil)
	return
}

func (ch ConsumersHandler) KillConsumers(connectionId primitive.ObjectID) error {
	var consumers []models.Consumer
	var station models.Station

	cursor, err := consumersCollection.Find(context.TODO(), bson.M{"connection_id": connectionId, "is_active": true})
	if err != nil {
		serv.Errorf("KillConsumers: " + err.Error())
	}
	if err = cursor.All(context.TODO(), &consumers); err != nil {
		serv.Errorf("KillConsumers: " + err.Error())
	}

	if len(consumers) > 0 {
		err = stationsCollection.FindOne(context.TODO(), bson.M{"_id": consumers[0].StationId}).Decode(&station)
		if err != nil {
			errMsg := "At station ID: " + consumers[0].StationId.Hex() + ": " + err.Error()
			serv.Errorf("KillConsumers: " + errMsg)
		}
		_, err = consumersCollection.UpdateMany(context.TODO(),
			bson.M{"connection_id": connectionId},
			bson.M{"$set": bson.M{"is_active": false}},
		)
		if err != nil {
			errMsg := "At station: " + station.Name + ": " + err.Error()
			serv.Errorf("KillConsumers: " + errMsg)
			return err
		}

		userType := "application"
		if consumers[0].CreatedByUser == "root" {
			userType = "root"
		}

		var message string
		var auditLogs []interface{}
		var newAuditLog models.AuditLog
		for _, consumer := range consumers {
			message = "Consumer " + consumer.Name + " has been disconnected by user " + consumers[0].CreatedByUser
			newAuditLog = models.AuditLog{
				ID:            primitive.NewObjectID(),
				StationName:   station.Name,
				Message:       message,
				CreatedByUser: consumers[0].CreatedByUser,
				CreationDate:  time.Now(),
				UserType:      userType,
			}
			auditLogs = append(auditLogs, newAuditLog)
		}
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			errMsg := "At station: " + station.Name + ": " + err.Error()
			serv.Errorf("KillConsumers: " + errMsg)
		}
	}

	return nil
}

func (ch ConsumersHandler) ReliveConsumers(connectionId primitive.ObjectID) error {
	_, err := consumersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": connectionId, "is_deleted": false},
		bson.M{"$set": bson.M{"is_active": true}},
	)
	if err != nil {
		serv.Errorf("ReliveConsumers: " + err.Error())
		return err
	}

	return nil
}
