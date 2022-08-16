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
	"errors"
	"memphis-broker/analytics"
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

type StationsHandler struct{ S *Server }

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
func removeStationResources(s *Server, station models.Station) error {
	err := s.RemoveStream(station.Name)
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

func (s *Server) createStationDirect(csr *createStationRequest) error {
	stationName := strings.ToLower(csr.StationName)
	err := validateStationName(stationName)
	if err != nil {
		serv.Warnf(err.Error())
		return err
	}

	exist, _, err := IsStationExist(stationName)
	if err != nil {
		serv.Errorf("CreateStation error: " + err.Error())
		return err
	}

	if exist {
		serv.Warnf("Station with that name already exists")
		return errors.New("memphis: station with that name already exists")
	}

	factoryName := strings.ToLower(csr.FactoryName)
	exist, factory, err := IsFactoryExist(factoryName)
	if err != nil {
		serv.Errorf("Server Error" + err.Error())
		return err
	}
	if !exist { // create this factory
		err := validateFactoryName(factoryName)
		if err != nil {
			serv.Warnf(err.Error())
			return err
		}

		factory = models.Factory{
			ID:            primitive.NewObjectID(),
			Name:          factoryName,
			Description:   "",
			CreatedByUser: csr.Username,
			CreationDate:  time.Now(),
			IsDeleted:     false,
		}
		_, err = factoriesCollection.InsertOne(context.TODO(), factory)
		if err != nil {
			serv.Errorf("CreateStation error: " + err.Error())
			return err
		}
	}

	var retentionType string
	var retentionValue int
	if csr.RetentionType != "" {
		retentionType = strings.ToLower(csr.RetentionType)
		err = validateRetentionType(retentionType)
		if err != nil {
			serv.Warnf(err.Error())
			return err
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
			return err
		}
	} else {
		storageType = "file"
	}

	replicas := csr.Replicas
	if replicas > 0 {
		err = validateReplicas(replicas)
		if err != nil {
			serv.Warnf(err.Error())
			return err
		}
	} else {
		replicas = 1
	}
	newStation := models.Station{
		ID:              primitive.NewObjectID(),
		Name:            stationName,
		FactoryId:       factory.ID,
		CreatedByUser:   csr.Username,
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

	err = s.CreateStream(newStation)
	if err != nil {
		serv.Warnf(err.Error())
		return err
	}

	_, err = stationsCollection.InsertOne(context.TODO(), newStation)
	if err != nil {
		serv.Errorf("CreateStation error: " + err.Error())
		return err
	}
	message := "Station " + stationName + " has been created"
	serv.Noticef(message)

	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		ID:            primitive.NewObjectID(),
		StationName:   stationName,
		Message:       message,
		CreatedByUser: csr.Username,
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
		analytics.IncrementStationsCounter()
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
		serv.Errorf("GetStationById error: " + err.Error())
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

	stationName := strings.ToLower(body.Name)
	err := validateStationName(stationName)
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
			serv.Warnf(err.Error())
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
			serv.Errorf("CreateStation error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
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

	var storageType string
	if body.StorageType != "" {
		storageType = strings.ToLower(body.StorageType)
		err = validateStorageType(storageType)
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

	err = sh.S.CreateStream(newStation)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	_, err = stationsCollection.InsertOne(context.TODO(), newStation)
	if err != nil {
		serv.Errorf("CreateStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	message := "Station " + stationName + " has been created"
	serv.Noticef(message)
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
		serv.Warnf("CreateStation error: " + err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.IncrementStationsCounter()
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

	stationName := strings.ToLower(body.StationName)
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

	serv.Noticef("Station " + stationName + " has been deleted")
	c.IndentedJSON(200, gin.H{})
}

func (s *Server) removeStationDirect(dsr *destroyStationRequest) error {
	stationName := strings.ToLower(dsr.StationName)
	exist, station, err := IsStationExist(stationName)
	if err != nil {
		serv.Errorf("RemoveStation error: " + err.Error())
		return err
	}
	if !exist {
		serv.Warnf("Station does not exist")
		return err
	}

	err = removeStationResources(s, station)
	if err != nil {
		serv.Errorf("RemoveStation error: " + err.Error())
		return err
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
		return err
	}

	serv.Noticef("Station " + stationName + " has been deleted")
	return nil
}

func (sh StationsHandler) GetTotalMessages(station models.Station) (int, error) {
	totalMessages, err := sh.S.GetTotalMessagesInStation(station)
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

	exist, station, err := IsStationExist(poisonMessage.StationName)
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

		cgInfo, err := sh.S.GetCgInfo(poisonMessage.StationName, poisonMessage.PoisonedCgs[i].CgName)
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
		for _, cg := range msg.PoisonedCgs {
			err := sh.S.ResendPoisonMessage("$memphis_dlq_"+msg.StationName+"_"+cg.CgName, []byte(msg.Message.Data))
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
	stationName := strings.ToLower(body.StationName)
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
	producedByHeader := hdr["producedBy"]

	if connectionIdHeader == "" || producedByHeader == "" {
		serv.Errorf("Error while getting notified about a poison message: Missing mandatory message headers, please upgrade the SDk version you are using")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Error while getting notified about a poison message: Missing mandatory message headers, please upgrade the SDk version you are using"})
		return
	}

	connectionId, _ := primitive.ObjectIDFromHex(connectionIdHeader)
	poisonedCgs, err := GetPoisonedCgsByMessage(stationName, models.MessageDetails{MessageSeq: int(sm.Sequence), ProducedBy: producedByHeader, TimeSent: sm.Time})
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

		totalPoisonMsgs, err := GetTotalPoisonMsgsByCg(stationName, cg.CgName)
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
