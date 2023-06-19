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
	"encoding/json"
	"fmt"
	"strings"

	"memphis/analytics"
	"memphis/conf"
	"memphis/db"
	"memphis/models"
	"memphis/utils"

	"github.com/gin-gonic/gin"
)

const sendNotificationType = "send_notification"

type IntegrationsHandler struct{ S *Server }

func (it IntegrationsHandler) CreateIntegration(c *gin.Context) {
	var message string
	var body models.CreateIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("[tenant name: %v]CreateIntegration: %v", body.TenantName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if body.TenantName == "" {
		body.TenantName = user.TenantName
	}

	exist, _, err := db.GetTenantByName(body.TenantName)
	if err != nil {
		serv.Errorf("[tenant name: %v][user name: %v]CreateIntegration: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("[tenant name: %v][user name: %v]CreateIntegration : tenant %v does not exist", user.TenantName, user.Username, body.TenantName)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	var integration models.Integration
	integrationType := strings.ToLower(body.Name)
	switch integrationType {
	case "slack":
		_, _, slackIntegration, errorCode, err := it.handleCreateSlackIntegration(body)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("[tenant name: %v][user name: %v]CreateSlackIntegration: %v", user.TenantName, user.Username, err.Error())
				message = "Server error"
			} else {
				serv.Warnf("[tenant name: %v][user name: %v]CreateSlackIntegration: %v", user.TenantName, user.Username, err.Error())
				message = err.Error()
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": message})
			return
		}
		integration = slackIntegration
	case "s3":
		s3Integration, errorCode, err := it.handleCreateS3Integration(body.TenantName, body.Keys)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("[tenant name: %v][user name: %v]CreateS3Integration: %v", user.TenantName, user.Username, err.Error())
				message = "Server error"
			} else {
				serv.Warnf("[tenant name: %v][user name: %v]CreateS3Integration: %v", user.TenantName, user.Username, err.Error())
				message = err.Error()
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": message})
			return
		}
		integration = s3Integration
	default:
		serv.Warnf("[tenant name: %v][user name: %v]CreateIntegration: Unsupported integration type - %v", user.TenantName, user.Username, integrationType)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Unsupported integration type - " + integrationType})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-create-integration-"+integrationType)
	}
	c.IndentedJSON(200, integration)
}

func (it IntegrationsHandler) UpdateIntegration(c *gin.Context) {
	var body models.CreateIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	if body.TenantName == "" {
		body.TenantName = DEFAULT_GLOBAL_ACCOUNT
	}

	exist, _, err := db.GetTenantByName(body.TenantName)
	if err != nil {
		serv.Errorf("[tenant name: %v]UpdateIntegration: %v", body.TenantName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("[tenant name: %v]UpdateIntegration : tenant %v does not exist", body.TenantName, body.TenantName)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	var integration models.Integration
	var message string
	switch strings.ToLower(body.Name) {
	case "slack":
		slackIntegration, errorCode, err := it.handleUpdateSlackIntegration("slack", body)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("[tenant name:%v]UpdateSlackIntegration: %v", body.TenantName, err.Error())
				message = "Server error"
			} else {
				serv.Warnf("[tenant name:%v]UpdateSlackIntegration: %v", body.TenantName, err.Error())
				message = err.Error()
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": message})
			return
		}
		integration = slackIntegration
	case "s3":
		s3Integration, errorCode, err := it.handleUpdateS3Integration(body)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("[tenant name: %v]UpdateS3Integration: %v", body.TenantName, err.Error())
				message = "Server error"
			} else {
				serv.Warnf("[tenant name: %v]UpdateS3Integration: %v", body.TenantName, err.Error())
				message = err.Error()
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": message})
			return
		}
		integration = s3Integration

	default:
		serv.Warnf("[tenant name: %v]UpdateIntegration: Unsupported integration type - %v", body.TenantName, body.Name)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Unsupported integration type - " + body.Name})
		return
	}

	c.IndentedJSON(200, integration)
}

func createIntegrationsKeysAndProperties(integrationType, authToken string, channelID string, pmAlert bool, svfAlert bool, disconnectAlert bool, accessKey, secretKey, bucketName, region string) (map[string]string, map[string]bool) {
	keys := make(map[string]string)
	properties := make(map[string]bool)
	switch integrationType {
	case "slack":
		keys["auth_token"] = authToken
		keys["channel_id"] = channelID
		properties[PoisonMAlert] = pmAlert
		properties[SchemaVAlert] = svfAlert
		properties[DisconEAlert] = disconnectAlert
	case "s3":
		keys["access_key"] = accessKey
		keys["secret_key"] = secretKey
		keys["bucket_name"] = bucketName
		keys["region"] = region
	}

	return keys, properties
}

func (it IntegrationsHandler) GetIntegrationDetails(c *gin.Context) {
	var body models.GetIntegrationDetailsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("[tenant name: %v]GetIntegrationDetails: Integration %v: %v", body.TenantName, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if body.TenantName == "" {
		body.TenantName = user.TenantName
	}

	exist, _, err := db.GetTenantByName(body.TenantName)
	if err != nil {
		serv.Errorf("[tenant name: %v][user name: %v]GetIntegrationDetails: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("[tenant name: %v][user name: %v]GetIntegrationDetails : tenant %v does not exist", user.TenantName, user.Username, body.TenantName)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	exist, integration, err := db.GetIntegration(strings.ToLower(body.Name), body.TenantName)
	if err != nil {
		serv.Errorf("[tenant name: %v][user name: %v]GetIntegrationDetails: Integration %v: %v", body.TenantName, user.Username, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	} else if !exist {
		c.IndentedJSON(200, nil)
		return
	}

	if integration.Name == "slack" && integration.Keys["auth_token"] != "" {
		integration.Keys["auth_token"] = "xoxb-****"
	}

	if integration.Name == "s3" && integration.Keys["secret_key"] != "" {
		lastCharsSecretKey := integration.Keys["secret_key"][len(integration.Keys["secret_key"])-4:]
		integration.Keys["secret_key"] = "****" + lastCharsSecretKey
	}
	c.IndentedJSON(200, integration)
}

func (it IntegrationsHandler) GetAllIntegrations(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		message := fmt.Sprintf("[tenant name: %v][user name: %v]GetAllIntegrations: %v", user.TenantName, user.Username, err.Error())
		serv.Errorf(message)
		c.AbortWithStatusJSON(500, gin.H{"message": message})
		return
	}

	_, integrations, err := db.GetAllIntegrationsByTenant(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant name: %v][user name: %v]GetAllIntegrations: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	for i := 0; i < len(integrations); i++ {
		if integrations[i].Name == "slack" && integrations[i].Keys["auth_token"] != "" {
			integrations[i].Keys["auth_token"] = "xoxb-****"
		}
		if integrations[i].Name == "s3" && integrations[i].Keys["secret_key"] != "" {
			lastCharsSecretKey := integrations[i].Keys["secret_key"][len(integrations[i].Keys["secret_key"])-4:]
			integrations[i].Keys["secret_key"] = "****" + lastCharsSecretKey
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-enter-integration-page")
	}

	c.IndentedJSON(200, integrations)
}

func (it IntegrationsHandler) DisconnectIntegration(c *gin.Context) {
	var body models.DisconnectIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("[tenant name: %v]DisconnectIntegration: Integration %v: %v", body.TenantName, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if body.TenantName == "" {
		body.TenantName = user.TenantName
	}

	exist, _, err := db.GetTenantByName(body.TenantName)
	if err != nil {
		serv.Errorf("[tenant name:%v]DisconnectIntegration: %v", body.TenantName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("[tenant name: %v]DisconnectIntegration : tenant %v does not exist", body.TenantName, body.TenantName)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	integrationType := strings.ToLower(body.Name)
	err = db.DeleteIntegration(integrationType, body.TenantName)
	if err != nil {
		serv.Errorf("[tenant name: %v]DisconnectIntegration: Integration %v: %v", body.TenantName, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if body.TenantName != conf.GlobalAccountName {
		body.TenantName = strings.ToLower(body.TenantName)
	}

	integrationUpdate := models.Integration{
		Name:       strings.ToLower(body.Name),
		Keys:       nil,
		Properties: nil,
		TenantName: body.TenantName,
	}

	msg, err := json.Marshal(integrationUpdate)
	if err != nil {
		serv.Errorf("[tenant name: %v]DisconnectIntegration: Integration %v: %v", body.TenantName, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		serv.Errorf("[tenant name: %v]DisconnectIntegration: Integration %v: %v", body.TenantName, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	switch body.Name {
	case "slack":
		update := models.SdkClientsUpdates{
			Type:   sendNotificationType,
			Update: false,
		}
		serv.SendUpdateToClients(update)
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-disconnect-integration-"+integrationType)
	}
	c.IndentedJSON(200, gin.H{})
}

func (it IntegrationsHandler) RequestIntegration(c *gin.Context) {
	var body models.RequestIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		param := analytics.EventParam{
			Name:  "request-content",
			Value: body.RequestContent,
		}
		analyticsParams := []analytics.EventParam{param}
		analytics.SendEventWithParams(user.Username, analyticsParams, "user-request-integration")
	}

	c.IndentedJSON(200, gin.H{})
}
