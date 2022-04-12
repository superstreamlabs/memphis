package handlers

import (
	"context"
	"memphis-control-plane/config"
	"memphis-control-plane/db"
	"memphis-control-plane/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var usersCollection *mongo.Collection = db.GetCollection("users")
var tokensCollection *mongo.Collection = db.GetCollection("tokens")
var imagesCollection *mongo.Collection = db.GetCollection("images")
var factoriesCollection *mongo.Collection = db.GetCollection("factories")
var stationsCollection *mongo.Collection = db.GetCollection("stations")
var connectionsCollection *mongo.Collection = db.GetCollection("connections")
var producersCollection *mongo.Collection = db.GetCollection("producers")
var consumersCollection *mongo.Collection = db.GetCollection("consumers")
var configuration = config.GetConfig()

func getUserDetailsFromMiddleware(c *gin.Context) models.User {
	user, _ := c.Get("user")
	return user.(models.User)
}

func IsUserExist(username string) (bool, models.User, error) {
	filter := bson.M{"username": username}
	var user models.User
	err := usersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return false, user, nil
	} else if err != nil {
		return false, user, err
	}
	return true, user, nil
}

func IsFactoryExist(factoryName string) (bool, models.Factory, error) {
	filter := bson.M{"name": factoryName}
	var factory models.Factory
	err := factoriesCollection.FindOne(context.TODO(), filter).Decode(&factory)
	if err == mongo.ErrNoDocuments {
		return false, factory, nil
	} else if err != nil {
		return false, factory, err
	}
	return true, factory, nil
}

func IsStationExist(stationName string) (bool, models.Station, error) {
	filter := bson.M{"name": stationName}
	var station models.Station
	err := stationsCollection.FindOne(context.TODO(), filter).Decode(&station)
	if err == mongo.ErrNoDocuments {
		return false, station, nil
	} else if err != nil {
		return false, station, err
	}
	return true, station, nil
}

func IsConnectionExist(connectionId primitive.ObjectID) (bool, models.Connection, error) {
	filter := bson.M{"_id": connectionId}
	var connection models.Connection
	err := connectionsCollection.FindOne(context.TODO(), filter).Decode(&connection)
	if err == mongo.ErrNoDocuments {
		return false, connection, nil
	} else if err != nil {
		return false, connection, err
	}
	return true, connection, nil
}
