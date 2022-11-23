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
	"strings"

	"memphis-broker/db"
	"memphis-broker/models"
	"memphis-broker/notifications"
	"memphis-broker/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IntegrationsHandler struct{ S *Server }

func (it IntegrationsHandler) CreateIntegration(c *gin.Context) {
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
		pmAlert, ok = body.Properties["poison_message_alert"]
		if !ok {
			pmAlert = false
		}
		svfAlert, ok = body.Properties["schema_validation_fail_alert"]
		if !ok {
			svfAlert = false
		}
		disconnectAlert, ok = body.Properties["disconnection_events_alert"]
		if !ok {
			disconnectAlert = false
		}

		slackIntegration, err := createSlackIntegration(authToken, channelID, pmAlert, svfAlert, disconnectAlert, body.UIUrl)
		if err != nil {
			serv.Errorf("CreateSlackIntegration error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		}
		integration = slackIntegration
	}
	c.IndentedJSON(200, integration)
}

func (it IntegrationsHandler) UpdateIntegration(c *gin.Context) {
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
		pmAlert, ok = body.Properties["poison_message_alert"]
		if !ok {
			pmAlert = false
		}
		svfAlert, ok = body.Properties["schema_validation_fail_alert"]
		if !ok {
			svfAlert = false
		}
		disconnectAlert, ok = body.Properties["disconnection_events_alert"]
		if !ok {
			disconnectAlert = false
		}

		slackIntegration, err := updateSlackIntegration(authToken, channelID, pmAlert, svfAlert, disconnectAlert, body.UIUrl)
		if err != nil {
			serv.Errorf("CreateSlackIntegration error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		}
		integration = slackIntegration
	}
	c.IndentedJSON(200, integration)
}

func createSlackIntegration(authToken string, channelID string, pmAlert bool, svfAlert bool, disconnectAlert bool, uiUrl string) (models.Integration, error) {
	keys := make(map[string]string)
	keys["auth_token"] = authToken
	keys["channel_id"] = channelID
	properties := make(map[string]bool)
	properties["poison_message_alert"] = pmAlert
	properties["schema_validation_fail_alert"] = svfAlert
	properties["disconnection_events_alert"] = disconnectAlert
	filter := bson.M{"name": "slack"}
	var slackIntegration models.Integration
	err := integrationsCollection.FindOne(context.TODO(),
		filter).Decode(&slackIntegration)
	if err == mongo.ErrNoDocuments {
		slackIntegration = models.Integration{
			ID:         primitive.NewObjectID(),
			Name:       "slack",
			Keys:       keys,
			Properties: properties,
			UIUrl:      uiUrl,
		}
		integrationsCollection.InsertOne(context.TODO(), slackIntegration)
	} else if err != nil {
		return slackIntegration, err
	}
	msg, err := json.Marshal(slackIntegration)
	if err != nil {
		return slackIntegration, err
	}
	err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		return slackIntegration, err
	}

	return slackIntegration, nil
}

func updateSlackIntegration(authToken string, channelID string, pmAlert bool, svfAlert bool, disconnectAlert bool, uiUrl string) (models.Integration, error) {
	keys := make(map[string]string)
	keys["auth_token"] = authToken
	keys["channel_id"] = channelID
	properties := make(map[string]bool)
	properties["poison_message_alert"] = pmAlert
	properties["schema_validation_fail_alert"] = svfAlert
	properties["disconnection_events_alert"] = disconnectAlert
	filter := bson.M{"name": "slack"}
	var slackIntegration models.Integration
	err := integrationsCollection.FindOneAndUpdate(context.TODO(),
		filter,
		bson.M{"$set": bson.M{"keys": keys, "properties": properties, "ui_url": uiUrl}}).Decode(&slackIntegration)
	if err == mongo.ErrNoDocuments {
		slackIntegration = models.Integration{
			ID:         primitive.NewObjectID(),
			Name:       "slack",
			Keys:       keys,
			Properties: properties,
			UIUrl:      uiUrl,
		}
		integrationsCollection.InsertOne(context.TODO(), slackIntegration)
	} else if err != nil {
		return slackIntegration, err
	}
	msg, err := json.Marshal(slackIntegration)
	if err != nil {
		return slackIntegration, err
	}
	err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		return slackIntegration, err
	}

	return slackIntegration, nil
}

func (it IntegrationsHandler) GetIntegrationDetails(c *gin.Context) {
	var body models.GetIntegrationDetailsRequest
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	filter := bson.M{"name": strings.ToLower(body.Name)}
	var integration models.Integration
	err := integrationsCollection.FindOne(context.TODO(),
		filter).Decode(&integration)
	if err == mongo.ErrNoDocuments {
	} else if err != nil {
		serv.Errorf("GetIntegrationDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
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

	c.IndentedJSON(200, integrations)
}

func (it IntegrationsHandler) DeleteIntegration(c *gin.Context) {
	var body models.DeleteIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	filter := bson.M{"name": strings.ToLower(body.Name)}
	_, err := integrationsCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		serv.Errorf("DeleteIntegration error: " + err.Error())
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
		serv.Errorf("CreateSlackIntegration error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		serv.Errorf("CreateSlackIntegration error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
}

func InitializeIntegrations(c *mongo.Client) error {
	notifications.IntegrationsCollection = db.GetCollection("integrations", c)
	notifications.NotificationIntegrationsMap = make(map[string]interface{})
	err := notifications.InitializeSlackConnection(c)
	if err != nil {
		return err
	}
	return nil
}
