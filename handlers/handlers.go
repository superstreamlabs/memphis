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
var applicationsCollection *mongo.Collection = db.GetCollection(db.Client, "applications")
var factoriesCollection *mongo.Collection = db.GetCollection(db.Client, "factories")
var configuration = config.GetConfig()

func getUserDetailsFromMiddleware(c *gin.Context) models.User {
	user, _ := c.Get("user")
	return user.(models.User)
}

func isApplicationExist(applicationName string) (bool, error) {
	filter := bson.M{"name": applicationName}
	var application models.Application
	err := applicationsCollection.FindOne(context.TODO(), filter).Decode(&application)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}