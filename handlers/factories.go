package handlers

import (
	// "context"
	// "strech-server/logger"
	// "strech-server/models"
	// // "strech-server/utils"
	// // "strings"
	// // "time"

	// "github.com/gin-gonic/gin"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	// "go.mongodb.org/mongo-driver/mongo"
)

// type FactoriesHandler struct{}

// func (umh FactoriesHandler) GetBoxFactories(c *gin.Context) {
// 	var boxes []models.Box

// 	cursor, err := boxesCollection.Find(context.TODO(), bson.M{})
// 	if err != nil {
// 		logger.Error("GetAllBoxes error: " + err.Error())
// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
// 		return
// 	}

// 	if err = cursor.All(context.TODO(), &boxes); err != nil {
// 		logger.Error("GetAllBoxes error: " + err.Error())
// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
// 		return
// 	}

// 	if len(boxes) == 0 {
// 		c.IndentedJSON(200, []string{})
// 	} else {
// 		c.IndentedJSON(200, boxes)
// 	}
// }