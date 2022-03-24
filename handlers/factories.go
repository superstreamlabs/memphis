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
)

type FactoriesHandler struct{}

func (umh FactoriesHandler) CreateFactory(c *gin.Context) {
	var body models.CreateFactorySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	factoryName := strings.ToLower(body.Name)
	exist, err := isFactoryExist(factoryName)
	if err != nil {
		logger.Error("CreateFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		c.AbortWithStatusJSON(400, gin.H{"message": "Factory with that name is already exist"})
		return
	}

	user := getUserDetailsFromMiddleware(c)
	newFactory := models.Factory{
		ID:            primitive.NewObjectID(),
		Name:          factoryName,
		Description:   strings.ToLower(body.Description),
		CreatedByUSer: user.Username,
		CreationDate:  time.Now(),
	}

	_, err = factoriesCollection.InsertOne(context.TODO(), newFactory)
	if err != nil {
		logger.Error("CreateFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{
		"id":              newFactory.ID,
		"name":            newFactory.Name,
		"description":     newFactory.Description,
		"created_by_user": newFactory.CreatedByUSer,
		"creation_date":   newFactory.CreationDate,
	})
}

func (umh FactoriesHandler) GetAllFactories(c *gin.Context) {
	var factories []models.Factory

	cursor, err := factoriesCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		logger.Error("GetAllFactories error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &factories); err != nil {
		logger.Error("GetAllFactories error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(factories) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, factories)
	}
}

func (umh FactoriesHandler) RemoveFactory(c *gin.Context) {
	var body models.RemoveFactorySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	factoryName := strings.ToLower(body.FactoryName)
	exist, err := isFactoryExist(factoryName)
	if err != nil {
		logger.Error("RemoveFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.AbortWithStatusJSON(400, gin.H{"message": "Factory with that name does exist"})
		return
	}

	_, err = factoriesCollection.DeleteOne(context.TODO(), bson.M{"name": factoryName})
	if err != nil {
		logger.Error("RemoveFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh FactoriesHandler) EditFactory(c *gin.Context) {
	var body models.EditFactorySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	factoryName := strings.ToLower(body.FactoryName)
	exist, err := isFactoryExist(factoryName)
	if err != nil {
		logger.Error("EditFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.AbortWithStatusJSON(400, gin.H{"message": "Factory with that name does not exist"})
		return
	}

	var factory models.Factory
	err = factoriesCollection.FindOne(context.TODO(), bson.M{"name": factoryName}).Decode(&factory)
	if err != nil {
		logger.Error("EditFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if body.NewName != "" {
		factory.Name = strings.ToLower(body.NewName)
	}

	if body.NewDescription != "" {
		factory.Description = strings.ToLower(body.NewDescription)
	}

	_, err = factoriesCollection.UpdateOne(context.TODO(),
		bson.M{"name": factoryName},
		bson.M{"$set": bson.M{"name": factory.Name, "description": factory.Description}},
	)
	if err != nil {
		logger.Error("EditFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, factory)
}
