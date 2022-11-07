// Credit for The NATS.IO Authors
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
// limitations under the License.package server
package server

import (
	"context"
	"errors"
	"memphis-broker/models"
	"memphis-broker/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TagsHandler struct{ S *Server }

func validateEntityType(entity string) error {
	switch entity {
	case "station", "schema", "user":
		return nil
	default:
		return errors.New("Entity type is not valid")
	}
}

func CreateTag(name string, entity_type string, entity_id primitive.ObjectID, color string) error {
	name = strings.ToLower(name)
	entity := strings.ToLower(entity_type)
	var newTag models.Tag
	stationArr := []primitive.ObjectID{}
	schemaArr := []primitive.ObjectID{}
	userArr := []primitive.ObjectID{}
	switch entity {
	case "station":
		stationArr = append(stationArr, entity_id)
	case "schema":
		schemaArr = append(schemaArr, entity_id)
		// case "user":
		// 	userArr = append(userArr, entity_id)
	}
	newTag = models.Tag{
		ID:       primitive.NewObjectID(),
		Name:     name,
		Color:    color,
		Stations: stationArr,
		Schemas:  schemaArr,
		Users:    userArr,
	}

	filter := bson.M{"name": newTag.Name}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":      newTag.ID,
			"name":     newTag.Name,
			"color":    newTag.Color,
			"stations": newTag.Stations,
			"schemas":  newTag.Schemas,
			"users":    newTag.Users,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := tagsCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func AddTagsToEntity(tags []models.CreateTag, entity_type string, entity_id primitive.ObjectID) error {
	if len(tags) == 0 {
		return nil
	}
	entity := strings.ToLower(entity_type)
	err := validateEntityType(entity)
	if err != nil {
		return err
	}
	var entityDBList string
	for _, tagToCreate := range tags {
		exist, _, err := IsTagExist(tagToCreate.Name)
		if err != nil {
			return err
		}
		if !exist {
			err = CreateTag(tagToCreate.Name, entity_type, entity_id, tagToCreate.Color)
			if err != nil {
				return err
			}
		} else {
			switch entity {
			case "station":
				entityDBList = "stations"
			case "schema":
				entityDBList = "schemas"
			case "user":
				entityDBList = "users"
			}
			filter := bson.M{"name": tagToCreate.Name}
			update := bson.M{
				"$addToSet": bson.M{entityDBList: entity_id},
			}
			opts := options.Update().SetUpsert(true)
			_, err = tagsCollection.UpdateOne(context.TODO(), filter, update, opts)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func DeleteTagsFromStation(id primitive.ObjectID) {
	_, err := tagsCollection.UpdateMany(context.TODO(), bson.M{}, bson.M{"$pull": bson.M{"stations": id}})
	if err != nil {
		serv.Errorf("DeleteTagsFromStation error: " + err.Error())
		return
	}
}

func DeleteTagsFromSchema(id primitive.ObjectID) {
	_, err := tagsCollection.UpdateMany(context.TODO(), bson.M{}, bson.M{"$pull": bson.M{"schemas": id}})
	if err != nil {
		serv.Errorf("DeleteTagsFromSchema error: " + err.Error())
		return
	}
}

func DeleteTagsFromUser(id primitive.ObjectID) {
	_, err := tagsCollection.UpdateMany(context.TODO(), bson.M{}, bson.M{"$pull": bson.M{"users": id}})
	if err != nil {
		serv.Errorf("DeleteTagsFromUser error: " + err.Error())
		return
	}
}

func (th TagsHandler) CreateNewTag(c *gin.Context) {
	var body models.CreateTag
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	name := strings.ToLower(body.Name)
	exist, _, err := IsTagExist(name)
	if err != nil {
		serv.Errorf("CreateNewTag error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		serv.Warnf("CreateNewTag error: Tag with the same name already exists")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Tag with the same name already exists"})
		return
	}
	var color string
	if len(body.Color) > 0 {
		color = body.Color
	} else {
		color = "101, 87, 255" // default memphis-purple color
	}
	var newTag models.Tag
	stationArr := []primitive.ObjectID{}
	schemaArr := []primitive.ObjectID{}
	userArr := []primitive.ObjectID{}
	newTag = models.Tag{
		ID:       primitive.NewObjectID(),
		Name:     name,
		Color:    color,
		Stations: stationArr,
		Schemas:  schemaArr,
		Users:    userArr,
	}

	filter := bson.M{"name": newTag.Name}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":      newTag.ID,
			"name":     newTag.Name,
			"color":    newTag.Color,
			"stations": newTag.Stations,
			"schemas":  newTag.Schemas,
			"users":    newTag.Users,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = tagsCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		serv.Errorf("CreateNewTag error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, newTag)
}

func (th TagsHandler) RemoveTag(c *gin.Context) {
	var body models.RemoveTagSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	name := strings.ToLower(body.Name)
	entity := strings.ToLower(body.EntityType)
	err := validateEntityType(entity)
	if err != nil {
		serv.Warnf("RemoveTag error: " + err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	var entity_id primitive.ObjectID
	var entityDBList string
	switch entity {
	case "station":
		station_name, err := StationNameFromStr(body.EntityName)
		if err != nil {
			serv.Warnf("RemoveTag error: " + err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
		exist, station, err := IsStationExist(station_name)
		if err != nil {
			serv.Errorf("RemoveTag error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			c.IndentedJSON(200, []string{})
			return
		}
		entity_id = station.ID
		entityDBList = "stations"

	case "schema":
		exist, schema, err := IsSchemaExist(body.EntityName)
		if err != nil {
			serv.Errorf("RemoveTag error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			c.IndentedJSON(200, []string{})
			return
		}
		entity_id = schema.ID
		entityDBList = "schemas"

	// case "user":
	// 	exist, user, err := IsUserExist(body.EntityName)
	// 	if err != nil {
	// 		serv.Errorf("RemoveTag error: " + err.Error())
	// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
	// 		return
	// 	}
	// 	if !exist {
	// 		c.IndentedJSON(200, []string{})
	// 		return
	// 	}
	// entity_id = user.ID
	// entityDBList = "schemas"

	default:
		serv.Warnf("RemoveTag error: unsupported entity type")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tags, unsupported entity type"})
		return
	}
	_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"name": name},
		bson.M{"$pull": bson.M{entityDBList: entity_id}})
	if err != nil {
		serv.Errorf("RemoveTag error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, []string{})
}

func (th TagsHandler) UpdateTagsForEntity(c *gin.Context) {
	var body models.UpdateTagsForEntitySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	var entityDBList string
	entity := strings.ToLower(body.EntityType)
	err := validateEntityType(entity)
	var entity_id primitive.ObjectID
	if err != nil {
		serv.Warnf("UpdateTagsForEntity error: " + err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	switch entity {
	case "station":
		station_name, err := StationNameFromStr(body.EntityName)
		if err != nil {
			serv.Warnf("UpdateTagsForEntity error: " + err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
		exist, station, err := IsStationExist(station_name)
		if err != nil {
			serv.Errorf("UpdateTagsForEntity error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			c.IndentedJSON(200, []string{})
			return
		}
		entity_id = station.ID
		entityDBList = "stations"

	case "schema":
		exist, schema, err := IsSchemaExist(body.EntityName)
		if err != nil {
			serv.Errorf("UpdateTagsForEntity error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			c.IndentedJSON(200, []string{})
			return
		}
		entity_id = schema.ID
		entityDBList = "schemas"

	// case "user":
	// 	exist, user, err := IsUserExist(body.EntityName)
	// 	if err != nil {
	// 		serv.Errorf("UpdateTagsForEntity error: " + err.Error())
	// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
	// 		return
	// 	}
	// 	if !exist {
	// 		c.IndentedJSON(200, []string{})
	// 		return
	// 	}
	// entity_id = user.ID
	// entityDBList = "schemas"

	default:
		serv.Warnf("UpdateTagsForEntity error: unsupported entity type")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tags, unsupported entity type"})
		return
	}
	if len(body.TagsToAdd) > 0 {
		for _, tagToAdd := range body.TagsToAdd {
			name := strings.ToLower(tagToAdd.Name)
			exist, tag, err := IsTagExist(name)
			if err != nil {
				serv.Errorf("UpdateTagsForEntity error: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			if !exist {
				err = CreateTag(name, body.EntityType, entity_id, tagToAdd.Color)
				if err != nil {
					serv.Errorf("UpdateTagsForEntity error: " + err.Error())
					c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
					return
				}
			} else {
				_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID}, bson.M{"$addToSet": bson.M{entityDBList: entity_id}})
				if err != nil {
					serv.Errorf("UpdateTagsForEntity error: " + err.Error())
					c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
					return
				}
			}

		}
	}
	if len(body.TagsToRemove) > 0 {
		for _, tagToRemove := range body.TagsToRemove {
			name := strings.ToLower(tagToRemove)
			exist, tag, err := IsTagExist(name)
			if err != nil {
				serv.Errorf("UpdateTagsForEntity error: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			if exist {
				_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID},
					bson.M{"$pull": bson.M{entityDBList: entity_id}})
				if err != nil {
					serv.Errorf("UpdateTagsForEntity error: " + err.Error())
					c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
					return
				}
			}
		}
	}
	var tags []models.CreateTag
	switch entity {
	case "station":
		tags, err = th.GetTagsByStation(entity_id)
		if err != nil {
			serv.Errorf("UpdateTagsForEntity error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	case "schema":
		tags, err = th.GetTagsBySchema(entity_id)
		if err != nil {
			serv.Errorf("UpdateTagsForEntity error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	case "user":
		tags, err = th.GetTagsByUser(entity_id)
		if err != nil {
			serv.Errorf("UpdateTagsForEntity error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}
	c.IndentedJSON(200, tags)
}

func (th TagsHandler) GetTagsByStation(station_id primitive.ObjectID) ([]models.CreateTag, error) {
	var tags []models.Tag
	var tagsRes []models.CreateTag
	cursor, err := tagsCollection.Find(context.TODO(), bson.M{"stations": station_id})
	if err != nil {
		return tagsRes, err
	}
	if err = cursor.All(context.TODO(), &tags); err != nil {
		return tagsRes, err
	}
	if len(tags) == 0 {
		tagsRes = []models.CreateTag{}
	}
	for _, tag := range tags {
		tagRes := models.CreateTag{
			Name:  tag.Name,
			Color: tag.Color,
		}
		tagsRes = append(tagsRes, tagRes)
	}
	return tagsRes, nil
}

func (th TagsHandler) GetTagsBySchema(schema_id primitive.ObjectID) ([]models.CreateTag, error) {
	var tags []models.Tag
	var tagsRes []models.CreateTag
	cursor, err := tagsCollection.Find(context.TODO(), bson.M{"schemas": schema_id})
	if err != nil {
		return tagsRes, err
	}
	if err = cursor.All(context.TODO(), &tags); err != nil {
		return tagsRes, err
	}
	if len(tags) == 0 {
		tagsRes = []models.CreateTag{}
	}
	for _, tag := range tags {
		tagRes := models.CreateTag{
			Name:  tag.Name,
			Color: tag.Color,
		}
		tagsRes = append(tagsRes, tagRes)
	}
	return tagsRes, nil
}

func (th TagsHandler) GetTagsByUser(user_id primitive.ObjectID) ([]models.CreateTag, error) {
	var tags []models.Tag
	var tagsRes []models.CreateTag
	cursor, err := tagsCollection.Find(context.TODO(), bson.M{"users": user_id})
	if err != nil {
		return tagsRes, err
	}
	if err = cursor.All(context.TODO(), &tags); err != nil {
		return tagsRes, err
	}
	if len(tags) == 0 {
		tagsRes = []models.CreateTag{}
	}
	for _, tag := range tags {
		tagRes := models.CreateTag{
			Name:  tag.Name,
			Color: tag.Color,
		}
		tagsRes = append(tagsRes, tagRes)
	}
	return tagsRes, nil
}

func (th TagsHandler) GetTags(c *gin.Context) {
	var body models.GetTagsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	entity := strings.ToLower(body.EntityType)
	if entity != "" {
		err := validateEntityType(entity)
		if err != nil {
			serv.Warnf("GetTags error: " + err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}
	var tags []models.Tag
	var tagsRes []models.CreateTag
	switch entity {
	case "station":
		cursor, err := tagsCollection.Find(context.TODO(), bson.M{"stations": bson.M{"$not": bson.M{"$size": 0}}})
		if err != nil {
			serv.Errorf("GetTags error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if err = cursor.All(context.TODO(), &tags); err != nil {
			serv.Errorf("GetTags error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	case "user":
		cursor, err := tagsCollection.Find(context.TODO(), bson.M{"users": bson.M{"$not": bson.M{"$size": 0}}})
		if err != nil {
			serv.Errorf("GetTags error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		if err = cursor.All(context.TODO(), &tags); err != nil {
			serv.Errorf("GetTags error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	case "schema":
		cursor, err := tagsCollection.Find(context.TODO(), bson.M{"schemas": bson.M{"$not": bson.M{"$size": 0}}})
		if err != nil {
			serv.Errorf("GetTags error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		if err = cursor.All(context.TODO(), &tags); err != nil {
			serv.Errorf("GetTags error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	default:
		cursor, err := tagsCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			serv.Errorf("GetTags error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		if err = cursor.All(context.TODO(), &tags); err != nil {
			serv.Errorf("GetTags error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}
	if len(tags) == 0 {
		tagsRes = []models.CreateTag{}
	}
	for _, tag := range tags {
		tagRes := models.CreateTag{
			Name:  tag.Name,
			Color: tag.Color,
		}
		tagsRes = append(tagsRes, tagRes)
	}
	c.IndentedJSON(200, tagsRes)
}

func (th TagsHandler) GetUsedTags(c *gin.Context) {
	var tags []models.Tag
	var tagsRes []models.CreateTag
	filter := bson.M{"$or": []interface{}{bson.M{"schemas": bson.M{"$exists": true, "$not": bson.M{"$size": 0}}}, bson.M{"stations": bson.M{"$exists": true, "$not": bson.M{"$size": 0}}}, bson.M{"users": bson.M{"$exists": true, "$not": bson.M{"$size": 0}}}}}
	cursor, err := tagsCollection.Find(context.TODO(), filter)
	if err != nil {
		serv.Errorf("GetUsedTags error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if err = cursor.All(context.TODO(), &tags); err != nil {
		serv.Errorf("GetUsedTags error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if len(tags) == 0 {
		tagsRes = []models.CreateTag{}
	}
	for _, tag := range tags {
		tagRes := models.CreateTag{
			Name:  tag.Name,
			Color: tag.Color,
		}
		tagsRes = append(tagsRes, tagRes)
	}

	c.IndentedJSON(200, tagsRes)
}
