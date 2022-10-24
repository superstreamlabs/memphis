// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package server

import (
	"context"
	"memphis-broker/models"
	"memphis-broker/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TagsHandler struct{ S *Server }

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
	switch entity {
	case "station":
		for _, tagToCreate := range tags {
			name := strings.ToLower(tagToCreate.Name)
			exist, tag, err := IsTagExist(name)
			if err != nil {
				return err
			}
			if !exist {
				err = CreateTag(name, entity, entity_id, tagToCreate.Color)
				if err != nil {
					return err
				}
			}
			_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID}, bson.M{"$addToSet": bson.M{"stations": entity_id}})
			if err != nil {
				return err
			}
		}

	case "schema":
		for _, tagToCreate := range tags {
			name := strings.ToLower(tagToCreate.Name)
			exist, tag, err := IsTagExist(name)
			if err != nil {
				return err
			}
			if !exist {
				err = CreateTag(name, entity_type, entity_id, tag.Color)
				if err != nil {
					return err
				}
				return nil
			}
			_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID}, bson.M{"$addToSet": bson.M{"schemas": entity_id}})
			if err != nil {
				return err
			}
		}
		// case "user":
		// for _, tagToCreate := range tags {
		// 	name := strings.ToLower(tagToCreate.Name)
		// 	exist, tag, err := IsTagExist(name)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	if !exist {
		// 		err = CreateTag(name, entity_type, entity_id, tag.Color)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		return nil
		// 	}
		// 	_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID}, bson.M{"$addToSet": bson.M{"users": entity_id}})
		// 	if err != nil {
		// 		return err
		// 	}
		// }
	}
	return nil
}

func DeleteTagsByStation(id primitive.ObjectID) {
	_, err := tagsCollection.UpdateMany(context.TODO(), bson.M{}, bson.M{"$pull": bson.M{"stations": id}})
	if err != nil {
		serv.Errorf("DeleteTagsByStation error: " + err.Error())
		return
	}
}

func DeleteTagsBySchema(id primitive.ObjectID) {
	_, err := tagsCollection.UpdateMany(context.TODO(), bson.M{}, bson.M{"$pull": bson.M{"schemas": id}})
	if err != nil {
		serv.Errorf("DeleteTagsBySchema error: " + err.Error())
		return
	}
}

func DeleteTagsByUser(id primitive.ObjectID) {
	_, err := tagsCollection.UpdateMany(context.TODO(), bson.M{}, bson.M{"$pull": bson.M{"users": id}})
	if err != nil {
		serv.Errorf("DeleteTagsByUser error: " + err.Error())
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
	exist, tag, err := IsTagExist(name)

	if err != nil {
		serv.Errorf("RemoveTag error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.IndentedJSON(200, []string{})
		return
	}
	entity := strings.ToLower(body.EntityType)
	switch entity {
	case "station":
		station_name, err := StationNameFromStr(body.EntityName)
		if err != nil {
			serv.Errorf("RemoveTag error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		exist, station, err := IsStationExist(station_name)
		if err != nil {
			serv.Errorf("RemoveTag error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			serv.Warnf("RemoveTag error: Station does not exist")
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tags, station does not exist"})
			return
		}
		_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID},
			bson.M{"$pull": bson.M{"stations": station.ID}})
		if err != nil {
			serv.Errorf("RemoveTag error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

	case "schema":
		exist, schema, err := IsSchemaExist(body.EntityName)
		if err != nil {
			serv.Errorf("RemoveTag error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			serv.Warnf("RemoveTag error: Schema does not exist")
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tags, schema does not exist"})
			return
		}
		_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID},
			bson.M{"$pull": bson.M{"schemas": schema.ID}})
		if err != nil {
			serv.Errorf("RemoveTag error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	// case "user":
	// 	exist, user, err := IsUserExist(body.EntityName)
	// 	if err != nil {
	// 		serv.Errorf("RemoveTag error: " + err.Error())
	// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
	// 		return
	// 	}
	// 	if !exist {
	// 		serv.Warnf("RemoveTag error: User with does not exist")
	// 		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tags, user does not exist"})
	//
	// 		return
	// 	}
	// 	_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID},
	// 		bson.M{"$pull": bson.M{"users": user.ID}})
	// 	if err != nil {
	// 		serv.Errorf("RemoveTag error: " + err.Error())
	// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
	// 		return
	// 	}
	default:
		serv.Warnf("RemoveTag error: wrong input")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tags, wrong input"})
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
	var tagsRes []models.Tag
	var entityDBList string
	entity := strings.ToLower(body.EntityType)
	switch entity {
	case "station":
		entityDBList = "stations"
	case "schema":
		entityDBList = "schemas"
	case "user":
		entityDBList = "users"
	default:
		serv.Errorf("UpdateTagsForEntity error: wrong entity type")
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(body.TagsToAdd) > 0 {
		for _, tagToAdd := range body.TagsToAdd {
			name := strings.ToLower(tagToAdd.Name)
			exist, _, err := IsTagExist(name)
			if err != nil {
				serv.Errorf("UpdateTagsForEntity error: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			if !exist {
				err = CreateTag(name, body.EntityType, body.EntityID, tagToAdd.Color)
				if err != nil {
					serv.Errorf("UpdateTagsForEntity error: " + err.Error())
					c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
					return
				}
			}
			_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tagToAdd.ID}, bson.M{"$addToSet": bson.M{entityDBList: body.EntityID}})
			if err != nil {
				serv.Errorf("UpdateTagsForEntity error: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
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
					bson.M{"$pull": bson.M{entityDBList: body.EntityID}})
				if err != nil {
					serv.Errorf("UpdateTagsForEntity error: " + err.Error())
					c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
					return
				}
			}
		}
	}
	tags, err := GetTagsByStation(body.EntityID)
	if err != nil {
		serv.Errorf("UpdateTagsForEntity error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	tagsRes = tags

	c.IndentedJSON(200, tagsRes)
}

func (th TagsHandler) GetTagsByStation(station_id primitive.ObjectID) ([]models.Tag, error) {
	return GetTagsByStation(station_id)
}

func GetTagsByStation(station_id primitive.ObjectID) ([]models.Tag, error) {
	var tags []models.Tag
	cursor, err := tagsCollection.Find(context.TODO(), bson.M{"stations": station_id})

	if err != nil {
		return tags, err
	}

	if err = cursor.All(context.TODO(), &tags); err != nil {
		return tags, err
	}

	if len(tags) == 0 {
		tags = []models.Tag{}
	}

	return tags, nil
}

func (th TagsHandler) GetTagsBySchema(schema_id primitive.ObjectID) ([]models.Tag, error) {
	return GetTagsBySchema(schema_id)
}
func GetTagsBySchema(schema_id primitive.ObjectID) ([]models.Tag, error) {
	var tags []models.Tag
	cursor, err := tagsCollection.Find(context.TODO(), bson.M{"schemas": schema_id})
	if err != nil {
		return tags, err
	}

	if err = cursor.All(context.TODO(), &tags); err != nil {
		return tags, err
	}

	if len(tags) == 0 {
		tags = []models.Tag{}
	}

	return tags, nil
}

func (th TagsHandler) GetTagsByUser(user_id primitive.ObjectID) ([]models.Tag, error) {
	var tags []models.Tag
	cursor, err := tagsCollection.Find(context.TODO(), bson.M{"users": user_id})
	if err != nil {
		return tags, err
	}

	if err = cursor.All(context.TODO(), &tags); err != nil {
		return tags, err
	}

	if len(tags) == 0 {
		tags = []models.Tag{}
	}

	return tags, nil
}

func (th TagsHandler) GetTags(c *gin.Context) {
	var body models.GetTagsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	entity := strings.ToLower(body.EntityType)
	var tags []models.Tag
	switch entity {
	case "stations":
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
	case "users":
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
	case "schemas":
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
		tags = []models.Tag{}
	}

	c.IndentedJSON(200, tags)
}

func (th TagsHandler) GetUsedTags(c *gin.Context) {
	var tags []models.Tag
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

	c.IndentedJSON(200, tags)
}
