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
	"errors"
	"sort"

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
	"k8s.io/utils/strings/slices"
)

type ConsumersHandler struct{ S *Server }

func validateName(name string) error {
	if len(name) == 0 {
		return errors.New("Consumer name can not be empty")
	}

	if len(name) > 32 {
		return errors.New("Consumer name/consumer group should be under 32 characters")
	}

	re := regexp.MustCompile("^[a-z0-9_]*$")

	validName := re.MatchString(name)
	if !validName {
		return errors.New("Consumer name/consumer group has to include only letters and _")
	}
	return nil
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
		bson.D{{"$project", bson.D{{"station", 0}, {"factory", 0}, {"connection", 0}}}},
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

func (s *Server) createConsumerDirect(ccr *createConsumerRequest, c *client) error {
	name := strings.ToLower(ccr.Name)
	err := validateName(name)
	if err != nil {
		serv.Warnf(err.Error())
		return err
	}

	exist, user, err := IsUserExist(c.memphisInfo.username)
	if err != nil {
		serv.Errorf("createConsumerDirect error: " + err.Error())
		return err
	}
	if !exist {
		serv.Warnf("createConsumerDirect error: User does not exist")
		return errors.New("User does not exist")
	}

	consumerGroup := strings.ToLower(ccr.ConsumerGroup)
	if consumerGroup != "" {
		err = validateName(consumerGroup)
		if err != nil {
			serv.Warnf(err.Error())
			return err
		}
	} else {
		consumerGroup = name
	}

	consumerType := strings.ToLower(ccr.ConsumerType)
	err = validateConsumerType(consumerType)
	if err != nil {
		serv.Warnf(err.Error())
		return err
	}

	connectionIdObj, err := primitive.ObjectIDFromHex(ccr.ConnectionId)
	if err != nil {
		serv.Warnf("Connection id is not valid")
		return err
	}
	exist, connection, err := IsConnectionExist(connectionIdObj)
	if err != nil {
		serv.Errorf("CreateConsumer error: " + err.Error())
		return err
	}
	if !exist {
		serv.Warnf("Connection id was not found")
		return errors.New("connection id was not found")
	}
	if !connection.IsActive {
		serv.Warnf("Connection is not active")
		return errors.New("connection is not active")
	}

	stationName := strings.ToLower(ccr.StationName)
	exist, station, err := IsStationExist(stationName)
	if err != nil {
		serv.Errorf("CreateConsumer error: " + err.Error())
		return err
	}
	if !exist {
		station, err = CreateDefaultStation(s, stationName, connection.CreatedByUser)
		if err != nil {
			serv.Errorf("creating default station error: " + err.Error())
			return err
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
			UserType:      "application",
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			serv.Errorf("CreateConsumer error: " + err.Error())
		}

		shouldSendAnalytics, _ := shouldSendAnalytics()
		if shouldSendAnalytics {
			analytics.IncrementStationsCounter()
		}
	}

	exist, _, err = IsConsumerExist(name, station.ID)
	if err != nil {
		serv.Errorf("CreateConsumer error: " + err.Error())
		return err
	}
	if exist {
		serv.Warnf("Consumer name has to be unique per station")
		return errors.New("memphis: consumer name has to be unique per station")
	}

	consumerGroupExist, consumerFromGroup, err := isConsumerGroupExist(consumerGroup, station.ID)
	if err != nil {
		serv.Errorf("CreateConsumer error: " + err.Error())
		return err
	}

	consumerId := primitive.NewObjectID()
	newConsumer := models.Consumer{
		ID:             consumerId,
		Name:           name,
		StationId:      station.ID,
		FactoryId:      station.FactoryId,
		Type:           consumerType,
		ConnectionId:   connectionIdObj,
		CreatedByUser:  connection.CreatedByUser,
		ConsumersGroup: consumerGroup,
		IsActive:       true,
		CreationDate:   time.Now(),
		IsDeleted:      false,
	}

	if consumerGroupExist {
		newConsumer.MaxAckTimeMs = consumerFromGroup.MaxAckTimeMs
		newConsumer.MaxMsgDeliveries = consumerFromGroup.MaxMsgDeliveries
	} else {
		newConsumer.MaxAckTimeMs = int64(ccr.MaxAckTimeMillis)
		newConsumer.MaxMsgDeliveries = ccr.MaxMsgDeliveries
		s.CreateConsumer(newConsumer, station)
		if err != nil {
			serv.Errorf("CreateConsumer error: " + err.Error())
			return err
		}
	}

	_, err = consumersCollection.InsertOne(context.TODO(), newConsumer)
	if err != nil {
		serv.Errorf("CreateConsumer error: " + err.Error())
		return err
	}
	message := "Consumer " + name + " has been created"
	serv.Noticef(message)
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		ID:            primitive.NewObjectID(),
		StationName:   stationName,
		Message:       message,
		CreatedByUser: user.Username,
		CreationDate:  time.Now(),
		UserType:      "application",
	}
	auditLogs = append(auditLogs, newAuditLog)
	err = CreateAuditLogs(auditLogs)
	if err != nil {
		serv.Errorf("CreateConsumer error: " + err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.IncrementConsumersCounter()
	}

	return nil
}

func (ch ConsumersHandler) GetAllConsumers(c *gin.Context) {
	var consumers []models.ExtendedConsumer
	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "factories"}, {"localField", "factory_id"}, {"foreignField", "_id"}, {"as", "factory"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$factory"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"consumers_group", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"max_ack_time_ms", 1}, {"max_msg_deliveries", 1}, {"station_name", "$station.name"}, {"factory_name", "$factory.name"}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"factory", 0}, {"connection", 0}}}},
	})

	if err != nil {
		serv.Errorf("GetAllConsumers error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		serv.Errorf("GetAllConsumers error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(consumers) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, consumers)
	}
}

func (ch ConsumersHandler) GetCgsByStation(station models.Station) ([]models.Cg, []models.Cg, []models.Cg, error) { // for socket io endpoint
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
			cgInfo, err := ch.S.GetCgInfo(station.Name, cg.Name)
			if err != nil {
				return cgs, cgs, cgs, err
			}

			totalPoisonMsgs, err := GetTotalPoisonMsgsByCg(station.Name, cg.Name)
			if err != nil {
				return cgs, cgs, cgs, err
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

	var consumers []models.ExtendedConsumer
	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"station_id", station.ID}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "factories"}, {"localField", "factory_id"}, {"foreignField", "_id"}, {"as", "factory"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$factory"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"consumers_group", 1}, {"creation_date", 1}, {"is_active", 1}, {"is_deleted", 1}, {"max_ack_time_ms", 1}, {"max_msg_deliveries", 1}, {"station_name", "$station.name"}, {"factory_name", "$factory.name"}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"factory", 0}, {"connection", 0}}}},
	})

	if err != nil {
		serv.Errorf("GetAllConsumersByStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		serv.Errorf("GetAllConsumersByStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(consumers) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, consumers)
	}
}

func (s *Server) destroyConsumerDirect(dcr *destroyConsumerRequest, c *client) error {
	stationName := strings.ToLower(dcr.StationName)
	name := strings.ToLower(dcr.ConsumerName)
	_, station, err := IsStationExist(stationName)
	if err != nil {
		serv.Errorf("DestroyConsumer error: " + err.Error())
		return err
	}

	var consumer models.Consumer
	err = consumersCollection.FindOneAndUpdate(context.TODO(),
		bson.M{"name": name, "station_id": station.ID, "is_active": true},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	).Decode(&consumer)
	if err == mongo.ErrNoDocuments {
		serv.Warnf("A consumer with the given details was not found")
		return errors.New("memphis: a consumer with the given details was not found")
	}
	if err != nil {
		serv.Errorf("DestroyConsumer error: " + err.Error())
		return err
	}

	_, err = consumersCollection.UpdateMany(context.TODO(),
		bson.M{"name": name, "station_id": station.ID},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	)
	if err != nil {
		serv.Errorf("DestroyConsumer error: " + err.Error())
		return err
	}

	// ensure not part of an active consumer group
	count, err := consumersCollection.CountDocuments(context.TODO(), bson.M{"station_id": station.ID, "consumers_group": consumer.ConsumersGroup, "is_deleted": false})
	if err != nil {
		serv.Errorf("DestroyConsumer error: " + err.Error())
		return err
	}

	if count == 0 { // no other members in this group
		err = s.RemoveConsumer(stationName, consumer.ConsumersGroup)
		if err != nil {
			serv.Errorf("DestroyConsumer error: " + err.Error())
			return err
		}

		err = RemovePoisonedCg(stationName, consumer.ConsumersGroup)
		if err != nil {
			serv.Errorf("DestroyConsumer error: " + err.Error())
			return err
		}
	}

	exist, user, err := IsUserExist(c.memphisInfo.username)
	if err != nil {
		serv.Errorf("destroyConsumerDirect error: " + err.Error())
		return err
	}
	if !exist {
		serv.Warnf("destroyConsumerDirect error: User does not exist")
		return errors.New("User does not exist")
	}

	message := "Consumer " + name + " has been deleted"
	serv.Noticef(message)
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		ID:            primitive.NewObjectID(),
		StationName:   stationName,
		Message:       message,
		CreatedByUser: user.Username,
		CreationDate:  time.Now(),
		UserType:      "application",
	}
	auditLogs = append(auditLogs, newAuditLog)
	err = CreateAuditLogs(auditLogs)
	if err != nil {
		serv.Errorf("DestroyConsumer error: " + err.Error())
	}

	return nil
}

func (ch ConsumersHandler) KillConsumers(connectionId primitive.ObjectID) error {
	var consumers []models.Consumer
	var station models.Station

	cursor, err := consumersCollection.Find(context.TODO(), bson.M{"connection_id": connectionId, "is_active": true})
	if err != nil {
		serv.Errorf("KillConsumers error: " + err.Error())
	}
	if err = cursor.All(context.TODO(), &consumers); err != nil {
		serv.Errorf("KillConsumers error: " + err.Error())
	}

	if len(consumers) > 0 {
		err = stationsCollection.FindOne(context.TODO(), bson.M{"_id": consumers[0].StationId}).Decode(&station)
		if err != nil {
			serv.Errorf("KillConsumers error: " + err.Error())
		}
		_, err = consumersCollection.UpdateMany(context.TODO(),
			bson.M{"connection_id": connectionId},
			bson.M{"$set": bson.M{"is_active": false}},
		)
		if err != nil {
			serv.Errorf("KillConsumers error: " + err.Error())
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
			message = "Consumer " + consumer.Name + " has been disconnected"
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
			serv.Errorf("KillConsumers error: " + err.Error())
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
		serv.Errorf("ReliveConsumers error: " + err.Error())
		return err
	}

	return nil
}
