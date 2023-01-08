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
package notifications

import (
	"context"

	"memphis-broker/models"

	"github.com/slack-go/slack"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeSlackConnection(c *mongo.Client) error {
	filter := bson.M{"name": "slack"}
	var slackIntegration models.Integration
	err := IntegrationsCollection.FindOne(context.TODO(),
		filter).Decode(&slackIntegration)
	if err == mongo.ErrNoDocuments {
		return nil
	} else if err != nil {
		return err
	}
	CacheSlackDetails(slackIntegration.Keys, slackIntegration.Properties)
	return nil
}

func IsSlackEnabled() (bool, error) {
	filter := bson.M{"name": "slack"}
	var slackIntegration models.Integration
	err := IntegrationsCollection.FindOne(context.TODO(),
		filter).Decode(&slackIntegration)
	if err == nil {
		return true, nil
	}

	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return false, err
}

func clearSlackCache() {
	delete(NotificationIntegrationsMap, "slack")
}

func CacheSlackDetails(keys map[string]string, properties map[string]bool) {
	var authToken, channelID string
	var poisonMessageAlert, schemaValidationFailAlert, disconnectionEventsAlert bool
	var slackIntegration models.SlackIntegration

	slackIntegration, ok := NotificationIntegrationsMap["slack"].(models.SlackIntegration)
	if !ok {
		slackIntegration = models.SlackIntegration{}
		slackIntegration.Keys = make(map[string]string)
		slackIntegration.Properties = make(map[string]bool)
	}
	if keys == nil {
		clearSlackCache()
		return
	}
	if properties == nil {
		poisonMessageAlert = false
		schemaValidationFailAlert = false
		disconnectionEventsAlert = false
	}
	authToken, ok = keys["auth_token"]
	if !ok {
		clearSlackCache()
		return
	}
	channelID, ok = keys["channel_id"]
	if !ok {
		clearSlackCache()
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
	NotificationIntegrationsMap["slack"] = slackIntegration
}

func SendMessageToSlackChannel(integration models.SlackIntegration, title string, message string) error {
	attachment := slack.Attachment{
		AuthorName: "Memphis",
		Title:      title,
		Text:       message,
		Color:      "#6557FF",
	}

	_, _, err := integration.Client.PostMessage(
		integration.Keys["channel_id"],
		slack.MsgOptionAttachments(attachment),
	)
	if err != nil {
		return err
	}
	return nil
}
