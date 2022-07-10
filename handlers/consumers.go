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
	"k8s.io/utils/strings/slices"
)

type ConsumersHandler struct{}

func validateName(name string) error {
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
	filter := bson.M{"consumers_group": consumerGroup, "station_id": stationId, "is_active": true}
	var consumer models.Consumer
	err := consumersCollection.FindOne(context.TODO(), filter).Decode(&consumer)
	if err == mongo.ErrNoDocuments {
		return false, models.Consumer{}, nil
	} else if err != nil {
		return false, models.Consumer{}, err
	}
	return true, consumer, nil
}

func GetConsumerGroupMembers(cgName string, stationName string) ([]models.CgMember, error) {
	var consumers []models.CgMember
	exist, station, err := IsStationExist(stationName)
	if err != nil {
		return consumers, err
	}
	if !exist {
		return consumers, errors.New("Station does not exist")
	}

	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"consumers_group", cgName}, {"station_id", station.ID}}}},
		bson.D{{"$lookup", bson.D{{"from", "connections"}, {"localField", "connection_id"}, {"foreignField", "_id"}, {"as", "connection"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$connection"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"name", 1}, {"created_by_user", 1}, {"is_active", 1}, {"is_deleted", 1}, {"max_ack_time_ms", 1}, {"max_msg_deliveries", 1}, {"client_address", "$connection.client_address"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"factory", 0}, {"connection", 0}}}},
	})

	if err != nil {
		logger.Error("GetConsumerGroupMembers error: " + err.Error())
		return consumers, err
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		logger.Error("GetConsumerGroupMembers error: " + err.Error())
		return consumers, err
	}

	return consumers, nil
}

func (ch ConsumersHandler) CreateConsumer(c *gin.Context) {
	var body models.CreateConsumerSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	name := strings.ToLower(body.Name)
	err := validateName(name)
	if err != nil {
		logger.Warn(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	consumerGroup := strings.ToLower(body.ConsumersGroup)
	if consumerGroup != "" {
		err = validateName(consumerGroup)
		if err != nil {
			logger.Warn(err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	} else {
		consumerGroup = name
	}

	consumerType := strings.ToLower(body.ConsumerType)
	err = validateConsumerType(consumerType)
	if err != nil {
		logger.Warn(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	connectionId, err := primitive.ObjectIDFromHex(body.ConnectionId)
	if err != nil {
		logger.Warn("Connection id is not valid")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Connection id is not valid"})
		return
	}
	exist, connection, err := IsConnectionExist(connectionId)
	if err != nil {
		logger.Error("CreateConsumer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		logger.Warn("Connection id was not found")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Connection id was not found"})
		return
	}
	if !connection.IsActive {
		logger.Warn("Connection is not active")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Connection is not active"})
		return
	}

	stationName := strings.ToLower(body.StationName)
	exist, station, err := IsStationExist(stationName)
	if err != nil {
		logger.Error("CreateConsumer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		station, err = CreateDefaultStation(stationName, connection.CreatedByUser)
		if err != nil {
			logger.Error("CreateConsumer error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		user := getUserDetailsFromMiddleware(c)
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
			logger.Warn("CreateConsumer error: " + err.Error())
		}

		shouldSendAnalytics, _ := shouldSendAnalytics()
		if shouldSendAnalytics {
			analytics.IncrementStationsCounter()
		}
	}

	exist, _, err = IsConsumerExist(name, station.ID)
	if err != nil {
		logger.Error("CreateConsumer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		logger.Warn("Consumer name has to be unique in a station level")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Consumer name has to be unique in a station level"})
		return
	}

	consumerGroupExist, consumerFromGroup, err := isConsumerGroupExist(consumerGroup, station.ID)
	if err != nil {
		logger.Error("CreateConsumer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	consumerId := primitive.NewObjectID()
	newConsumer := models.Consumer{
		ID:             consumerId,
		Name:           name,
		StationId:      station.ID,
		FactoryId:      station.FactoryId,
		Type:           consumerType,
		ConnectionId:   connectionId,
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
		newConsumer.MaxAckTimeMs = body.MaxAckTimeMs
		newConsumer.MaxMsgDeliveries = body.MaxMsgDeliveries
		broker.CreateConsumer(newConsumer, station)
		if err != nil {
			logger.Error("CreateConsumer error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	_, err = consumersCollection.InsertOne(context.TODO(), newConsumer)
	if err != nil {
		logger.Error("CreateConsumer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	user := getUserDetailsFromMiddleware(c)
	message := "Consumer " + name + " has been created"
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
		logger.Warn("CreateConsumer error: " + err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.IncrementConsumersCounter()
	}

	c.IndentedJSON(200, gin.H{
		"consumer_id": consumerId,
	})
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
		logger.Error("GetAllConsumers error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		logger.Error("GetAllConsumers error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(consumers) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, consumers)
	}
}

func (ch ConsumersHandler) GetConsumersByStation(station models.Station) ([]models.ExtendedConsumer, []models.ExtendedConsumer, []models.ExtendedConsumer, error) { // for socket io endpoint
	var consumers []models.ExtendedConsumer

	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"station_id", station.ID}}}},
		bson.D{{"$sort", bson.D{{"creation_date", -1}}}},
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
		return consumers, consumers, consumers, err
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		return consumers, consumers, consumers, err
	}

	if len(consumers) == 0 {
		consumers = []models.ExtendedConsumer{}
	}

	var activeConsumers []models.ExtendedConsumer
	var killedConsumers []models.ExtendedConsumer
	var destroyedConsumers []models.ExtendedConsumer
	consumersNames := []string{}

	for _, consumer := range consumers {
		if slices.Contains(consumersNames, consumer.Name) {
			continue
		}

		consumersNames = append(consumersNames, consumer.Name)
		if consumer.IsActive {
			activeConsumers = append(activeConsumers, consumer)
		} else if !consumer.IsDeleted && !consumer.IsActive {
			killedConsumers = append(killedConsumers, consumer)
		} else if consumer.IsDeleted {
			destroyedConsumers = append(destroyedConsumers, consumer)
		}
	}

	if len(activeConsumers) == 0 {
		activeConsumers = []models.ExtendedConsumer{}
	}

	if len(killedConsumers) == 0 {
		killedConsumers = []models.ExtendedConsumer{}
	}

	if len(destroyedConsumers) == 0 {
		destroyedConsumers = []models.ExtendedConsumer{}
	}

	return activeConsumers, killedConsumers, destroyedConsumers, nil
}

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
		logger.Warn("Station does not exist")
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
		logger.Error("GetAllConsumersByStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		logger.Error("GetAllConsumersByStation error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(consumers) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, consumers)
	}
}

func (ch ConsumersHandler) DestroyConsumer(c *gin.Context) {
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}

	var body models.DestroyConsumerSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName := strings.ToLower(body.StationName)
	name := strings.ToLower(body.Name)
	_, station, err := IsStationExist(stationName)
	if err != nil {
		logger.Error("DestroyConsumer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	var consumer models.Consumer
	err = consumersCollection.FindOneAndUpdate(context.TODO(),
		bson.M{"name": name, "station_id": station.ID, "is_active": true},
		bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
	).Decode(&consumer)
	if err == mongo.ErrNoDocuments {
		logger.Warn("A consumer with the given details was not found")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "A consumer with the given details was not found"})
		return
	}
	if err != nil {
		logger.Error("DestroyConsumer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if consumer.ConsumersGroup == "" {
		err = broker.RemoveConsumer(stationName, name)
		if err != nil {
			logger.Error("DestroyConsumer error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	} else { // ensure not part of an active consumer group
		var consumers []models.Consumer

		cursor, err := consumersCollection.Find(context.TODO(), bson.M{"consumers_group": consumer.ConsumersGroup, "is_active": true})
		if err != nil {
			logger.Error("DestroyConsumer error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		if err = cursor.All(context.TODO(), &consumers); err != nil {
			logger.Error("DestroyConsumer error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		if len(consumers) == 0 { // no other active members in this group
			err = broker.RemoveConsumer(stationName, consumer.ConsumersGroup)
			if err != nil {
				logger.Error("DestroyConsumer error: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
		}
	}
	user := getUserDetailsFromMiddleware(c)
	message := "Consumer " + name + " has been deleted"
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
		logger.Warn("DestroyConsumer error: " + err.Error())
	}
	c.IndentedJSON(200, gin.H{})
}

func (ch ConsumersHandler) KillConsumers(connectionId primitive.ObjectID) error {
	var consumers []models.Consumer
	var station models.Station

	cursor, err := consumersCollection.Find(context.TODO(), bson.M{"connection_id": connectionId, "is_active": true})
	if err != nil {
		logger.Warn("KillConsumers error: " + err.Error())
	}
	if err = cursor.All(context.TODO(), &consumers); err != nil {
		logger.Warn("KillConsumers error: " + err.Error())
	}

	if len(consumers) > 0 {
		err = stationsCollection.FindOne(context.TODO(), bson.M{"_id": consumers[0].StationId}).Decode(&station)
		if err != nil {
			logger.Warn("KillConsumers error: " + err.Error())
		}
		_, err = consumersCollection.UpdateMany(context.TODO(),
			bson.M{"connection_id": connectionId},
			bson.M{"$set": bson.M{"is_active": false}},
		)
		if err != nil {
			logger.Error("KillConsumers error: " + err.Error())
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
			logger.Warn("KillConsumers error: " + err.Error())
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
		logger.Error("ReliveConsumers error: " + err.Error())
		return err
	}

	return nil
}
