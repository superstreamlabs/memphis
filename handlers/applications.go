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

type ApplicationsHandler struct{}

func (umh ApplicationsHandler) CreateApplication(c *gin.Context) {
	var body models.CreateApplicationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	applicationName := strings.ToLower(body.Name)
	exist, err := isApplicationExist(applicationName)
	if err != nil {
		logger.Error("CreateApplication error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		c.AbortWithStatusJSON(400, gin.H{"message": "An application with that name is already exist"})
		return
	}

	user := getUserDetailsFromMiddleware(c)
	newApplication := models.Application{
		ID:            primitive.NewObjectID(),
		Name:          applicationName,
		Description:   strings.ToLower(body.Description),
		CreatedByUSer: user.Username,
		CreationDate:  time.Now(),
	}

	_, err = applicationsCollection.InsertOne(context.TODO(), newApplication)
	if err != nil {
		logger.Error("CreateApplication error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{
		"id":              newApplication.ID,
		"name":            newApplication.Name,
		"description":     newApplication.Description,
		"created_by_user": newApplication.CreatedByUSer,
		"creation_date":   newApplication.CreationDate,
	})
}

func (umh ApplicationsHandler) GetAllApplications(c *gin.Context) {
	var applications []models.Application

	cursor, err := applicationsCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		logger.Error("GetAllApplications error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &applications); err != nil {
		logger.Error("GetAllApplications error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(applications) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, applications)
	}
}

func (umh ApplicationsHandler) RemoveApplication(c *gin.Context) {
	var body models.RemoveApplicationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	applicationName := strings.ToLower(body.ApplicationName)
	exist, err := isApplicationExist(applicationName)
	if err != nil {
		logger.Error("RemoveApplication error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.AbortWithStatusJSON(400, gin.H{"message": "An application with that name is not exist"})
		return
	}

	_, err = applicationsCollection.DeleteOne(context.TODO(), bson.M{"name": applicationName})
	if err != nil {
		logger.Error("RemoveApplication error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh ApplicationsHandler) EditApplication(c *gin.Context) {
	var body models.EditApplicationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	applicationName := strings.ToLower(body.ApplicationName)
	exist, err := isApplicationExist(applicationName)
	if err != nil {
		logger.Error("EditApplication error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.AbortWithStatusJSON(400, gin.H{"message": "An application with that name is not exist"})
		return
	}

	var application models.Application
	err = applicationsCollection.FindOne(context.TODO(), bson.M{"name": applicationName}).Decode(&application)
	if err != nil {
		logger.Error("EditApplication error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if body.NewName != "" {
		application.Name = strings.ToLower(body.NewName)
	}

	if body.NewDescription != "" {
		application.Description = strings.ToLower(body.NewDescription)
	}

	_, err = applicationsCollection.UpdateOne(context.TODO(),
		bson.M{"name": applicationName},
		bson.M{"$set": bson.M{"name": application.Name, "description": application.Description}},
	)
	if err != nil {
		logger.Error("EditApplication error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, application)
}
