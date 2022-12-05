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
	"encoding/json"
	"errors"
	"strings"

	"memphis-broker/analytics"
	"memphis-broker/db"
	"memphis-broker/models"
	"memphis-broker/notifications"
	"memphis-broker/utils"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

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
		var authToken, channelID, uiUrl string
		var pmAlert, svfAlert, disconnectAlert bool
		authToken, ok := body.Keys["auth_token"]
		if !ok {
			serv.Warnf("CreateIntegration error: Must provide auth token for slack integration")
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Must provide auth token for slack integration"})
		}
		channelID, ok = body.Keys["channel_id"]
		if !ok {
			if !ok {
				serv.Warnf("CreateIntegration error: Must provide channel ID for slack integration")
				c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Must provide channel ID for slack integration"})
			}
		}
		uiUrl = body.UIUrl
		if uiUrl == "" {
			serv.Warnf("CreateIntegration error: Must provide channel ID for slack integration")
			c.AbortWithStatusJSON(500, gin.H{"message": "Must provide UI url for slack integration"})
		}

		pmAlert, ok = body.Properties[notifications.PoisonMAlert]
		if !ok {
			pmAlert = false
		}
		svfAlert, ok = body.Properties[notifications.SchemaVAlert]
		if !ok {
			svfAlert = false
		}
		disconnectAlert, ok = body.Properties[notifications.DisconEAlert]
		if !ok {
			disconnectAlert = false
		}

		slackIntegration, err := createSlackIntegration(authToken, channelID, pmAlert, svfAlert, disconnectAlert, body.UIUrl)
		if err != nil {
			if strings.Contains(err.Error(), "Invalid auth token") || strings.Contains(err.Error(), "Invalid channel ID") || strings.Contains(err.Error(), "already exists") {
				serv.Warnf("CreateSlackIntegration error: " + err.Error())
				c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
				return
			} else {
				serv.Errorf("CreateSlackIntegration error: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
		}
		integration = slackIntegration
		if integration.Keys["auth_token"] != "" {
			integration.Keys["auth_token"] = "xoxb-****"
		}
	default:
		serv.Warnf("CreateIntegration error: Unsupported integration type")
		c.AbortWithStatusJSON(400, gin.H{"message": "CreateIntegration error: Unsupported integration type"})
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
		var authToken, channelID, uiUrl string
		var pmAlert, svfAlert, disconnectAlert bool
		authToken, ok := body.Keys["auth_token"]
		if !ok {
			serv.Warnf("CreateIntegration error: Must provide auth token for slack integration")
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Must provide auth token for slack integration"})
		}
		channelID, ok = body.Keys["channel_id"]
		if !ok {
			if !ok {
				serv.Warnf("CreateIntegration error: Must provide channel ID for slack integration")
				c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Must provide channel ID for slack integration"})
			}
		}
		uiUrl = body.UIUrl
		if uiUrl == "" {
			serv.Warnf("CreateIntegration error: Must provide channel ID for slack integration")
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Must provide channel ID for slack integration"})
		}
		pmAlert, ok = body.Properties[notifications.PoisonMAlert]
		if !ok {
			pmAlert = false
		}
		svfAlert, ok = body.Properties[notifications.SchemaVAlert]
		if !ok {
			svfAlert = false
		}
		disconnectAlert, ok = body.Properties[notifications.DisconEAlert]
		if !ok {
			disconnectAlert = false
		}

		slackIntegration, err := updateSlackIntegration(authToken, channelID, pmAlert, svfAlert, disconnectAlert, body.UIUrl)
		if err != nil {
			if strings.Contains(err.Error(), "Invalid auth token") || strings.Contains(err.Error(), "Invalid channel ID") {
				serv.Warnf("UpdateSlackIntegration error: " + err.Error())
				c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
				return
			} else {
				serv.Errorf("UpdateSlackIntegration error: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
		}
		integration = slackIntegration
		if integration.Keys["auth_token"] != "" {
			integration.Keys["auth_token"] = "xoxb-****"
		}
	default:
		serv.Warnf("CreateIntegration error: Unsupported integration type")
		c.AbortWithStatusJSON(400, gin.H{"message": "CreateIntegration error: Unsupported integration type"})
	}

	c.IndentedJSON(200, integration)
}

func createSlackIntegration(authToken string, channelID string, pmAlert bool, svfAlert bool, disconnectAlert bool, uiUrl string) (models.Integration, error) {
	var slackIntegration models.Integration
	filter := bson.M{"name": "slack"}
	err := integrationsCollection.FindOne(context.TODO(),
		filter).Decode(&slackIntegration)
	if err == mongo.ErrNoDocuments {
		err := testSlackIntegration(authToken, channelID, "Slack integration with Memphis was added successfully")
		if err != nil {
			return slackIntegration, err
		}
		keys, properties := createSlackKeysAndProperties(authToken, channelID, pmAlert, svfAlert, disconnectAlert, uiUrl)
		slackIntegration = models.Integration{
			ID:         primitive.NewObjectID(),
			Name:       "slack",
			Keys:       keys,
			Properties: properties,
		}
		_, insertErr := integrationsCollection.InsertOne(context.TODO(), slackIntegration)
		if insertErr != nil {
			return slackIntegration, insertErr
		}

		integrationToUpdate := models.CreateIntegrationSchema{
			Name:       "slack",
			Keys:       keys,
			Properties: properties,
			UIUrl:      uiUrl,
		}
		msg, err := json.Marshal(integrationToUpdate)
		if err != nil {
			return slackIntegration, err
		}
		err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
		if err != nil {
			return slackIntegration, err
		}

		return slackIntegration, nil
	} else if err != nil {
		return slackIntegration, err
	}
	return slackIntegration, errors.New("Slack integration already exists")
}

func updateSlackIntegration(authToken string, channelID string, pmAlert bool, svfAlert bool, disconnectAlert bool, uiUrl string) (models.Integration, error) {
	var slackIntegration models.Integration
	if authToken == "" {
		var integrationFromDb models.Integration
		filter := bson.M{"name": "slack"}
		err := integrationsCollection.FindOne(context.TODO(), filter).Decode(&integrationFromDb)
		if err != nil {
			return slackIntegration, err
		}
		authToken = integrationFromDb.Keys["auth_token"]
	}

	err := testSlackIntegration(authToken, channelID, "Slack integration with Memphis was updated successfully")
	if err != nil {
		return slackIntegration, err
	}
	keys, properties := createSlackKeysAndProperties(authToken, channelID, pmAlert, svfAlert, disconnectAlert, uiUrl)
	filter := bson.M{"name": "slack"}
	err = integrationsCollection.FindOneAndUpdate(context.TODO(),
		filter,
		bson.M{"$set": bson.M{"keys": keys, "properties": properties}}).Decode(&slackIntegration)
	if err == mongo.ErrNoDocuments {
		slackIntegration = models.Integration{
			ID:         primitive.NewObjectID(),
			Name:       "slack",
			Keys:       keys,
			Properties: properties,
		}
		integrationsCollection.InsertOne(context.TODO(), slackIntegration)
	} else if err != nil {
		return slackIntegration, err
	}

	integrationToUpdate := models.CreateIntegrationSchema{
		Name:       "slack",
		Keys:       keys,
		Properties: properties,
		UIUrl:      uiUrl,
	}

	msg, err := json.Marshal(integrationToUpdate)
	if err != nil {
		return slackIntegration, err
	}
	err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		return slackIntegration, err
	}

	slackIntegration.Keys = keys
	slackIntegration.Properties = properties
	return slackIntegration, nil
}

func testSlackIntegration(authToken string, channelID string, message string) error {
	slackClientTemp := slack.New(authToken)
	_, err := slackClientTemp.AuthTest()
	if err != nil {
		return errors.New("Invalid auth token")
	}
	attachment := slack.Attachment{
		AuthorName: "Memphis",
		Text:       message,
		Color:      "#6557FF",
	}

	_, _, err = slackClientTemp.PostMessage(
		channelID,
		slack.MsgOptionAttachments(attachment),
	)
	if err != nil {
		return errors.New("Invalid channel ID")
	}
	return nil
}

func createSlackKeysAndProperties(authToken string, channelID string, pmAlert bool, svfAlert bool, disconnectAlert bool, uiUrl string) (map[string]string, map[string]bool) {
	keys := make(map[string]string)
	keys["auth_token"] = authToken
	keys["channel_id"] = channelID
	properties := make(map[string]bool)
	properties[notifications.PoisonMAlert] = pmAlert
	properties[notifications.SchemaVAlert] = svfAlert
	properties[notifications.DisconEAlert] = disconnectAlert
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
		serv.Errorf("GetIntegrationDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if integration.Name == "slack" && integration.Keys["auth_token"] != "" {
		integration.Keys["auth_token"] = "xoxb-****"
	}
	c.IndentedJSON(200, integration)
}

func (it IntegrationsHandler) GetAllIntegrations(c *gin.Context) {
	var integrations []models.Integration
	cursor, err := integrationsCollection.Find(context.TODO(), bson.M{})
	if err == mongo.ErrNoDocuments {
	} else if err != nil {
		serv.Errorf("GetAllIntegrations error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if err = cursor.All(context.TODO(), &integrations); err != nil {
		serv.Errorf("GetAllIntegrations error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	for i := 0; i < len(integrations); i++ {
		if integrations[i].Name == "slack" && integrations[i].Keys["auth_token"] != "" {
			integrations[i].Keys["auth_token"] = "xoxb-****"
		}
	}
	c.IndentedJSON(200, integrations)
}

func (it IntegrationsHandler) DisconnectIntegration(c *gin.Context) {
	var body models.DisconnectIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	integrationType := strings.ToLower(body.Name)
	filter := bson.M{"name": integrationType}
	_, err := integrationsCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		serv.Errorf("DisconnectIntegration error: " + err.Error())
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
		serv.Errorf("DisconnectIntegration error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		serv.Errorf("DisconnectIntegration error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-disconnect-integration-"+integrationType)
	}
	c.IndentedJSON(200, gin.H{})
}

func InitializeIntegrations(c *mongo.Client) error {
	notifications.IntegrationsCollection = db.GetCollection("integrations", c)
	notifications.NotificationIntegrationsMap = make(map[string]interface{})
	notifications.NotificationFunctionsMap = make(map[string]interface{})
	notifications.NotificationFunctionsMap["slack"] = notifications.SendMessageToSlackChannel
	err := notifications.InitializeSlackConnection(c)
	if err != nil {
		return err
	}
	return nil
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
