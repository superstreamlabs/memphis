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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/memphisdev/memphis/analytics"
	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"
	"github.com/memphisdev/memphis/utils"

	"github.com/gin-gonic/gin"
)

type TagsHandler struct{ S *Server }

func validateEntityType(entity string) error {
	switch entity {
	case "station", "schema", "user":
		return nil
	default:
		return errors.New("entity type is not valid")
	}
}

func CreateTag(name string, entity_type string, entity_id int, color string, tenantName string) error {
	name = strings.ToLower(name)
	entity := strings.ToLower(entity_type)
	stationArr := []int{}
	schemaArr := []int{}
	userArr := []int{}
	switch entity {
	case "station":
		stationArr = append(stationArr, entity_id)
	case "schema":
		schemaArr = append(schemaArr, entity_id)
		// case "user":
		// 	userArr = append(userArr, entity_id)
	}
	_, err := db.InsertNewTag(name, color, stationArr, schemaArr, userArr, tenantName)
	if err != nil {
		return err
	}
	return nil
}

func AddTagsToEntity(tags []models.CreateTag, entity_type string, entity_id int, tenantName, color string) error {
	if len(tags) == 0 {
		return nil
	}
	entity := strings.ToLower(entity_type)
	err := validateEntityType(entity)
	if err != nil {
		return err
	}
	for _, tagToCreate := range tags {
		exist, _, err := db.GetTagByName(tagToCreate.Name, tenantName)
		if err != nil {
			return err
		}
		if !exist {
			err = CreateTag(tagToCreate.Name, entity_type, entity_id, tagToCreate.Color, tenantName)
			if err != nil {
				return err
			}
		} else {
			err = db.InsertEntityToTag(tagToCreate.Name, entity_type, entity_id, tenantName, color)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func DeleteTagsFromStation(id int) {
	err := db.RemoveAllTagsFromEntity("stations", id)
	if err != nil {
		serv.Errorf("DeleteTagsFromStation: Station ID %v: %v", strconv.Itoa(id), err.Error())
		return
	}
}

func DeleteTagsFromSchema(id int) {
	err := db.RemoveAllTagsFromEntity("schemas", id)
	if err != nil {
		serv.Errorf("DeleteTagsFromSchema: Schema ID %v: %v", strconv.Itoa(id), err.Error())
		return
	}
}

func DeleteTagsFromUser(id int) {
	err := db.RemoveAllTagsFromEntity("users", id)
	if err != nil {
		serv.Errorf("DeleteTagsFromUser: User ID %v: %v", strconv.Itoa(id), err.Error())
		return
	}
}

func (th TagsHandler) CreateNewTag(c *gin.Context) {
	var body models.CreateTag
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	tenantName := user.TenantName
	if err != nil {
		serv.Errorf("CreateNewTag at getUserDetailsFromMiddleware: Tag %v: %v", body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	name := strings.ToLower(body.Name)
	exist, _, err := db.GetTagByName(name, tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateNewTag at db.GetTagByName: Tag %v: %v", user.TenantName, user.Username, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		errMsg := fmt.Sprintf("Tag with the name %v already exists", body.Name)
		serv.Warnf("[tenant: %v][user: %v]CreateNewTag: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
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
	newTag, err := db.InsertNewTag(name, color, stationArr, schemaArr, userArr, tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateNewTag at db.InsertNewTag: Tag %v: %v", user.TenantName, user.Username, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	message := fmt.Sprintf("[tenant: %v][user: %v]New Tag %v has been created ", user.TenantName, user.Username, newTag.Name)
	serv.Noticef(message)

	c.IndentedJSON(200, newTag)
}

func (th TagsHandler) RemoveTag(c *gin.Context) {
	var body models.RemoveTagSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveTag at getUserDetailsFromMiddleware: Tag %v: %v", body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	name := strings.ToLower(body.Name)
	entity := strings.ToLower(body.EntityType)
	err = validateEntityType(entity)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]RemoveTag at validateEntityType: Tag %v at %v %v: %v", user.TenantName, user.Username, body.Name, entity, body.EntityName, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	var entity_id int
	var stationName string
	var message string

	tenantName := user.TenantName
	switch entity {
	case "station":
		station_name, err := StationNameFromStr(body.EntityName)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]RemoveTag at StationNameFromStr: Tag %v at %v %v: %v", user.TenantName, user.Username, body.Name, entity, body.EntityName, err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
		exist, station, err := db.GetStationByName(station_name.Ext(), tenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]RemoveTag at GetStationByName: Tag %v: %v", user.TenantName, user.Username, body.Name, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			c.IndentedJSON(200, []string{})
			return
		}
		entity_id = station.ID
		stationName = station_name.Ext()
		message = fmt.Sprintf("Tag %v has been deleted from station %v by user %v", name, stationName, user.Username)

	case "schema":
		exist, schema, err := db.GetSchemaByName(body.EntityName, tenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]RemoveTag at GetSchemaByName: Tag %v: %v", user.TenantName, user.Username, body.Name, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			c.IndentedJSON(200, []string{})
			return
		}
		entity_id = schema.ID
		message = fmt.Sprintf("Tag %v has been deleted from schema %v by user name %v", name, schema.Name, user.Username)

	// case "user":
	// 	exist, user, err := memphis_cache.GetUser(body.EntityName)
	// 	if err != nil {
	// 		serv.Errorf("RemoveTag: Tag " + body.Name + ": " + err.Error())
	// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
	// 		return
	// 	}
	// 	if !exist {
	// 		c.IndentedJSON(200, []string{})
	// 		return
	// 	}
	// 	entity_id = user.ID

	default:
		serv.Warnf("[tenant: %v][user: %v]RemoveTag: Tag %v at %v %v: unsupported entity type", user.TenantName, user.Username, body.Name, entity, body.EntityName)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tag " + body.Name + ", unsupported entity type"})
		return
	}

	err = db.RemoveTagFromEntity(name, entity, entity_id)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RemoveTag: Tag %v: %v", user.TenantName, user.Username, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	serv.Noticef("[tenant: %v][user: %v]: %v", user.TenantName, user.Username, message)
	if entity == "station" {
		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			StationName:       stationName,
			Message:           message,
			CreatedBy:         user.ID,
			CreatedByUsername: user.Username,
			CreatedAt:         time.Now(),
			TenantName:        user.TenantName,
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]RemoveTag: Tag %v at %v %v - create audit logs error: %v", user.TenantName, user.Username, body.Name, entity, body.EntityName, err.Error())
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

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("UpdateTagsForEntity at getUserDetailsFromMiddleware: %v %v: %v", body.EntityType, body.EntityName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	entity := strings.ToLower(body.EntityType)
	err = validateEntityType(entity)
	var entity_id int
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]UpdateTagsForEntity at validateEntityType: %v %v: %v", user.TenantName, user.Username, entity, body.EntityName, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	tenantName := user.TenantName
	var stationName StationName
	var schemaName string
	switch entity {
	case "station":
		station_name, err := StationNameFromStr(body.EntityName)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]UpdateTagsForEntity at StationNameFromStr: %v %v: %v", user.TenantName, user.Username, entity, body.EntityName, err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
		exist, station, err := db.GetStationByName(station_name.Ext(), tenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]UpdateTagsForEntity at db.GetStationByName: Station %v: %v", user.TenantName, user.Username, body.EntityName, err.Error())
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
		exist, schema, err := db.GetSchemaByName(body.EntityName, tenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]UpdateTagsForEntity db.GetSchemaByName: Schema %v: %v", user.TenantName, user.Username, body.EntityName, err.Error())
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
	// 	exist, user, err := memphis_cache.GetUser(body.EntityName)
	// 	if err != nil {
	// 		serv.Errorf("UpdateTagsForEntity: User " + body.EntityName + ": " + err.Error())
	// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
	// 		return
	// 	}
	// 	if !exist {
	// 		c.IndentedJSON(200, []string{})
	// 		return
	// 	}
	// 	entity_id = user.ID

	default:
		serv.Warnf("[tenant: %v][user: %v]UpdateTagsForEntity: %v %v: unsupproted entity type", user.TenantName, user.Username, entity, body.EntityName)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not remove tags, unsupported entity type"})
		return
	}
	var message string

	if len(body.TagsToAdd) > 0 {
		for _, tagToAdd := range body.TagsToAdd {
			name := strings.ToLower(tagToAdd.Name)
			exist, tag, err := db.GetTagByName(name, tenantName)
			if err != nil {
				serv.Errorf("[tenant: %v][user: %v]UpdateTagsForEntity at GetTagByName: %v %v: %v", user.TenantName, user.Username, body.EntityType, body.EntityName, err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			if !exist {
				err = CreateTag(name, body.EntityType, entity_id, tagToAdd.Color, tenantName)
				if err != nil {
					serv.Errorf("[tenant: %v][user: %v]UpdateTagsForEntity at CreateTag: %v %v: %v", user.TenantName, user.Username, body.EntityType, body.EntityName, err.Error())
					c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
					return
				}
			} else {
				err = db.InsertEntityToTag(tag.Name, entity, entity_id, tenantName, "")
				if err != nil {
					serv.Errorf("[tenant: %v][user: %v]UpdateTagsForEntity at db.InsertEntityToTag: %v %v: %v", user.TenantName, user.Username, body.EntityType, body.EntityName, err.Error())
					c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
					return
				}
			}

			analyticsEventName := ""
			analyticsParams := []analytics.EventParam{}
			if entity == "station" {
				message = "Tag " + name + " has been added to station " + stationName.Ext() + " by user " + user.Username

				var auditLogs []interface{}
				newAuditLog := models.AuditLog{
					StationName:       stationName.Intern(),
					Message:           message,
					CreatedBy:         user.ID,
					CreatedByUsername: user.Username,
					CreatedAt:         time.Now(),
					TenantName:        user.TenantName,
				}

				auditLogs = append(auditLogs, newAuditLog)
				err = CreateAuditLogs(auditLogs)
				if err != nil {
					serv.Warnf("[tenant: %v][user: %v]UpdateTagsForEntity: %v %v - create audit logs error: %v", user.TenantName, user.Username, entity, body.EntityName, err.Error())
				}

				analyticsEventName = "user-tag-station"
				param := analytics.EventParam{
					Name:  "station-name",
					Value: stationName.Ext(),
				}
				analyticsParams = append(analyticsParams, param)
			} else if entity == "schema" {
				message = "Tag " + name + " has been added to schema " + schemaName + " by user " + user.Username
				analyticsEventName = "user-tag-schema"
				param := analytics.EventParam{
					Name:  "schema-name",
					Value: schemaName,
				}
				analyticsParams = append(analyticsParams, param)
			} else {
				message = "Tag " + name + " has been added to user " + "by user " + user.Username
				analyticsEventName = "user-tag-user"
				param := analytics.EventParam{
					Name:  "username",
					Value: user.Username,
				}
				analyticsParams = append(analyticsParams, param)
			}

			shouldSendAnalytics, _ := shouldSendAnalytics()
			if shouldSendAnalytics {
				analyticsParams := map[string]interface{}{"tag-name": name}
				analytics.SendEvent(user.TenantName, user.Username, analyticsParams, analyticsEventName)
			}

			serv.Noticef("[tenant: %v][user: %v] %v", user.TenantName, user.Username, message)
		}
	}
	if len(body.TagsToRemove) > 0 {
		for _, tagToRemove := range body.TagsToRemove {
			name := strings.ToLower(tagToRemove)
			exist, tag, err := db.GetTagByName(name, tenantName)
			if err != nil {
				serv.Errorf("[tenant: %v][user: %v]UpdateTagsForEntity at GetTagByName: %v %v: %v", user.TenantName, user.Username, body.EntityType, body.EntityName, err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			if exist {
				err = db.InsertEntityToTag(tag.Name, entity, entity_id, tenantName, "")
				if err != nil {
					serv.Errorf("[tenant: %v][user: %v]UpdateTagsForEntity at InsertEntityToTag: %v %v: %v", user.TenantName, user.Username, body.EntityType, body.EntityName, err.Error())
					c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
					return
				}
			}
			if entity == "station" {
				message = "Tag " + name + " has been deletd from station " + stationName.Ext() + " by user " + user.Username

				var auditLogs []interface{}
				newAuditLog := models.AuditLog{
					StationName:       stationName.Intern(),
					Message:           message,
					CreatedBy:         user.ID,
					CreatedByUsername: user.Username,
					CreatedAt:         time.Now(),
					TenantName:        user.TenantName,
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
			serv.Noticef("[tenant: %v][user: %v] %v", user.TenantName, user.Username, message)
		}
	}
	tags, err := th.GetTagsByEntityWithID(entity, entity_id)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]UpdateTagsForEntity at GetTagsByEntityWithID: %v %v: %v", user.TenantName, user.Username, entity, body.EntityName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, tags)
}
func (th TagsHandler) GetTagsByEntityWithID(entity string, id int) ([]models.CreateTag, error) {
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

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetTags: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	entity := strings.ToLower(body.EntityType)
	if entity != "" {
		err := validateEntityType(entity)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]GetTags at validateEntityType: %v: %v", user.TenantName, user.Username, body.EntityType, err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	}

	tags, err := db.GetTagsByEntityType(entity, user.TenantName)
	if err != nil {
		desc := ""
		if entity == "" {
			desc = "All Tags"
		} else {
			desc = entity
		}
		serv.Errorf("[tenant: %v][user: %v]GetTags at GetTagsByEntityType: %v: %v", user.TenantName, user.Username, desc, err.Error())
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
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetUsedTags: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	tags, err := db.GetAllUsedTags(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetUsedTags at db.GetAllUsedTags: %v", user.TenantName, user.Username, err.Error())
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
