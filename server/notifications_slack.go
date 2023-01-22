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
	"github.com/slack-go/slack"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"memphis-broker/models"
)

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
