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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StationsHandler struct{ S *Server }

const (
	stationObjectName = "Station"
)

type StationName struct {
	internal string
	external string
}

func (sn StationName) Ext() string {
	return sn.external
}

func (sn StationName) Intern() string {
	return sn.internal
}

func StationNameFromStr(name string) (StationName, error) {
	extern := strings.ToLower(name)

	err := validateName(extern, stationObjectName)
	if err != nil {
		return StationName{}, err
	}

	intern := replaceDelimiters(name)

	return StationName{internal: intern, external: extern}, nil
}

func StationNameFromStreamName(streamName string) StationName {
	intern := streamName
	extern := revertDelimiters(intern)

	return StationName{internal: intern, external: extern}
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
func removeStationResources(s *Server, station models.Station) error {
	err := s.RemoveStream(station.Name)
	if err != nil {
		return err
	}

	DeleteTagsByStation(station.ID)

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

	err = RemovePoisonMsgsByStation(station.Name)
	if err != nil {
		serv.Warnf("removeStationResources error: " + err.Error())
	}

	err = RemoveAllAuditLogsByStation(station.Name)
	if err != nil {
		serv.Warnf("removeStationResources error: " + err.Error())
	}

	return nil
}

func (s *Server) createStationDirect(c *client, reply string, msg []byte) {
	var csr createStationRequest
	if err := json.Unmarshal(msg, &csr); err != nil {
		s.Warnf("failed creating station: %v", err.Error())
		respondWithErr(s, reply, err)
		return
	}
	stationName, err := StationNameFromStr(csr.StationName)
	if err != nil {
		serv.Warnf(err.Error())
		respondWithErr(s, reply, err)
		return
	}

	exist, _, err := IsStationExist(stationName)
	if err != nil {
		serv.Errorf("CreateStation error: " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	if exist {
		serv.Warnf("Station with that name already exists")
		respondWithErr(s, reply, errors.New("memphis: station with that name already exists"))
		return
	}

	var retentionType string
	var retentionValue int
	if csr.RetentionType != "" {
		retentionType = strings.ToLower(csr.RetentionType)
		err = validateRetentionType(retentionType)
		if err != nil {
			serv.Warnf(err.Error())
			respondWithErr(s, reply, err)
			return
		}
		retentionValue = csr.RetentionValue
	} else {
		retentionType = "message_age_sec"
		retentionValue = 604800 // 1 week
	}

	var storageType string
	if csr.StorageType != "" {
		storageType = strings.ToLower(csr.StorageType)
		err = validateStorageType(storageType)
		if err != nil {
			serv.Warnf(err.Error())
			respondWithErr(s, reply, err)
			return
		}
	} else {
		storageType = "file"
	}

	replicas := csr.Replicas
	if replicas > 0 {
		err = validateReplicas(replicas)
		if err != nil {
			serv.Warnf(err.Error())
			respondWithErr(s, reply, err)
			return
		}
	} else {
		replicas = 1
	}
	newStation := models.Station{
		ID:              primitive.NewObjectID(),
		Name:            stationName.Ext(),
		CreatedByUser:   c.memphisInfo.username,
		CreationDate:    time.Now(),
		IsDeleted:       false,
		RetentionType:   retentionType,
		RetentionValue:  retentionValue,
		StorageType:     storageType,
		Replicas:        replicas,
		DedupEnabled:    csr.DedupEnabled,
		DedupWindowInMs: csr.DedupWindowMillis,
		LastUpdate:      time.Now(),
		Functions:       []models.Function{},
	}

	err = s.CreateStream(stationName, newStation)
	if err != nil {
		serv.Warnf(err.Error())
		respondWithErr(s, reply, err)
		return
	}

	_, err = stationsCollection.InsertOne(context.TODO(), newStation)
	if err != nil {
		serv.Errorf("CreateStation error: " + err.Error())
		respondWithErr(s, reply, err)
		return
	}
	message := "Station " + stationName.Ext() + " has been created"
	serv.Noticef(message)

	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		ID:            primitive.NewObjectID(),
		StationName:   stationName.Ext(),
		Message:       message,
		CreatedByUser: c.memphisInfo.username,
		CreationDate:  time.Now(),
		UserType:      "application",
	}
	auditLogs = append(auditLogs, newAuditLog)
	err = CreateAuditLogs(auditLogs)
	if err != nil {
		serv.Warnf("create audit logs error: " + err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(c.memphisInfo.username, "user-create-station")
	}

	respondWithErr(s, reply, nil)
	return
}

func (sh StationsHandler) GetStation(c *gin.Context) {
	var body models.GetStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	tagsHandler := TagsHandler{S: sh.S}

	var station models.GetStationResponseSchema
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
		serv.Errorf("GetStationById error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	tags, err := tagsHandler.GetTagsByStation(station.ID)
	if err != nil {
		serv.Errorf("GetStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	station.Tags = tags

	c.IndentedJSON(200, station)
}

func (sh StationsHandler) GetStationsDetails() ([]models.ExtendedStationDetails, error) {
	var exStations []models.ExtendedStationDetails
	var stations []models.Station

	poisonMsgsHandler := PoisonMessagesHandler{S: sh.S}
	cursor, err := stationsCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"$or", []interface{}{
			bson.D{{"is_deleted", false}},
			bson.D{{"is_deleted", bson.D{{"$exists", false}}}},
		}}}}},
	})

	if err != nil {
		return []models.ExtendedStationDetails{}, err
	}

	if err = cursor.All(context.TODO(), &stations); err != nil {
		return []models.ExtendedStationDetails{}, err
	}

	if len(stations) == 0 {
		return []models.ExtendedStationDetails{}, nil
	} else {
		tagsHandler := TagsHandler{S: sh.S}
		for _, station := range stations {
			totalMessages, err := sh.GetTotalMessages(station.Name)
			if err != nil {
				return []models.ExtendedStationDetails{}, err
			}
			poisonMessages, err := poisonMsgsHandler.GetTotalPoisonMsgsByStation(station.Name)
			if err != nil {
				return []models.ExtendedStationDetails{}, err
			}
			tags, err := tagsHandler.GetTagsByStation(station.ID)
			if err != nil {
				return []models.ExtendedStationDetails{}, err
			}
			exStations = append(exStations, models.ExtendedStationDetails{Station: station, PoisonMessages: poisonMessages, TotalMessages: totalMessages, Tags: tags})
		}
		return exStations, nil
	}
}

func (sh StationsHandler) GetAllStationsDetails() ([]models.ExtendedStation, error) {
	var stations []models.ExtendedStation
	cursor, err := stationsCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"$or", []interface{}{
			bson.D{{"is_deleted", false}},
			bson.D{{"is_deleted", bson.D{{"$exists", false}}}},
		}}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"retention_type", 1}, {"retention_value", 1}, {"storage_type", 1}, {"replicas", 1}, {"dedup_enabled", 1}, {"dedup_window_in_ms", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"last_update", 1}, {"functions", 1}}}},
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
		poisonMsgsHandler := PoisonMessagesHandler{S: sh.S}
		tagsHandler := TagsHandler{S: sh.S}
		for i := 0; i < len(stations); i++ {
			totalMessages, err := sh.GetTotalMessages(stations[i].Name)
			if err != nil {
				return []models.ExtendedStation{}, err
			}
			poisonMessages, err := poisonMsgsHandler.GetTotalPoisonMsgsByStation(stations[i].Name)
			if err != nil {
				return []models.ExtendedStation{}, err
			}
			tags, err := tagsHandler.GetTagsByStation(stations[i].ID)
			if err != nil {
				return []models.ExtendedStation{}, err
			}

			stations[i].TotalMessages = totalMessages
			stations[i].PoisonMessages = poisonMessages
			stations[i].Tags = tags
		}
		return stations, nil
	}
}

func (sh StationsHandler) GetStations(c *gin.Context) {
	stations, err := sh.GetStationsDetails()
	if err != nil {
		serv.Errorf("GetStations error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, gin.H{
		"stations": stations,
	})
}

func (sh StationsHandler) GetAllStations(c *gin.Context) {
	stations, err := sh.GetAllStationsDetails()
	if err != nil {
		serv.Errorf("GetAllStations error: " + err.Error())
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

	stationName, err := StationNameFromStr(body.Name)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, _, err := IsStationExist(stationName)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		serv.Warnf("Station with the same name is already exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station with the same name is already exist"})
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("CreateStation error: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}

	var retentionType string
	if body.RetentionType != "" && body.RetentionValue > 0 {
		retentionType = strings.ToLower(body.RetentionType)
		err = validateRetentionType(retentionType)
		if err != nil {
			serv.Warnf(err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	} else {
		retentionType = "message_age_sec"
		body.RetentionValue = 604800 // 1 week
	}

	if body.StorageType != "" {
		body.StorageType = strings.ToLower(body.StorageType)
		err = validateStorageType(body.StorageType)
		if err != nil {
			serv.Warnf(err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	} else {
		body.StorageType = "file"
	}

	if body.Replicas > 0 {
		err = validateReplicas(body.Replicas)
		if err != nil {
			serv.Warnf(err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	} else {
		body.Replicas = 1
	}

	newStation := models.Station{
		ID:              primitive.NewObjectID(),
		Name:            stationName.Ext(),
		RetentionType:   retentionType,
		RetentionValue:  body.RetentionValue,
		StorageType:     body.StorageType,
		Replicas:        body.Replicas,
		DedupEnabled:    body.DedupEnabled,
		DedupWindowInMs: body.DedupWindowInMs,
		CreatedByUser:   user.Username,
		CreationDate:    time.Now(),
		LastUpdate:      time.Now(),
		Functions:       []models.Function{},
		IsDeleted:       false,
	}

	err = sh.S.CreateStream(stationName, newStation)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
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
		serv.Errorf("CreateStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if updateResults.MatchedCount > 0 {
		serv.Warnf("Station with the same name is already exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station with the same name is already exist"})
		return
	}

	if len(body.Tags) > 0 {
		err = AddTagsToEntity(body.Tags, "station", newStation.ID)
		if err != nil {
			serv.Errorf("Failed creating tag: %v", err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	message := "Station " + stationName.Ext() + " has been created"
	serv.Noticef(message)
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		ID:            primitive.NewObjectID(),
		StationName:   stationName.Ext(),
		Message:       message,
		CreatedByUser: user.Username,
		CreationDate:  time.Now(),
		UserType:      user.UserType,
	}
	auditLogs = append(auditLogs, newAuditLog)
	err = CreateAuditLogs(auditLogs)
	if err != nil {
		serv.Warnf("CreateStation error: " + err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-create-station")
	}

	c.IndentedJSON(200, newStation)
}

func (sh StationsHandler) RemoveStation(c *gin.Context) {
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}
	var body models.RemoveStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, station, err := IsStationExist(stationName)
	if err != nil {
		serv.Errorf("RemoveStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("Station does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station does not exist"})
		return
	}

	err = removeStationResources(sh.S, station)
	if err != nil {
		serv.Errorf("RemoveStation error: " + err.Error())
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
		serv.Errorf("RemoveStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-remove-station")
	}

	serv.Noticef("Station " + stationName.Ext() + " has been deleted")
	c.IndentedJSON(200, gin.H{})
}

func (s *Server) removeStationDirect(reply string, msg []byte) {
	var dsr destroyStationRequest
	if err := json.Unmarshal(msg, &dsr); err != nil {
		s.Warnf("failed destroying station: %v", err.Error())
		respondWithErr(s, reply, err)
		return
	}
	stationName, err := StationNameFromStr(dsr.StationName)
	if err != nil {
		serv.Errorf("RemoveStation error: " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	exist, station, err := IsStationExist(stationName)
	if err != nil {
		serv.Errorf("RemoveStation error: " + err.Error())
		respondWithErr(s, reply, err)
		return
	}
	if !exist {
		serv.Warnf("Station does not exist")
		respondWithErr(s, reply, err)
		return
	}

	err = removeStationResources(s, station)
	if err != nil {
		serv.Errorf("RemoveStation error: " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	_, err = stationsCollection.UpdateOne(context.TODO(),
		bson.M{
			"name": stationName.Ext(),
			"$or": []interface{}{
				bson.M{"is_deleted": false},
				bson.M{"is_deleted": bson.M{"$exists": false}},
			},
		},
		bson.M{"$set": bson.M{"is_deleted": true}},
	)
	if err != nil {
		serv.Errorf("RemoveStation error: " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	serv.Noticef("Station " + stationName.Ext() + " has been deleted")
	respondWithErr(s, reply, nil)
	return
}

func (sh StationsHandler) GetTotalMessages(stationNameExt string) (int, error) {
	stationName, err := StationNameFromStr(stationNameExt)
	if err != nil {
		return 0, err
	}
	totalMessages, err := sh.S.GetTotalMessagesInStation(stationName)
	return totalMessages, err
}

func (sh StationsHandler) GetTotalMessagesAcrossAllStations() (int, error) {
	totalMessages, err := sh.S.GetTotalMessagesAcrossAllStations()
	return totalMessages, err
}

func (sh StationsHandler) GetAvgMsgSize(station models.Station) (int64, error) {
	avgMsgSize, err := sh.S.GetAvgMsgSizeInStation(station)
	return avgMsgSize, err
}

func (sh StationsHandler) GetMessages(station models.Station, messagesToFetch int) ([]models.MessageDetails, error) {
	messages, err := sh.S.GetMessages(station, messagesToFetch)
	if err != nil {
		return messages, err
	}

	return messages, nil
}

func getCgStatus(members []models.CgMember) (bool, bool) {
	deletedCount := 0
	for _, member := range members {
		if member.IsActive {
			return true, false
		}

		if member.IsDeleted {
			deletedCount++
		}
	}

	if len(members) == deletedCount {
		return false, true
	}

	return false, false
}

func (sh StationsHandler) GetPoisonMessageJourneyDetails(poisonMsgId string) (models.PoisonMessage, error) {
	messageId, _ := primitive.ObjectIDFromHex(poisonMsgId)
	poisonMessage, err := GetPoisonMsgById(messageId)
	if err != nil {
		return poisonMessage, err
	}

	stationName, err := StationNameFromStr(poisonMessage.StationName)
	if err != nil {
		return poisonMessage, err
	}
	exist, station, err := IsStationExist(stationName)
	if err != nil {
		return poisonMessage, err
	}
	if !exist {
		return poisonMessage, errors.New("Station does not exist")
	}

	filter := bson.M{"name": poisonMessage.Producer.Name, "station_id": station.ID, "connection_id": poisonMessage.Producer.ConnectionId}
	var producer models.Producer
	err = producersCollection.FindOne(context.TODO(), filter).Decode(&producer)
	if err == mongo.ErrNoDocuments {
		return poisonMessage, errors.New("Producer does not exist")
	} else if err != nil {
		return poisonMessage, err
	}

	poisonMessage.Producer.CreatedByUser = producer.CreatedByUser
	poisonMessage.Producer.IsActive = producer.IsActive
	poisonMessage.Producer.IsDeleted = producer.IsDeleted

	for i, _ := range poisonMessage.PoisonedCgs {
		cgMembers, err := GetConsumerGroupMembers(poisonMessage.PoisonedCgs[i].CgName, station)
		if err != nil {
			return poisonMessage, err
		}

		isActive, isDeleted := getCgStatus(cgMembers)

		stationName, err := StationNameFromStr(poisonMessage.StationName)
		if err != nil {
			return poisonMessage, err
		}
		cgInfo, err := sh.S.GetCgInfo(stationName, poisonMessage.PoisonedCgs[i].CgName)
		if err != nil {
			return poisonMessage, err
		}

		totalPoisonMsgs, err := GetTotalPoisonMsgsByCg(poisonMessage.StationName, poisonMessage.PoisonedCgs[i].CgName)
		if err != nil {
			return poisonMessage, err
		}

		poisonMessage.PoisonedCgs[i].MaxAckTimeMs = cgMembers[0].MaxAckTimeMs
		poisonMessage.PoisonedCgs[i].MaxMsgDeliveries = cgMembers[0].MaxMsgDeliveries
		poisonMessage.PoisonedCgs[i].UnprocessedMessages = int(cgInfo.NumPending)
		poisonMessage.PoisonedCgs[i].InProcessMessages = cgInfo.NumAckPending
		poisonMessage.PoisonedCgs[i].TotalPoisonMessages = totalPoisonMsgs
		poisonMessage.PoisonedCgs[i].CgMembers = cgMembers
		poisonMessage.PoisonedCgs[i].IsActive = isActive
		poisonMessage.PoisonedCgs[i].IsDeleted = isDeleted
	}

	return poisonMessage, nil
}

func (sh StationsHandler) GetPoisonMessageJourney(c *gin.Context) {
	var body models.GetPoisonMessageJourneySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	poisonMessage, err := sh.GetPoisonMessageJourneyDetails(body.MessageId)
	if err == mongo.ErrNoDocuments {
		serv.Warnf("GetPoisonMessageJourney error: " + err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Poison message does not exist"})
		return
	}
	if err != nil {
		serv.Errorf("GetPoisonMessageJourney error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-enter-message-journey")
	}

	c.IndentedJSON(200, poisonMessage)
}

func (sh StationsHandler) AckPoisonMessages(c *gin.Context) {
	var body models.AckPoisonMessagesSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	_, err := poisonMessagesCollection.DeleteMany(context.TODO(), bson.M{"_id": bson.M{"$in": body.PoisonMessageIds}})
	if err != nil {
		serv.Errorf("AckPoisonMessage error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-ack-poison-message")
	}

	c.IndentedJSON(200, gin.H{})
}

func (sh StationsHandler) ResendPoisonMessages(c *gin.Context) {
	var body models.ResendPoisonMessagesSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	var msgs []models.PoisonMessage
	cursor, err := poisonMessagesCollection.Find(context.TODO(), bson.M{"_id": bson.M{"$in": body.PoisonMessageIds}})
	if err != nil {
		serv.Errorf("ResendPoisonMessages error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if err = cursor.All(context.TODO(), &msgs); err != nil {
		serv.Errorf("ResendPoisonMessages error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	for _, msg := range msgs {
		stationName := replaceDelimiters(msg.StationName)
		for _, cg := range msg.PoisonedCgs {
			cgName := replaceDelimiters(cg.CgName)
			err := sh.S.ResendPoisonMessage("$memphis_dlq_"+stationName+"_"+cgName, []byte(msg.Message.Data))
			if err != nil {
				serv.Errorf("ResendPoisonMessages error: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
		}
	}

	if err != nil {
		serv.Errorf("ResendPoisonMessages error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-resend-poison-message")
	}

	c.IndentedJSON(200, gin.H{})
}

func (sh StationsHandler) GetMessageDetails(c *gin.Context) {
	var body models.GetMessageDetailsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	if body.IsPoisonMessage {
		poisonMessage, err := sh.GetPoisonMessageJourneyDetails(body.MessageId)
		if err == mongo.ErrNoDocuments {
			serv.Warnf("GetMessageDetails error: " + err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Poison message does not exist"})
			return
		}
		if err != nil {
			serv.Errorf("GetMessageDetails error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		c.IndentedJSON(200, poisonMessage)
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Errorf("GetMessageDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	exist, station, err := IsStationExist(stationName)
	if !exist {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station does not exist"})
		return
	}
	if err != nil {
		serv.Errorf("GetMessageDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	sm, err := sh.S.GetMessage(stationName, uint64(body.MessageSeq))

	if err != nil {
		serv.Errorf("GetMessageDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	hdr, err := DecodeHeader(sm.Header)
	if err != nil {
		serv.Errorf("GetMessageDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	connectionIdHeader := hdr["connectionId"]
	producedByHeader := strings.ToLower(hdr["producedBy"])

	if connectionIdHeader == "" || producedByHeader == "" {
		serv.Errorf("Error while getting notified about a poison message: Missing mandatory message headers, please upgrade the SDK version you are using")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Error while getting notified about a poison message: Missing mandatory message headers, please upgrade the SDK version you are using"})
		return
	}

	connectionId, _ := primitive.ObjectIDFromHex(connectionIdHeader)
	poisonedCgs, err := GetPoisonedCgsByMessage(stationName.Ext(), models.MessageDetails{MessageSeq: int(sm.Sequence), ProducedBy: producedByHeader, TimeSent: sm.Time})
	if err != nil {
		serv.Errorf("GetMessageDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	for i, cg := range poisonedCgs {
		cgInfo, err := sh.S.GetCgInfo(stationName, cg.CgName)
		if err != nil {
			serv.Errorf("GetMessageDetails error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		totalPoisonMsgs, err := GetTotalPoisonMsgsByCg(stationName.Ext(), cg.CgName)
		if err != nil {
			serv.Errorf("GetMessageDetails error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		cgMembers, err := GetConsumerGroupMembers(cg.CgName, station)
		if err != nil {
			serv.Errorf("GetMessageDetails error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		isActive, isDeleted := getCgStatus(cgMembers)

		poisonedCgs[i].MaxAckTimeMs = cgMembers[0].MaxAckTimeMs
		poisonedCgs[i].MaxMsgDeliveries = cgMembers[0].MaxMsgDeliveries
		poisonedCgs[i].UnprocessedMessages = int(cgInfo.NumPending)
		poisonedCgs[i].InProcessMessages = cgInfo.NumAckPending
		poisonedCgs[i].TotalPoisonMessages = totalPoisonMsgs
		poisonedCgs[i].IsActive = isActive
		poisonedCgs[i].IsDeleted = isDeleted
	}

	filter := bson.M{"name": producedByHeader, "station_id": station.ID, "connection_id": connectionId}
	var producer models.Producer
	err = producersCollection.FindOne(context.TODO(), filter).Decode(&producer)
	if err != nil {
		serv.Errorf("GetMessageDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	_, conn, err := IsConnectionExist(connectionId)
	if err != nil {
		serv.Errorf("GetMessageDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	msg := models.Message{
		MessageSeq: body.MessageSeq,
		Message: models.MessagePayload{
			TimeSent: sm.Time,
			Size:     len(sm.Subject) + len(sm.Data) + len(sm.Header),
			Data:     string(sm.Data),
		},
		Producer: models.ProducerDetails{
			Name:          producedByHeader,
			ConnectionId:  connectionId,
			ClientAddress: conn.ClientAddress,
			CreatedByUser: producer.CreatedByUser,
			IsActive:      producer.IsActive,
			IsDeleted:     producer.IsDeleted,
		},
		PoisonedCgs: poisonedCgs,
	}
	c.IndentedJSON(200, msg)
}
