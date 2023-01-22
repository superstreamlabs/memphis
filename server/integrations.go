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
	"memphis-broker/db"

	"memphis-broker/models"

	"github.com/slack-go/slack"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var IntegrationsCache map[string]interface{}
var NotificationFunctionsMap map[string]interface{}
var IntegrationsCollection *mongo.Collection

const PoisonMAlert = "poison_message_alert"
const SchemaVAlert = "schema_validation_fail_alert"
const DisconEAlert = "disconnection_events_alert"

func InitializeIntegrations(c *mongo.Client) error {
	IntegrationsCollection = db.GetCollection("integrations", c)
	IntegrationsCache = make(map[string]interface{})
	NotificationFunctionsMap = make(map[string]interface{})
	NotificationFunctionsMap["slack"] = SendMessageToSlackChannel

	err := InitializeConnection(c, "slack")
	if err != nil {
		return err
	}
	err = InitializeConnection(c, "s3")
	if err != nil {
		return err
	}

	keys, properties := CreateIntegrationsKeysAndProperties("slack", configuration.SANDBOX_SLACK_BOT_TOKEN, configuration.SANDBOX_SLACK_CHANNEL_ID, true, true, true, "", "", "", "")

	if configuration.SANDBOX_ENV == "true" {
		CreateSlackIntegration(keys, properties, configuration.SANDBOX_UI_URL)
	}
	return nil
}

func InitializeConnection(c *mongo.Client, integrationsType string) error {
	filter := bson.M{"name": integrationsType}
	var integration models.Integration
	err := IntegrationsCollection.FindOne(context.TODO(),
		filter).Decode(&integration)
	if err == mongo.ErrNoDocuments {
		return nil
	} else if err != nil {
		return err
	}
	CacheDetails(integrationsType, integration.Keys, integration.Properties)
	return nil
}

func clearCache(integrationsType string) {
	delete(IntegrationsCache, integrationsType)
}

func CacheDetails(integrationType string, keys map[string]string, properties map[string]bool) {
	switch integrationType {
	case "slack":
		var authToken, channelID string
		var poisonMessageAlert, schemaValidationFailAlert, disconnectionEventsAlert bool
		var slackIntegration models.SlackIntegration

		slackIntegration, ok := IntegrationsCache["slack"].(models.SlackIntegration)
		if !ok {
			slackIntegration = models.SlackIntegration{}
			slackIntegration.Keys = make(map[string]string)
			slackIntegration.Properties = make(map[string]bool)
		}
		if keys == nil {
			clearCache("slack")
			return
		}
		if properties == nil {
			poisonMessageAlert = false
			schemaValidationFailAlert = false
			disconnectionEventsAlert = false
		}
		authToken, ok = keys["auth_token"]
		if !ok {
			clearCache("slack")
			return
		}
		channelID, ok = keys["channel_id"]
		if !ok {
			clearCache("slack")
			return
		}
		poisonMessageAlert, ok = properties[PoisonMAlert]
		if !ok {
			poisonMessageAlert = false
		}
		schemaValidationFailAlert, ok = properties[SchemaVAlert]
		if !ok {
			schemaValidationFailAlert = false
		}
		disconnectionEventsAlert, ok = properties[DisconEAlert]
		if !ok {
			disconnectionEventsAlert = false
		}
		if slackIntegration.Keys["auth_token"] != authToken {
			slackIntegration.Keys["auth_token"] = authToken
			if authToken != "" {
				slackIntegration.Client = slack.New(authToken)
			}
		}

		slackIntegration.Keys["channel_id"] = channelID
		slackIntegration.Properties[PoisonMAlert] = poisonMessageAlert
		slackIntegration.Properties[SchemaVAlert] = schemaValidationFailAlert
		slackIntegration.Properties[DisconEAlert] = disconnectionEventsAlert
		slackIntegration.Name = "slack"
		IntegrationsCache["slack"] = slackIntegration
	case "s3":
		s3Integration, ok := IntegrationsCache["s3"].(models.S3Integration)
		if !ok {
			s3Integration = models.S3Integration{}
			s3Integration.Keys = make(map[string]string)
			s3Integration.Properties = make(map[string]bool)
		}
		if keys == nil {
			clearCache("s3")
			return
		}

		s3Integration.Keys["access_key"] = keys["access_key"]
		s3Integration.Keys["secret_key"] = keys["secret_key"]
		s3Integration.Keys["bucket_name"] = keys["bucket_name"]
		s3Integration.Keys["region"] = keys["region"]
		s3Integration.Name = "s3"
		IntegrationsCache["s3"] = s3Integration

	}

}
