// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"errors"

	"memphis-broker/models"
	"memphis-broker/utils"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FactoriesHandler struct{ S *Server }

func validateFactoryName(factoryName string) error {
	re := regexp.MustCompile("^[a-z0-9_]*$")

	validName := re.MatchString(factoryName)
	if !validName {
		return errors.New("factory name has to include only letters, numbers and _")
	}
	return nil
}

// TODO remove the stations resources - functions, connectors
func removeStations(s *Server, factoryId primitive.ObjectID) error {
	var stations []models.Station
	cursor, err := stationsCollection.Find(context.TODO(), bson.M{
		"factory_id": factoryId,
		"$or": []interface{}{
			bson.M{"is_deleted": false},
			bson.M{"is_deleted": bson.M{"$exists": false}},
		},
	})
	if err != nil {
		return err
	}

	if err = cursor.All(context.TODO(), &stations); err != nil {
		return err
	}

	for _, station := range stations {
		err = s.RemoveStream(station.Name)
		if err != nil {
			return err
		}

		_, err = producersCollection.UpdateMany(context.TODO(),
			bson.M{"station_id": station.ID},
			bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
		)
		if err != nil {
			return err
		}

		_, err = consumersCollection.UpdateMany(context.TODO(),
			bson.M{"station_id": station.ID},
			bson.M{"$set": bson.M{"is_active": false, "is_deleted": true}},
		)
		if err != nil {
			return err
		}

		err = RemovePoisonMsgsByStation(station.Name)
		if err != nil {
			serv.Errorf("removeStations error: " + err.Error())
		}

		err = RemoveAllAuditLogsByStation(station.Name)
		if err != nil {
			serv.Errorf("removeStations error: " + err.Error())
		}
	}

	_, err = stationsCollection.UpdateMany(context.TODO(),
		bson.M{
			"factory_id": factoryId,
			"$or": []interface{}{
				bson.M{"is_deleted": false},
				bson.M{"is_deleted": bson.M{"$exists": false}},
			},
		},
		bson.M{"$set": bson.M{"is_deleted": true}},
	)
	if err != nil {
		return err
	}

	return nil
}

func (fh FactoriesHandler) CreateFactory(c *gin.Context) {
	var body models.CreateFactorySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	factoryName := strings.ToLower(body.Name)
	err := validateFactoryName(factoryName)
	if err != nil {
		serv.Errorf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, _, err := IsFactoryExist(factoryName)
	if err != nil {
		serv.Errorf("CreateFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		serv.Errorf("Factory with that name is already exist")
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
		IsDeleted:     false,
	}

	_, err = factoriesCollection.InsertOne(context.TODO(), newFactory)
	if err != nil {
		serv.Errorf("CreateFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	serv.Noticef("Factory " + factoryName + " has been created")
	c.IndentedJSON(200, gin.H{
		"id":              newFactory.ID,
		"name":            newFactory.Name,
		"description":     newFactory.Description,
		"created_by_user": newFactory.CreatedByUser,
		"creation_date":   newFactory.CreationDate,
	})
}

var ErrFactoryAlreadyExists = errors.New("memphis: factory already exists")

func createFactoryDirect(cfr *createFactoryRequest) error {
	factoryName := strings.ToLower(cfr.FactoryName)
	err := validateFactoryName(factoryName)
	if err != nil {
		serv.Errorf(err.Error())
		return err
	}

	exist, _, err := IsFactoryExist(factoryName)
	if err != nil {
		serv.Errorf("CreateFactory error: " + err.Error())
		return err
	}

	if exist {
		serv.Errorf("Factory with that name already exists")
		return ErrFactoryAlreadyExists
	}

	newFactory := models.Factory{
		ID:            primitive.NewObjectID(),
		Name:          factoryName,
		Description:   strings.ToLower(cfr.FactoryDesc),
		CreatedByUser: cfr.Username,
		CreationDate:  time.Now(),
		IsDeleted:     false,
	}

	_, err = factoriesCollection.InsertOne(context.TODO(), newFactory)
	if err != nil {
		serv.Errorf("CreateFactory error: " + err.Error())
		return err
	}

	serv.Noticef("Factory " + factoryName + " has been created")
	return nil
}

func (fh FactoriesHandler) GetFactoryDetails(factoryName string) (map[string]interface{}, error) {
	var factory models.Factory
	err := factoriesCollection.FindOne(context.TODO(), bson.M{
		"name": factoryName,
		"$or": []interface{}{
			bson.M{"is_deleted": false},
			bson.M{"is_deleted": bson.M{"$exists": false}},
		},
	}).Decode(&factory)
	if err != nil {
		return map[string]interface{}{}, err
	}

	stations := make([]models.Station, 0)
	cursor, err := stationsCollection.Find(context.TODO(), bson.M{
		"factory_id": factory.ID,
		"$or": []interface{}{
			bson.M{"is_deleted": false},
			bson.M{"is_deleted": bson.M{"$exists": false}},
		},
	})
	if err != nil {
		return map[string]interface{}{}, err
	}

	if err = cursor.All(context.TODO(), &stations); err != nil {
		return map[string]interface{}{}, err
	}

	_, user, err := IsUserExist(factory.CreatedByUser)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{
		"id":              factory.ID,
		"name":            factory.Name,
		"description":     factory.Description,
		"created_by_user": factory.CreatedByUser,
		"creation_date":   factory.CreationDate,
		"stations":        stations,
		"user_avatar_id":  user.AvatarId,
	}, nil
}

func (fh FactoriesHandler) GetFactory(c *gin.Context) {
	var body models.GetFactorySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	factoryName := strings.ToLower(body.FactoryName)

	factory, err := fh.GetFactoryDetails(factoryName)
	if err == mongo.ErrNoDocuments {
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Factory does not exist"})
		return
	} else if err != nil {
		serv.Errorf("GetFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, factory)
}

func (fh FactoriesHandler) GetAllFactoriesDetails() ([]models.ExtendedFactory, error) {
	var factories []models.ExtendedFactory
	cursor, err := factoriesCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", bson.D{{"$or", []interface{}{
			bson.D{{"is_deleted", false}},
			bson.D{{"is_deleted", bson.D{{"$exists", false}}}},
		}}}}},
		bson.D{{"$lookup", bson.D{{"from", "users"}, {"localField", "created_by_user"}, {"foreignField", "username"}, {"as", "user"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$user"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"description", 1}, {"created_by_user", 1}, {"creation_date", 1}, {"user_avatar_id", "$user.avatar_id"}}}},
		bson.D{{"$project", bson.D{{"user", 0}}}},
	})

	if err != nil {
		return factories, err
	}

	if err = cursor.All(context.TODO(), &factories); err != nil {
		return factories, err
	}

	if len(factories) == 0 {
		return []models.ExtendedFactory{}, nil
	}

	return factories, nil
}

func (fh FactoriesHandler) GetAllFactories(c *gin.Context) {
	factories, err := fh.GetAllFactoriesDetails()
	if err != nil {
		serv.Errorf("GetAllFactories error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, factories)
}

func (fh FactoriesHandler) RemoveFactory(c *gin.Context) {
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}
	var body models.RemoveFactorySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	factoryName := strings.ToLower(body.FactoryName)
	exist, factory, err := IsFactoryExist(factoryName)
	if err != nil {
		serv.Errorf("RemoveFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Errorf("Factory does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Factory does not exist"})
		return
	}

	err = removeStations(fh.S, factory.ID)
	if err != nil {
		serv.Errorf("RemoveFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	_, err = factoriesCollection.UpdateOne(context.TODO(),
		bson.M{
			"name": factoryName,
			"$or": []interface{}{
				bson.M{"is_deleted": false},
				bson.M{"is_deleted": bson.M{"$exists": false}},
			},
		},
		bson.M{"$set": bson.M{"is_deleted": true}},
	)
	if err != nil {
		serv.Errorf("RemoveFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	serv.Noticef("Factory " + factoryName + " has been deleted")
	c.IndentedJSON(200, gin.H{})
}

func (s *Server) RemoveFactoryDirect(dfr *destroyFactoryRequest) error {
	factoryName := strings.ToLower(dfr.FactoryName)
	exist, factory, err := IsFactoryExist(factoryName)
	if err != nil {
		serv.Errorf("RemoveFactory error: " + err.Error())
		return err
	}
	if !exist {
		serv.Errorf("Factory does not exist")
		return errors.New("memphis: factory does not exist")
	}

	err = removeStations(s, factory.ID)
	if err != nil {
		serv.Errorf("RemoveFactory error: " + err.Error())
		return err
	}

	_, err = factoriesCollection.UpdateOne(context.TODO(),
		bson.M{
			"name": factoryName,
			"$or": []interface{}{
				bson.M{"is_deleted": false},
				bson.M{"is_deleted": bson.M{"$exists": false}},
			},
		},
		bson.M{"$set": bson.M{"is_deleted": true}},
	)
	if err != nil {
		serv.Errorf("RemoveFactory error: " + err.Error())
		return err
	}

	serv.Noticef("Factory " + factoryName + " has been deleted")
	return nil
}

func (fh FactoriesHandler) EditFactory(c *gin.Context) {
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}
	var body models.EditFactorySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	factoryName := strings.ToLower(body.FactoryName)
	exist, factory, err := IsFactoryExist(factoryName)
	if err != nil {
		serv.Errorf("EditFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Errorf("Factory with that name does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Factory with that name does not exist"})
		return
	}

	newName := strings.ToLower(body.NewName)
	exist, _, err = IsFactoryExist(newName)
	if err != nil {
		serv.Errorf("EditFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		serv.Errorf("Factory with that name already exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Factory with that name already exist"})
		return
	}

	if body.NewName != "" {
		factory.Name = newName
	}

	if body.NewDescription != "" {
		factory.Description = strings.ToLower(body.NewDescription)
	}

	_, err = factoriesCollection.UpdateOne(context.TODO(),
		bson.M{
			"name": factoryName,
			"$or": []interface{}{
				bson.M{"is_deleted": false},
				bson.M{"is_deleted": bson.M{"$exists": false}},
			},
		},
		bson.M{"$set": bson.M{"name": factory.Name, "description": factory.Description}},
	)
	if err != nil {
		serv.Errorf("EditFactory error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, factory)
}
