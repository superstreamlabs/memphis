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

const (
	tagObjectName = "Tag"
)

func CreateTag(name string, entity_type string, entity_id primitive.ObjectID, background_color string, text_color string) error {
	name = strings.ToLower(name)
	exist, _, err := IsTagExist(name)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	var newTag models.Tag
	stationArr := []primitive.ObjectID{}
	schemaArr := []primitive.ObjectID{}
	userArr := []primitive.ObjectID{}
	switch entity_type {
	case "station":
		stationArr = append(stationArr, entity_id)
		// case "schema":
		// 	schemaArr = append(schemaArr, entity_id)
		// case "user":
		// 	userArr = append(userArr, entity_id)
	}
	newTag = models.Tag{
		ID:       primitive.NewObjectID(),
		Name:     name,
		ColorBG:  background_color,
		ColorTXT: text_color,
		Stations: stationArr,
		Schemas:  schemaArr,
		Users:    userArr,
	}

	filter := bson.M{"name": newTag.Name}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":       newTag.ID,
			"name":      newTag.Name,
			"color_bg":  newTag.ColorBG,
			"color_txt": newTag.ColorTXT,
			"stations":  newTag.Stations,
			"schemas":   newTag.Schemas,
			"users":     newTag.Users,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = tagsCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func AddTagsToEntity(tags []models.CreateTag, entity_type string, entity_id primitive.ObjectID) error {
	if len(tags) == 0 {
		return nil
	}
	switch entity_type {
	case "station":
		for _, tagToCreate := range tags {
			name := strings.ToLower(tagToCreate.Name)
			exist, tag, err := IsTagExist(name)
			if err != nil {
				return err
			}
			if !exist {
				err = CreateTag(name, entity_type, entity_id, tag.ColorBG, tag.ColorTXT)
				if err != nil {
					return err
				}
				return nil
			}
			_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID}, bson.M{"$addToSet": bson.M{"stations": entity_id}})
			if err != nil {
				return err
			}
		}

		// case "schema":
		// for _, tagToCreate := range tags {
		// 	name := strings.ToLower(tagToCreate.Name)
		// 	exist, tag, err := IsTagExist(name)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	if !exist {
		// 		err = CreateTag(name, entity_type, entity_id, tag.ColorBG, tag.ColorTXT)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		return nil
		// 	}
		// 	_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID}, bson.M{"$addToSet": bson.M{"schemas": schema.ID}})
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		// case "user":
		// for _, tagToCreate := range tags {
		// 	name := strings.ToLower(tagToCreate.Name)
		// 	exist, tag, err := IsTagExist(name)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	if !exist {
		// 		err = CreateTag(name, entity_type, entity_id, tag.ColorBG, tag.ColorTXT)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		return nil
		// 	}
		// 	_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID}, bson.M{"$addToSet": bson.M{"users": user.ID}})
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
		serv.Errorf("Failed deleting tags: %v", err.Error())
		return
	}
}

// func DeleteTagsBySchema(id primitive.ObjectID) {
// 	_, err: = tagsCollection.UpdateMany(context.TODO(), bson.M{}, bson.M{"$pull": bson.M{"schemas": id}})
// 	if err != nil {
// 		serv.Errorf("Failed deleting tags: %v", err.Error())
// 		return
// 	}
// }

func DeleteTagsByUser(id primitive.ObjectID) {
	_, err := tagsCollection.UpdateMany(context.TODO(), bson.M{}, bson.M{"$pull": bson.M{"users": id}})
	if err != nil {
		serv.Errorf("Failed deleting tags: %v", err.Error())
		return
	}
}

func checkIfEmptyAndDelete(name string) {
	exist, tag, _ := IsTagExist(name)
	if exist {

		if len(tag.Schemas) == 0 && len(tag.Stations) == 0 && len(tag.Users) == 0 {
			_, err := tagsCollection.DeleteOne(context.TODO(), bson.M{"_id": tag.ID})
			if err != nil {
				serv.Errorf("Delete tag error:" + err.Error())
			}
		}
	}
}

func (th TagsHandler) CreateTags(c *gin.Context) {
	var body models.CreateTagsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	station_name, err := StationNameFromStr(body.EntityName)
	if err != nil {
		serv.Errorf("RemoveTags error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	exist, station, err := IsStationExist(station_name)
	if err != nil {
		serv.Errorf("RemoveTags error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("Station does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tags, station does not exist"})
		return
	}
	err = AddTagsToEntity(body.Tags, body.EntityType, station.ID)
	if err != nil {
		serv.Errorf("Failed creating tag: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, []string{})
}

func (th TagsHandler) RemoveTags(c *gin.Context) {
	var body models.RemoveTagsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	for _, tagName := range body.Names {
		name := strings.ToLower(tagName)
		exist, tag, err := IsTagExist(name)

		if err != nil {
			serv.Errorf("RemoveTags error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			c.IndentedJSON(200, []string{})
			return
		}
		switch body.EntityType {
		case "station":
			station_name, err := StationNameFromStr(body.EntityName)
			if err != nil {
				serv.Errorf("RemoveTags error: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			exist, station, err := IsStationExist(station_name)
			if err != nil {
				serv.Errorf("RemoveTags error: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			if !exist {
				serv.Warnf("Station does not exist")
				c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tags, station does not exist"})
				return
			}
			_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID},
				bson.M{"$pull": bson.M{"stations": station.ID}})
			if err != nil {
				serv.Errorf("RemoveTags error: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}

		// case "schema":
		// 	exist, schema, err := IsSchemaExist(body.EntityName)
		// 	if err != nil {
		// 		serv.Errorf("RemoveTags error: " + err.Error())
		// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		// 		return
		// 	}
		// 	if !exist {
		// 		serv.Warnf("Schema does not exist")
		// 		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tags, schema does not exist"})
		// 		return
		// 	}
		// 	_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID},
		// 		bson.M{"$pull": bson.M{"schemas": schema.ID}})
		// 	if err != nil {
		// 		serv.Errorf("RemoveTags error: " + err.Error())
		// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		// 		return
		// 	}
		// case "user":
		// 	exist, user, err := IsUserExist(body.EntityName)
		// 	if err != nil {
		// 		serv.Errorf("RemoveTags error: " + err.Error())
		// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		// 		return
		// 	}
		// 	if !exist {
		// 		serv.Warnf("User with does not exist")
		// 		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tags, user does not exist"})
		//
		// 		return
		// 	}
		// 	_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID},
		// 		bson.M{"$pull": bson.M{"users": user.ID}})
		// 	if err != nil {
		// 		serv.Errorf("RemoveTags error: " + err.Error())
		// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		// 		return
		// 	}
		default:
			serv.Warnf("RemoveTags wrong input")
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tags, wrong input"})
			return
		}
		checkIfEmptyAndDelete(name)
	}
	c.IndentedJSON(200, []string{})
}

func (th TagsHandler) GetTagsByStation(station_id primitive.ObjectID) ([]models.Tag, error) {
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
	from := strings.ToLower(body.EntityType)
	var tags []models.Tag
	switch from {
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
