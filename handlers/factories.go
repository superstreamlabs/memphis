package handlers

import (
	"context"
	"errors"
	"memphis-control-plane/broker"
	"memphis-control-plane/logger"
	"memphis-control-plane/models"
	"memphis-control-plane/utils"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FactoriesHandler struct{}

func validateFactoryName(factoryName string) error {
	re := regexp.MustCompile("^[a-z0-9_]*$")

	validName := re.MatchString(factoryName)
	if !validName {
		return errors.New("factory name has to include only letters, numbers and _")
	}
	return nil
}

// TODO remove the stations resources - functions, connectors
func removeStations(factoryId primitive.ObjectID) error {
	var stations []models.Station
	cursor, err := stationsCollection.Find(context.TODO(), bson.M{"factory_id": factoryId})
	if err != nil {
		return err
	}

	if err = cursor.All(context.TODO(), &stations); err != nil {
		return err
	}

	for _, station := range stations {
		err = broker.RemoveStream(station.Name)
		if err != nil {
			return err
		}

		_, err = producersCollection.DeleteMany(context.TODO(), bson.M{"station_id": station.ID})
		if err != nil {
			return err
		}

		_, err = consumersCollection.DeleteMany(context.TODO(), bson.M{"station_id": station.ID})
		if err != nil {
			return err
		}
	}

	_, err = stationsCollection.DeleteMany(context.TODO(), bson.M{"factory_id": factoryId})
	if err != nil {
		return err
	}

	return nil
}

func (umh FactoriesHandler) CreateFactory(c *gin.Context) {
	var body models.CreateFactorySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	factoryName := strings.ToLower(body.Name)
	err := validateFactoryName(factoryName)
	if err != nil {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, _, err := IsFactoryExist(factoryName)
	if err != nil {
		logger.Error("CreateFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Factory with that name is already exist"})
		return
	}

	user := getUserDetailsFromMiddleware(c)
	newFactory := models.Factory{
		ID:            primitive.NewObjectID(),
		Name:          factoryName,
		Description:   strings.ToLower(body.Description),
		CreatedByUser: user.Username,
		CreationDate:  time.Now(),
	}

	_, err = factoriesCollection.InsertOne(context.TODO(), newFactory)
	if err != nil {
		logger.Error("CreateFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	logger.Info("Factory " + factoryName + " has been created")
	c.IndentedJSON(200, gin.H{
		"id":              newFactory.ID,
		"name":            newFactory.Name,
		"description":     newFactory.Description,
		"created_by_user": newFactory.CreatedByUser,
		"creation_date":   newFactory.CreationDate,
	})
}

func (umh FactoriesHandler) GetFactory(c *gin.Context) {
	var body models.GetFactorySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	var factory models.Factory
	err := factoriesCollection.FindOne(context.TODO(), bson.M{"name": body.FactoryName}).Decode(&factory)
	if err == mongo.ErrNoDocuments {
		c.AbortWithStatusJSON(404, gin.H{"message": "Factory does not exist"})
		return
	} else if err != nil {
		logger.Error("GetFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	stations := make([]models.Station, 0)
	cursor, err := stationsCollection.Find(context.TODO(), bson.M{"factory_id": factory.ID})
	if err != nil {
		logger.Error("GetFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &stations); err != nil {
		logger.Error("GetFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{
		"id":              factory.ID,
		"name":            factory.Name,
		"description":     factory.Description,
		"created_by_user": factory.CreatedByUser,
		"creation_date":   factory.CreationDate,
		"stations":        stations,
	})
}

func (umh FactoriesHandler) GetAllFactories(c *gin.Context) {
	type extendedFactory struct {
		ID            primitive.ObjectID `json:"id" bson:"_id"`
		Name          string             `json:"name" bson:"name"`
		Description   string             `json:"description" bson:"description"`
		CreatedByUser string             `json:"created_by_user" bson:"created_by_user"`
		CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
		UserAvatarId  int                `json:"user_avatar_id" bson:"user_avatar_id"`
	}

	var factories []extendedFactory
	cursor, err := factoriesCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$lookup", bson.D{{"from", "users"}, {"localField", "created_by_user"}, {"foreignField", "username"}, {"as", "user"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$user"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"description", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"user_avatar_id", "$user.avatar_id"}}}},
		bson.D{{"$project", bson.D{{"user", 0}}}},
	})

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
	exist, factory, err := IsFactoryExist(factoryName)
	if err != nil {
		logger.Error("RemoveFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Factory does not exist"})
		return
	}

	err = removeStations(factory.ID)
	if err != nil {
		logger.Error("RemoveFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	_, err = factoriesCollection.DeleteOne(context.TODO(), bson.M{"name": factoryName})
	if err != nil {
		logger.Error("RemoveFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	logger.Info("Factory " + factoryName + " has been created")
	c.IndentedJSON(200, gin.H{})
}

func (umh FactoriesHandler) EditFactory(c *gin.Context) {
	var body models.EditFactorySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	factoryName := strings.ToLower(body.FactoryName)
	exist, _, err := IsFactoryExist(factoryName)
	if err != nil {
		logger.Error("EditFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Factory with that name does not exist"})
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
