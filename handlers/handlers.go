package handlers

import (
	"strech-server/config"
	"strech-server/db"
	"strech-server/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var usersCollection *mongo.Collection = db.GetCollection(db.Client, "users")
var configuration = config.GetConfig()
var boxesCollection *mongo.Collection = db.GetCollection(db.Client, "boxes")

func getUserDetailsFromMiddleware(c *gin.Context) models.User {
	user, _ := c.Get("user")
	return user.(models.User)
}