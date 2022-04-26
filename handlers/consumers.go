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

type ConsumersHandler struct{}

type extendedConsumer struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	Type          string             `json:"type" bson:"type"`
	ConnectionId  primitive.ObjectID `json:"connection_id" bson:"connection_id"`
	CreatedByUser string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
	StationName   string             `json:"station_name" bson:"station_name"`
	FactoryName   string             `json:"factory_name" bson:"factory_name"`
}

func validateName(name string) error {
	re := regexp.MustCompile("^[a-z_]*$")

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

func isConsumerGroupExist(consumerGroup string, stationId primitive.ObjectID) (bool, error) {
	filter := bson.M{"consumers_group": consumerGroup, "station_id": stationId, "is_active": true}
	var consumer models.Consumer
	err := consumersCollection.FindOne(context.TODO(), filter).Decode(&consumer)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func isConsumersWithoutGroupAndNameLikeGroupExist(consumerGroup string, stationId primitive.ObjectID) (bool, error) {
	filter := bson.M{"name": consumerGroup, "consumers_group": "", "station_id": stationId, "is_active": true}
	var consumer models.Consumer
	err := consumersCollection.FindOne(context.TODO(), filter).Decode(&consumer)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (umh ConsumersHandler) CreateConsumer(c *gin.Context) {
	var body models.CreateConsumerSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	name := strings.ToLower(body.Name)
	err := validateName(name)
	if err != nil {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	consumerGroup :=  strings.ToLower(body.ConsumersGroup)
	if consumerGroup != "" {
		err = validateName(consumerGroup)
		if err != nil {
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}

	consumerType := strings.ToLower(body.ConsumerType)
	err = validateConsumerType(consumerType)
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
		logger.Error("CreateConsumer error: " + err.Error())
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
		logger.Error("CreateConsumer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station was not found"})
		return
	}

	exist, _, err = IsConsumerExist(name, station.ID)
	if err != nil {
		logger.Error("CreateConsumer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Consumer name has to be unique in a station level"})
		return
	}

	var consumerGroupExist bool
	if consumerGroup != "" {
		exist, err = isConsumersWithoutGroupAndNameLikeGroupExist(consumerGroup, station.ID)
		if err != nil {
			logger.Error("CreateConsumer error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if exist {
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can not give your consumer group the same name like another active consumer on the same station"})
			return
		}

		consumerGroupExist, err = isConsumerGroupExist(consumerGroup, station.ID)
		if err != nil {
			logger.Error("CreateConsumer error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	} else {
		exist, err = isConsumerGroupExist(name, station.ID)
		if err != nil {
			logger.Error("CreateConsumer error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if exist {
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can not give your consumer the same name like another active consumer group name on the same station"})
			return
		}
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
	}

	if consumerGroup == "" || !consumerGroupExist {
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

	logger.Info("Consumer " + name + " has been created")
	c.IndentedJSON(200, gin.H{
		"consumer_id": consumerId,
	})
}

func (umh ConsumersHandler) GetAllConsumers(c *gin.Context) {
	var consumers []extendedConsumer
	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"is_active", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "factories"}, {"localField", "factory_id"}, {"foreignField", "_id"}, {"as", "factory"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$factory"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"consumer_group", 1}, {"creation_date", 1}, {"station_name", "$station.name"}, {"factory_name", "$factory.name"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"factory", 0}}}},
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

func (umh ConsumersHandler) GetAllConsumersByStation(c *gin.Context) {
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
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station does not exist"})
		return
	}

	var consumers []extendedConsumer
	cursor, err := consumersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"station_id", station.ID}, {"is_active", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "stations"}, {"localField", "station_id"}, {"foreignField", "_id"}, {"as", "station"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$station"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$lookup", bson.D{{"from", "factories"}, {"localField", "factory_id"}, {"foreignField", "_id"}, {"as", "factory"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$factory"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"station_name", "$station.name"}, {"factory_name", "$factory.name"}}}},
		bson.D{{"$project", bson.D{{"station", 0}, {"factory", 0}}}},
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

func (umh ConsumersHandler) DestroyConsumer(c *gin.Context) {
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
	err = consumersCollection.FindOneAndDelete(context.TODO(), bson.M{"name": name, "station_id": station.ID}).Decode(&consumer)
	if err != nil {
		logger.Error("DestroyConsumer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if err == mongo.ErrNoDocuments {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "A consumer with the given details was not found"})
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

		cursor, err := consumersCollection.Find(context.TODO(), bson.M{"consumer_group": consumer.ConsumersGroup, "is_active": true})
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

	logger.Info("Consumer " + name + " has been deleted")
	c.IndentedJSON(200, gin.H{})
}

func (umh ConsumersHandler) GetAllConsumersByConnection(connectionId primitive.ObjectID) ([]models.Consumer, error) {
	var consumers []models.Consumer

	cursor, err := consumersCollection.Find(context.TODO(), bson.M{"connection_id": connectionId})
	if err != nil {
		logger.Error("GetAllConsumersByConnection error: " + err.Error())
		return consumers, err
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		logger.Error("GetAllConsumersByConnection error: " + err.Error())
		return consumers, err
	}

	return consumers, nil
}

func (umh ConsumersHandler) RemoveConsumers(connectionId primitive.ObjectID) error {
	_, err := consumersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": connectionId},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		logger.Error("RemoveConsumers error: " + err.Error())
		return err
	}

	var consumers []models.Consumer
	cursor, err := consumersCollection.Find(context.TODO(), bson.M{"connection_id": connectionId})
	if err != nil {
		logger.Error("RemoveConsumers error: " + err.Error())
		return err
	}

	if err = cursor.All(context.TODO(), &consumers); err != nil {
		logger.Error("RemoveConsumers error: " + err.Error())
		return err
	}

	return nil
}

func (umh ConsumersHandler) ReliveConsumers(connectionId primitive.ObjectID) error {
	_, err := consumersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": connectionId},
		bson.M{"$set": bson.M{"is_active": true}},
	)
	if err != nil {
		logger.Error("ReliveConsumers error: " + err.Error())
		return err
	}

	return nil
}
