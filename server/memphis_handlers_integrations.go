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

	"memphis-broker/models"
	"memphis-broker/slack"
	"memphis-broker/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IntegrationsHandler struct{ S *Server }

func (it IntegrationsHandler) CreateSlackIntegration(c *gin.Context) {
	var body models.SlackIntegrationRequest
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	keys := make(map[string]string)
	keys["auth_token"] = body.AuthToken
	keys["channel_id"] = body.ChannelID
	properties := make(map[string]bool)
	properties["poison_message_alert"] = body.PoisonMessageAlert
	properties["schema_validation_fail_alert"] = body.SchemaValidationFailAlert
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
		}
		integrationsCollection.InsertOne(context.TODO(), slackIntegration)
	} else if err != nil {
		serv.Errorf("CreateSlackIntegration error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	slack.UpdateSlackDetails(body.AuthToken, body.ChannelID, body.PoisonMessageAlert, body.SchemaValidationFailAlert)

	c.IndentedJSON(200, slackIntegration)
}

func (it IntegrationsHandler) UpdateSlackIntegration(c *gin.Context) {
	var body models.SlackIntegrationRequest
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	keys := make(map[string]string)
	keys["auth_token"] = body.AuthToken
	keys["channel_id"] = body.ChannelID
	properties := make(map[string]bool)
	properties["poison_message_alert"] = body.PoisonMessageAlert
	properties["schema_validation_fail_alert"] = body.SchemaValidationFailAlert
	filter := bson.M{"name": "slack"}
	var slackIntegration models.Integration
	err := integrationsCollection.FindOneAndUpdate(context.TODO(),
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
		serv.Errorf("CreateSlackIntegration error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	slack.UpdateSlackDetails(body.AuthToken, body.ChannelID, body.PoisonMessageAlert, body.SchemaValidationFailAlert)

	c.IndentedJSON(200, slackIntegration)
}

func (it IntegrationsHandler) GetIntegrationDetails(c *gin.Context) {
	var body models.GetIntegrationDetailsRequest
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	filter := bson.M{"name": body.Name}
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
