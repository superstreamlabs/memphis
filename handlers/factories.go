package handlers

import (
	// "context"
	// "strech-server/logger"
	"strech-server/models"
	"strech-server/utils"
	// // "strings"
	// // "time"

	"github.com/gin-gonic/gin"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	// "go.mongodb.org/mongo-driver/mongo"
)

type FactoriesHandler struct{}

func (umh FactoriesHandler) GetFactoryById(c *gin.Context) {
	var body models.GetFactoryByIdSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	// user := getUserDetailsFromMiddleware(c)
}

func (umh FactoriesHandler) GetApplicationFactories(c *gin.Context) {
	var body models.GetApplicationFactoriesSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
}

func (umh FactoriesHandler) CreateFactory(c *gin.Context) {
	var body models.CreateFactorySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
}

func (umh FactoriesHandler) RemoveFactory(c *gin.Context) {
	var body models.RemoveFactorySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
}
