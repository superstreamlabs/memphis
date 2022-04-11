package handlers

import (
	"context"
	"errors"
	"memphis-control-plane/logger"
	"memphis-control-plane/models"
	"memphis-control-plane/utils"
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

// TODO
func validateProducerName(name string) error {
	if name == "" {
		return errors.New("Producer name is not valid")
	}

	return nil
}

func validateProducerType(producerType string) error {
	if producerType != "application" && producerType != "connector" {
		return errors.New("Consumer type has to be one of the following application/connector")
	}
	return nil
}

func (umh ProducersHandler) CreateProducer(name string, stationName string, connectionId primitive.ObjectID, producerType string, username string) error {
	err := validateProducerName(name)
	if err != nil {
		return err
	}

	err = validateProducerType(producerType)
	if err != nil {
		return err
	}

	exist, _, err := IsConnectionExist(connectionId)
	if err != nil {
		logger.Error("CreateProducer error: " + err.Error())
		return err
	}
	if !exist {
		return errors.New("Connection id was not found")
	}

	exist, _, err = IsUserExist(username)
	if err != nil {
		logger.Error("CreateProducer error: " + err.Error())
		return err
	}
	if !exist {
		return errors.New("User was not found")
	}

	exist, station, err := IsStationExist(stationName)
	if err != nil {
		logger.Error("CreateProducer error: " + err.Error())
		return err
	}
	if !exist {
		return errors.New("Station was not found")
	}

	producerId := primitive.NewObjectID()
	newProducer := models.Producer{
		ID:            producerId,
		Name:          name,
		StationId:     station.ID,
		FactoryId:     station.FactoryId,
		Type:          producerType,
		ConnectionId:  connectionId,
		CreatedByUser: username,
		IsActive:      true,
		CreationDate:  time.Now(),
	}

	_, err = producersCollection.InsertOne(context.TODO(), newProducer)
	if err != nil {
		logger.Error("CreateProducer error: " + err.Error())
		return err
	}
	return nil
}

func (umh ProducersHandler) GetAllProducers(c *gin.Context) {
	var producers []extendedProducer
	cursor, err := producersCollection.Aggregate(context.TODO(), mongo.Pipeline{
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
		c.AbortWithStatusJSON(400, gin.H{"message": "Station does not exist"})
		return
	}

	var producers []extendedProducer
	cursor, err := producersCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"station_id", station.ID}}}},
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

func (umh ProducersHandler) RemoveProducer(producerId primitive.ObjectID) error {
	_, err := producersCollection.UpdateOne(context.TODO(),
		bson.M{"_id": producerId},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		logger.Error("RemoveProducer error: " + err.Error())
		return err
	}

	return nil
}

func (umh ProducersHandler) RemoveProducers(connectionId primitive.ObjectID) error {
	_, err := producersCollection.UpdateOne(context.TODO(),
		bson.M{"connection_id": connectionId},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		logger.Error("RemoveProducers error: " + err.Error())
		return err
	}

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
