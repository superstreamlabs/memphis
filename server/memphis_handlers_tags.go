// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"errors"
	"memphis/analytics"
	"memphis/db"
	"memphis/models"
	"memphis/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	_, err := db.UpsertNewTag(name, color, stationArr, schemaArr, userArr)
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
	for _, tagToCreate := range tags {
		exist, _, err := db.GetTagByName(tagToCreate.Name)
		if err != nil {
			return err
		}
		if !exist {
			err = CreateTag(tagToCreate.Name, entity_type, entity_id, tagToCreate.Color)
			if err != nil {
				return err
			}
		} else {
			err = db.UpsertEntityToTag(tagToCreate.Name, entity_type, entity_id)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func DeleteTagsFromStation(id primitive.ObjectID) {
	err := db.RemoveAllTagsFromEntity("stations", id)
	if err != nil {
		serv.Errorf("DeleteTagsFromStation: Station ID " + id.Hex() + ": " + err.Error())
		return
	}
}

func DeleteTagsFromSchema(id primitive.ObjectID) {
	err := db.RemoveAllTagsFromEntity("schemas", id)
	if err != nil {
		serv.Errorf("DeleteTagsFromSchema: Schema ID " + id.Hex() + ": " + err.Error())
		return
	}
}

func DeleteTagsFromUser(id primitive.ObjectID) {
	err := db.RemoveAllTagsFromEntity("users", id)
	if err != nil {
		serv.Errorf("DeleteTagsFromUser: User ID " + id.Hex() + ": " + err.Error())
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
	exist, _, err := db.GetTagByName(name)
	if err != nil {
		serv.Errorf("CreateNewTag: Tag " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		errMsg := "Tag with the name " + body.Name + " already exists"
		serv.Warnf("CreateNewTag: " + errMsg)
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}
	var color string
	if len(body.Color) > 0 {
		color = body.Color
	} else {
		color = "101, 87, 255" // default memphis-purple color
	}
	stationArr := []int{}
	schemaArr := []int{}
	userArr := []int{}
	newTag, err := db.UpsertNewTagV1(name, color, stationArr, schemaArr, userArr)
	if err != nil {
		serv.Errorf("CreateNewTag: Tag " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("CreateNewTag: Tag " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}
	message := "New Tag " + newTag.Name + " has been created " + " by user " + user.Username
	serv.Noticef(message)

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
		serv.Warnf("RemoveTag: Tag " + body.Name + " at " + entity + " " + body.EntityName + ": " + err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	var entity_id primitive.ObjectID
	var stationName string
	var message string

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveTag: Tag " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}
	switch entity {
	case "station":
		station_name, err := StationNameFromStr(body.EntityName)
		if err != nil {
			serv.Warnf("RemoveTag: Tag " + body.Name + " at " + entity + " " + body.EntityName + ": " + err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
		exist, station, err := db.GetStationByName(station_name.Ext())
		if err != nil {
			serv.Errorf("RemoveTag: Tag " + body.Name + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			c.IndentedJSON(200, []string{})
			return
		}
		entity_id = station.ID
		stationName = station_name.Ext()
		message = "Tag " + name + " has been deleted from station " + stationName + " by user " + user.Username

	case "schema":
		exist, schema, err := db.GetSchemaByName(body.EntityName)
		if err != nil {
			serv.Errorf("RemoveTag: Tag " + body.Name + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			c.IndentedJSON(200, []string{})
			return
		}
		entity_id = schema.ID
		message = "Tag " + name + " has been deleted from schema" + schema.Name + " by user " + user.Username

	// case "user":
	// 	exist, user, err := db.GetUserByUsername(body.EntityName)
	// 	if err != nil {
	// 		serv.Errorf("RemoveTag: Tag " + body.Name + ": " + err.Error())
	// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
	// 		return
	// 	}
	// 	if !exist {
	// 		c.IndentedJSON(200, []string{})
	// 		return
	// 	}
	// entity_id = user.ID

	default:
		serv.Warnf("RemoveTag: Tag " + body.Name + " at " + entity + " " + body.EntityName + ": unsupported entity type")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tag " + body.Name + ", unsupported entity type"})
		return
	}

	err = db.RemoveTagFromEntity(name, entity, entity_id)
	if err != nil {
		serv.Errorf("RemoveTag: Tag " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	serv.Noticef(message)
	if entity == "station" {
		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			ID:            primitive.NewObjectID(),
			StationName:   stationName,
			Message:       message,
			CreatedByUser: user.Username,
			CreationDate:  time.Now(),
			UserType:      user.UserType,
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			serv.Warnf("RemoveTag: Tag " + body.Name + " at " + entity + " " + body.EntityName + " - create audit logs error: " + err.Error())
		}
	}

	c.IndentedJSON(200, []string{})
}

func (th TagsHandler) UpdateTagsForEntity(c *gin.Context) {
	var body models.UpdateTagsForEntitySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	entity := strings.ToLower(body.EntityType)
	err := validateEntityType(entity)
	var entity_id primitive.ObjectID
	if err != nil {
		serv.Warnf("UpdateTagsForEntity: " + entity + " " + body.EntityName + ": " + err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	var stationName StationName
	var schemaName string
	switch entity {
	case "station":
		station_name, err := StationNameFromStr(body.EntityName)
		if err != nil {
			serv.Warnf("UpdateTagsForEntity: " + entity + " " + body.EntityName + ": " + err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
		exist, station, err := db.GetStationByName(station_name.Ext())
		if err != nil {
			serv.Errorf("UpdateTagsForEntity: Station " + body.EntityName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			c.IndentedJSON(200, []string{})
			return
		}
		entity_id = station.ID
		stationName = station_name

	case "schema":
		exist, schema, err := db.GetSchemaByName(body.EntityName)
		if err != nil {
			serv.Errorf("UpdateTagsForEntity: Schema " + body.EntityName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			c.IndentedJSON(200, []string{})
			return
		}
		entity_id = schema.ID
		schemaName = schema.Name

	// case "user":
	// 	exist, user, err := db.GetUserByUsername(body.EntityName)
	// 	if err != nil {
	// 		serv.Errorf("UpdateTagsForEntity: User " + body.EntityName + ": " + err.Error())
	// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
	// 		return
	// 	}
	// 	if !exist {
	// 		c.IndentedJSON(200, []string{})
	// 		return
	// 	}
	// entity_id = user.ID

	default:
		serv.Warnf("UpdateTagsForEntity: " + entity + " " + body.EntityName + ": unsupported entity type")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tags, unsupported entity type"})
		return
	}
	var message string
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("UpdateTagsForEntity: " + body.EntityType + " " + body.EntityName + ": " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}

	if len(body.TagsToAdd) > 0 {
		for _, tagToAdd := range body.TagsToAdd {
			name := strings.ToLower(tagToAdd.Name)
			exist, tag, err := db.GetTagByName(name)
			if err != nil {
				serv.Errorf("UpdateTagsForEntity: " + body.EntityType + " " + body.EntityName + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			if !exist {
				err = CreateTag(name, body.EntityType, entity_id, tagToAdd.Color)
				if err != nil {
					serv.Errorf("UpdateTagsForEntity: " + body.EntityType + " " + body.EntityName + ": " + err.Error())
					c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
					return
				}
			} else {
				err = db.UpsertEntityToTag(tag.Name, entity, entity_id)
				if err != nil {
					serv.Errorf("UpdateTagsForEntity: " + body.EntityType + " " + body.EntityName + ": " + err.Error())
					c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
					return
				}
			}

			analyticsEventName := ""
			if entity == "station" {
				message = "Tag " + name + " has been added to station " + stationName.Ext() + " by user " + user.Username

				var auditLogs []interface{}
				newAuditLog := models.AuditLog{
					ID:            primitive.NewObjectID(),
					StationName:   stationName.Intern(),
					Message:       message,
					CreatedByUser: user.Username,
					CreationDate:  time.Now(),
					UserType:      user.UserType,
				}

				auditLogs = append(auditLogs, newAuditLog)
				err = CreateAuditLogs(auditLogs)
				if err != nil {
					serv.Warnf("UpdateTagsForEntity: " + entity + " " + body.EntityName + " - create audit logs error: " + err.Error())
				}

				analyticsEventName = "user-tag-station"
			} else if entity == "schema" {
				message = "Tag " + name + " has been added to schema " + schemaName + " by user " + user.Username
				analyticsEventName = "user-tag-schema"
			} else {
				message = "Tag " + name + " has been added to user " + "by user " + user.Username
				analyticsEventName = "user-tag-user"
			}

			shouldSendAnalytics, _ := shouldSendAnalytics()
			if shouldSendAnalytics {
				param := analytics.EventParam{
					Name:  "tag-name",
					Value: name,
				}
				analyticsParams := []analytics.EventParam{param}
				analytics.SendEventWithParams(user.Username, analyticsParams, analyticsEventName)
			}

			serv.Noticef(message)
		}
	}
	if len(body.TagsToRemove) > 0 {
		for _, tagToRemove := range body.TagsToRemove {
			name := strings.ToLower(tagToRemove)
			exist, tag, err := db.GetTagByName(name)
			if err != nil {
				serv.Errorf("UpdateTagsForEntity: " + body.EntityType + " " + body.EntityName + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			if exist {
				err = db.UpsertEntityToTag(tag.Name, entity, entity_id)
				if err != nil {
					serv.Errorf("UpdateTagsForEntity: " + body.EntityType + " " + body.EntityName + ": " + err.Error())
					c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
					return
				}
			}
			if entity == "station" {
				message = "Tag " + name + " has been deletd from station " + stationName.Ext() + " by user " + user.Username

				var auditLogs []interface{}
				newAuditLog := models.AuditLog{
					ID:            primitive.NewObjectID(),
					StationName:   stationName.Intern(),
					Message:       message,
					CreatedByUser: user.Username,
					CreationDate:  time.Now(),
					UserType:      user.UserType,
				}

				auditLogs = append(auditLogs, newAuditLog)
				err = CreateAuditLogs(auditLogs)
				if err != nil {
					serv.Warnf("UpdateTagsForEntity: " + entity + " " + body.EntityName + " - create audit logs error: " + err.Error())
				}
			} else if entity == "schema" {
				message = "Tag " + name + " has been deleted from schema " + schemaName + " by user " + user.Username
			} else {
				message = "Tag " + name + " has been deleted " + "by user " + user.Username

			}
			serv.Noticef(message)
		}
	}
	tags, err := th.GetTagsByEntityWithID(entity, entity_id)
	if err != nil {
		serv.Errorf("UpdateTagsForEntity: " + entity + " " + body.EntityName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, tags)
}
func (th TagsHandler) GetTagsByEntityWithID(entity string, id primitive.ObjectID) ([]models.CreateTag, error) {
	tags, err := db.GetTagsByEntityID(entity, id)
	if err != nil {
		return []models.CreateTag{}, err
	}
	var tagsRes []models.CreateTag
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
			serv.Warnf("GetTags: " + body.EntityType + ": " + err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}
	tags, err := db.GetTagsByEntityType(entity)
	if err != nil {
		desc := ""
		if entity == "" {
			desc = "All Tags"
		} else {
			desc = entity
		}
		serv.Errorf("GetTags: " + desc + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	var tagsRes []models.CreateTag
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
	tags, err := db.GetAllUsedTags()
	if err != nil {
		serv.Errorf("GetUsedTags: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	var tagsRes []models.CreateTag
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
