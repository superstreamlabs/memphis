package handlers

import (
	"context"
	"strech-server/config"
	"strech-server/db"
	"strech-server/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var usersCollection *mongo.Collection = db.GetCollection(db.Client, "users")
var factoriesCollection *mongo.Collection = db.GetCollection(db.Client, "factories")
var stationsCollection *mongo.Collection = db.GetCollection(db.Client, "stations")
var configuration = config.GetConfig()

func getUserDetailsFromMiddleware(c *gin.Context) models.User {
	user, _ := c.Get("user")
	return user.(models.User)
}

func isFactoryExist(factoryName string) (bool, models.Factory, error) {
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

func isStationExist(stationName string) (bool, models.Station, error) {
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
