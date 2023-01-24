// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"context"
	"encoding/json"
	"strings"

	"memphis-broker/analytics"
	"memphis-broker/models"
	"memphis-broker/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const sendNotificationType = "send_notification"

type IntegrationsHandler struct{ S *Server }

func (it IntegrationsHandler) CreateIntegration(c *gin.Context) {
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}

	var body models.CreateIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	var integration models.Integration
	integrationType := strings.ToLower(body.Name)
	switch integrationType {
	case "slack":
		_, _, slackIntegration, errorCode, err := it.handleCreateSlackIntegration(integrationType, body)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("CreateSlackIntegration: " + err.Error())
			} else {
				serv.Warnf("CreateSlackIntegration: " + err.Error())
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": err.Error()})
			return
		}
		integration = slackIntegration
	case "s3":
		s3Integration, errorCode, err := it.handleCreateS3Integration(body.Keys, "s3")
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("CreateS3Integration: " + err.Error())
			} else {
				serv.Warnf("CreateS3Integration: " + err.Error())
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": err.Error()})
			return
		}
		integration = s3Integration
	default:
		serv.Warnf("CreateIntegration: Unsupported integration type")
		c.AbortWithStatusJSON(400, gin.H{"message": "CreateIntegration error: Unsupported integration type"})
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
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}

	var body models.CreateIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	var integration models.Integration
	switch strings.ToLower(body.Name) {
	case "slack":
		slackIntegration, errorCode, err := it.handleUpdateSlackIntegration("slack", body)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("UpdateSlackIntegration: " + err.Error())
			} else {
				serv.Warnf("UpdateSlackIntegration: " + err.Error())
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": err.Error()})
			return
		}
		integration = slackIntegration
	case "s3":
		s3Integration, errorCode, err := it.handleUpdateS3Integration(body)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("UpdateS3Integration: " + err.Error())
			} else {
				serv.Warnf("UpdateS3Integration: " + err.Error())
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": err.Error()})
			return
		}
		integration = s3Integration

	default:
		serv.Warnf("UpdateIntegration: Unsupported integration type - " + body.Name)
		c.AbortWithStatusJSON(400, gin.H{"message": "Unsupported integration type - " + body.Name})
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
	filter := bson.M{"name": strings.ToLower(body.Name)}
	var integration models.Integration
	err := integrationsCollection.FindOne(context.TODO(),
		filter).Decode(&integration)
	if err == mongo.ErrNoDocuments {
		c.IndentedJSON(200, nil)
		return
	} else if err != nil {
		serv.Errorf("GetIntegrationDetails: Integration " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
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
	var integrations []models.Integration
	cursor, err := integrationsCollection.Find(context.TODO(), bson.M{})
	if err == mongo.ErrNoDocuments {
	} else if err != nil {
		serv.Errorf("GetAllIntegrations: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if err = cursor.All(context.TODO(), &integrations); err != nil {
		serv.Errorf("GetAllIntegrations: " + err.Error())
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
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}

	var body models.DisconnectIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	integrationType := strings.ToLower(body.Name)
	filter := bson.M{"name": integrationType}
	_, err := integrationsCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		serv.Errorf("DisconnectIntegration: Integration " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	integrationUpdate := models.Integration{
		Name:       strings.ToLower(body.Name),
		Keys:       nil,
		Properties: nil,
	}

	msg, err := json.Marshal(integrationUpdate)
	if err != nil {
		serv.Errorf("DisconnectIntegration: Integration " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		serv.Errorf("DisconnectIntegration: Integration " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	switch body.Name {
	case "slack":
		update := models.ConfigurationsUpdate{
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
