package handlers

import (
	"context"
	"strech-server/logger"
	"strech-server/models"
	"strech-server/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BoxesHandler struct{}

func isBoxExist(boxName string) (bool, error) {
	filter := bson.M{"name": boxName}
	var box models.Box
	err := boxesCollection.FindOne(context.TODO(), filter).Decode(&box)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (umh BoxesHandler) CreateBox(c *gin.Context) {
	var body models.CreateBoxSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	boxname := strings.ToLower(body.Name)
	exist, err := isBoxExist(boxname)
	if err != nil {
		logger.Error("CreateBox error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		c.AbortWithStatusJSON(400, gin.H{"message": "A box with that name is already exist"})
		return
	}

	user := getUserDetailsFromMiddleware(c)
	newBox := models.Box{
		ID:            primitive.NewObjectID(),
		Name:          boxname,
		Description:   body.Description,
		CreatedByUSer: user.Username,
		CreationDate:  time.Now(),
	}

	_, err = boxesCollection.InsertOne(context.TODO(), newBox)
	if err != nil {
		logger.Error("CreateBox error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{
		"id":              newBox.ID,
		"name":            newBox.Name,
		"description":     newBox.Description,
		"created_by_user": newBox.CreatedByUSer,
		"creation_date":   newBox.CreationDate,
	})
}

func (umh BoxesHandler) GetAllBoxes(c *gin.Context) {
	var boxes []models.Box

	cursor, err := boxesCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		logger.Error("GetAllBoxes error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &boxes); err != nil {
		logger.Error("GetAllBoxes error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(boxes) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, boxes)
	}
}

func (umh BoxesHandler) RemoveBox(c *gin.Context) {
	var body models.RemoveBoxSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	_, err := boxesCollection.DeleteOne(context.TODO(), bson.M{"_id": body.BoxId})
	if err != nil {
		logger.Error("RemoveBox error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh BoxesHandler) EditBox(c *gin.Context) {
	var body models.EditBoxSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	boxname := strings.ToLower(body.Name)
	_, err := boxesCollection.UpdateOne(context.TODO(),
		bson.M{"_id": body.BoxId},
		bson.M{"$set": bson.M{"name": boxname, "description": body.Description}},
	)
	if err != nil {
		logger.Error("EditBox error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	var box models.Box
	err = boxesCollection.FindOne(context.TODO(), bson.M{"_id": body.BoxId}).Decode(&box)
	if err != nil {
		logger.Error("EditBox error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, box)
}
