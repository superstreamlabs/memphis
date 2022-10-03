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
	"errors"
	"memphis-broker/models"
	"memphis-broker/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TagsHandler struct{ S *Server }

const (
	tagObjectName = "Tag"
)

func validateTagName(name string) error {
	return validateName(name, tagObjectName)
}

func CreateTag(name string, from string, from_name string, background_color string, text_color string) {
	name = strings.ToLower(name)
	err := validateTagName(name)
	if err != nil {
		serv.Warnf("failed creating a tag: %v", err.Error())
		return
	}
	exist, _, err := IsTagExist(name)
	if err != nil {
		serv.Warnf("failed creating a tag: %v", err.Error())
		return
	}
	if exist {
		serv.Warnf("Tag with that name already exists")
		return
	}
	var newTag models.Tag
	stationArr := []primitive.ObjectID{}
	schemaArr := []primitive.ObjectID{}
	userArr := []primitive.ObjectID{}
	switch from {
	case "station":
		exist, station, err := IsStationExist(from_name)
		if err != nil {
			serv.Warnf("failed creating a tag: %v", err.Error())
			return
		}
		if !exist {
			serv.Warnf("Station with this name does not exist")
			return
		}
		stationArr = append(stationArr, station.ID)
	case "schema":
		exist, schema, err := IsSchemaExist(from_name)
		if err != nil {
			serv.Warnf("failed creating a tag: %v", err.Error())
			return
		}
		if !exist {
			serv.Warnf("Schema with this name does not exist")
			return
		}
		schemaArr = append(schemaArr, schema.ID)
	case "user":
		exist, user, err := IsUserExist(from_name)
		if err != nil {
			serv.Warnf("failed creating a tag: %v", err.Error())
			return
		}
		if !exist {
			serv.Warnf("User with this name does not exist")
			return
		}
		userArr = append(userArr, user.ID)
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
	_, err = tagsCollection.InsertOne(context.TODO(), newTag)
	if err != nil {
		serv.Warnf("failed creating a tag: %v", err.Error())
		return
	}
}

func AddTag(name string, to string, to_name string, background_color string, text_color string) {
	name = strings.ToLower(name)
	err := validateTagName(name)
	if err != nil {
		serv.Warnf("failed creating a tag: %v", err.Error())
		return
	}
	exist, tag, err := IsTagExist(name)
	if err != nil {
		serv.Warnf("failed creating a tag: %v", err.Error())
		return
	}
	if !exist {
		CreateTag(name, to, to_name, background_color, text_color)
		return
	}
	switch to {
	case "station":
		exist, station, err := IsStationExist(to_name)
		if err != nil {
			serv.Warnf("failed creating a tag: %v", err.Error())
			return
		}
		if !exist {
			serv.Warnf("Station with this name does not exist")
			return
		}
		_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID}, bson.M{"$addToSet": bson.M{"stations": station.ID}})
		if err != nil {
			serv.Warnf("failed adding tag: %v", err.Error())
			return
		}

	case "schema":
		exist, schema, err := IsSchemaExist(to_name)
		if err != nil {
			serv.Warnf("failed creating a tag: %v", err.Error())
			return
		}
		if !exist {
			serv.Warnf("Schema with this name does not exist")
			return
		}
		_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID}, bson.M{"$addToSet": bson.M{"schemas": schema.ID}})
		if err != nil {
			serv.Warnf("failed adding tag: %v", err.Error())
			return
		}
	case "user":
		exist, user, err := IsUserExist(to_name)
		if err != nil {
			serv.Warnf("failed creating a tag: %v", err.Error())
			return
		}
		if !exist {
			serv.Warnf("User with this name does not exist")
			return
		}
		_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID}, bson.M{"$addToSet": bson.M{"users": user.ID}})
		if err != nil {
			serv.Warnf("failed adding tag: %v", err.Error())
			return
		}
	}
}

func DeleteTagsByStation(station_name string) {
	exist, station, err := IsStationExist(station_name)
	if err != nil {
		serv.Warnf("failed deleting tags: %v", err.Error())
		return
	}
	if !exist {
		serv.Warnf("Station with this name does not exist")
		return
	}
	_, err = tagsCollection.UpdateMany(context.TODO(), bson.M{}, bson.M{"$pull": bson.M{"stations": station.ID}})
	if err != nil {
		serv.Warnf("failed deleting tags: %v", err.Error())
		return
	}
}

func DeleteTagsBySchema(schema_name string) {
	exist, schema, err := IsSchemaExist(schema_name)
	if err != nil {
		serv.Warnf("failed deleting tags: %v", err.Error())
		return
	}
	if !exist {
		serv.Warnf("Schema with this name does not exist")
		return
	}
	_, err = tagsCollection.UpdateMany(context.TODO(), bson.M{}, bson.M{"$pull": bson.M{"schemas": schema.ID}})
	if err != nil {
		serv.Warnf("failed deleting tags: %v", err.Error())
		return
	}
}

func DeleteTagsByUser(user_name string) {
	exist, user, err := IsUserExist(user_name)
	if err != nil {
		serv.Warnf("failed deleting tags: %v", err.Error())
		return
	}
	if !exist {
		serv.Warnf("User with this name does not exist")
		return
	}
	_, err = tagsCollection.UpdateMany(context.TODO(), bson.M{}, bson.M{"$pull": bson.M{"users": user.ID}})
	if err != nil {
		serv.Warnf("failed deleting tags: %v", err.Error())
		return
	}
}

func checkIfEmptyAndDelete(name string) error {
	exist, tag, err := IsTagExist(name)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("Tag with this name does not exist")
	}
	if len(tag.Schemas) == 0 && len(tag.Stations) == 0 && len(tag.Users) == 0 {
		_, err = tagsCollection.DeleteOne(context.TODO(), bson.M{"_id": tag.ID})
		if err != nil {
			return err
		}
	}
	return nil
}

func (th TagsHandler) CreateTag(c *gin.Context) {
	var body models.CreateTagSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	name := strings.ToLower(body.Name)
	err := validateTagName(name)
	if err != nil {
		serv.Warnf("failed creating tag: %v", err.Error())
		return
	}
}

func (th TagsHandler) RemoveTag(c *gin.Context) {
	var body models.RemoveTagSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	name := strings.ToLower(body.Name)
	err := validateTagName(name)
	if err != nil {
		serv.Errorf("RemoveTag error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	exist, tag, err := IsTagExist(name)
	if err != nil {
		serv.Errorf("RemoveTag error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Errorf("Tag with this name does not exist")
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	switch body.From {
	case "station":
		exist, station, err := IsStationExist(body.FromName)
		if err != nil {
			serv.Errorf("RemoveTag error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			serv.Errorf("Station with this name does not exist")
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
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
		exist, schema, err := IsSchemaExist(body.FromName)
		if err != nil {
			serv.Errorf("RemoveTag error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			serv.Errorf("Schema with this name does not exist")
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID},
			bson.M{"$pull": bson.M{"schemas": schema.ID}})
		if err != nil {
			serv.Errorf("RemoveTag error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	case "user":
		exist, user, err := IsUserExist(body.FromName)
		if err != nil {
			serv.Errorf("RemoveTag error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			serv.Errorf("User with this name does not exist")
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		_, err = tagsCollection.UpdateOne(context.TODO(), bson.M{"_id": tag.ID},
			bson.M{"$pull": bson.M{"users": user.ID}})
		if err != nil {
			serv.Errorf("RemoveTag error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	default:
		serv.Errorf("RemoveTag error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	err = checkIfEmptyAndDelete(name)
	if err != nil {
		serv.Warnf("failed deleting tag: %v", err.Error())
		return
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

func (th TagsHandler) GetAllTags(c *gin.Context) {
	var tags []models.Tag
	cursor, err := tagsCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		serv.Errorf("GetAllTags error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &tags); err != nil {
		serv.Errorf("GetAllTags error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(tags) == 0 {
		tags = []models.Tag{}
	}

	c.IndentedJSON(200, tags)
}
