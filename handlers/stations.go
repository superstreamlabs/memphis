package handlers

import (
	"context"
	"strech-server/logger"
	"strech-server/models"
	"strech-server/utils"

	// // "strings"
	// // "time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type StationsHandler struct{}

func (umh StationsHandler) GetStationById(c *gin.Context) {
	var body models.GetStationByIdSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	var station models.Station
	err := stationsCollection.FindOne(context.TODO(), bson.M{"_id": body.StationId}).Decode(&station)
	if err == mongo.ErrNoDocuments {
		c.AbortWithStatusJSON(404, gin.H{"message": "Station does not exist"})
		return
	} else if err != nil {
		logger.Error("GetStationById error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, station)
}

func (umh StationsHandler) GetFactoryStations(c *gin.Context) {
	var body models.GetFactoryStationsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
}

func (umh StationsHandler) CreateStation(c *gin.Context) {
	var body models.CreateStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
}

func (umh StationsHandler) RemoveStation(c *gin.Context) {
	var body models.RemoveStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
}
