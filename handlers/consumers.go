package handlers

import (
	"context"
	"errors"
	"memphis-control-plane/logger"
	"memphis-control-plane/models"
	"memphis-control-plane/utils"
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

// TODO
func validateConsumerName(name string) error {
	if name == "" {
		return errors.New("Consumer name is not valid")
	}

	return nil
}

func validateConsumerType(consumerType string) error {
	if consumerType != "application" && consumerType != "connector" {
		return errors.New("Consumer type has to be one of the following application/connector")
	}
	return nil
}

func (umh ConsumersHandler) CreateConsumer(c *gin.Context) {
	var body models.CreateConsumerSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	name := strings.ToLower(body.Name)
	err := validateConsumerName(name)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"message": err.Error()})
		return
	}

	consumerType := strings.ToLower(body.ConsumerType)
	err = validateConsumerType(consumerType)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"message": err.Error()})
		return
	}

	connectionId, err := primitive.ObjectIDFromHex(body.ConnectionId)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"message": "Connection id is not valid"})
		return
	}
	exist, connection, err := IsConnectionExist(connectionId)
	if err != nil {
		logger.Error("CreateConsumer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.AbortWithStatusJSON(400, gin.H{"message": "Connection id was not found"})
		return
	}
	if !connection.IsActive {
		c.AbortWithStatusJSON(400, gin.H{"message": "Connection is not active"})
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
		c.AbortWithStatusJSON(400, gin.H{"message": "Station was not found"})
		return
	}

	consumerId := primitive.NewObjectID()
	newConsumer := models.Consumer{
		ID:            consumerId,
		Name:          name,
		StationId:     station.ID,
		FactoryId:     station.FactoryId,
		Type:          consumerType,
		ConnectionId:  connectionId,
		CreatedByUser: connection.CreatedByUser,
		IsActive:      true,
		CreationDate:  time.Now(),
	}

	_, err = consumersCollection.InsertOne(context.TODO(), newConsumer)
	if err != nil {
		logger.Error("CreateConsumer error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

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
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"connection_id", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"station_name", "$station.name"}, {"factory_name", "$factory.name"}}}},
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
		c.AbortWithStatusJSON(400, gin.H{"message": "Station does not exist"})
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

func (umh ConsumersHandler) RemoveConsumer(consumerId primitive.ObjectID) error {
	_, err := consumersCollection.UpdateOne(context.TODO(),
		bson.M{"_id": consumerId},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		logger.Error("RemoveConsumer error: " + err.Error())
		return err
	}

	return nil
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
