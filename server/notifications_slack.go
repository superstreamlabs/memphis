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
	"errors"
	"memphis/db"
	"memphis/models"
	"strings"

	"github.com/slack-go/slack"
)

func IsSlackEnabled() (bool, error) {
	exist, _, err := db.GetIntegration("slack")
	if !exist {
		return false, nil
	}
	if err == nil {
		return true, nil
	}

	return false, err
}

func sendMessageToSlackChannel(integration models.SlackIntegration, title string, message string) error {
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

func cacheDetailsSlack(keys map[string]string, properties map[string]bool) {
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
}

func (it IntegrationsHandler) getSlackIntegrationDetails(integrationType string, body models.CreateIntegrationSchema) (map[string]string, map[string]bool, int, error) {
	var authToken, channelID, uiUrl string
	var pmAlert, svfAlert, disconnectAlert bool
	authToken, ok := body.Keys["auth_token"]
	if !ok {
		return map[string]string{}, map[string]bool{}, configuration.SHOWABLE_ERROR_STATUS_CODE, errors.New("Must provide auth token for slack integration")
	}
	channelID, ok = body.Keys["channel_id"]
	if !ok {
		if !ok {
			return map[string]string{}, map[string]bool{}, configuration.SHOWABLE_ERROR_STATUS_CODE, errors.New("Must provide channel ID for slack integration")
		}
	}
	uiUrl = body.UIUrl
	if uiUrl == "" {
		return map[string]string{}, map[string]bool{}, 500, errors.New("Must provide UI url for slack integration")
	}

	pmAlert, ok = body.Properties[PoisonMAlert]
	if !ok {
		pmAlert = false
	}
	svfAlert, ok = body.Properties[SchemaVAlert]
	if !ok {
		svfAlert = false
	}
	disconnectAlert, ok = body.Properties[DisconEAlert]
	if !ok {
		disconnectAlert = false
	}

	keys, properties := createIntegrationsKeysAndProperties(integrationType, authToken, channelID, pmAlert, svfAlert, disconnectAlert, "", "", "", "")
	return keys, properties, 0, nil
}

func (it IntegrationsHandler) handleCreateSlackIntegration(integrationType string, body models.CreateIntegrationSchema) (map[string]string, map[string]bool, models.IntegrationV1, int, error) {
	keys, properties, errorCode, err := it.getSlackIntegrationDetails(integrationType, body)
	if err != nil {
		return keys, properties, models.IntegrationV1{}, errorCode, err
	}
	if UI_HOST == "" {
		UI_HOST = strings.ToLower(body.UIUrl)
	}
	slackIntegration, err := createSlackIntegration(keys, properties, UI_HOST)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid auth token") || strings.Contains(err.Error(), "Invalid channel ID") || strings.Contains(err.Error(), "already exists") {
			return map[string]string{}, map[string]bool{}, models.IntegrationV1{}, configuration.SHOWABLE_ERROR_STATUS_CODE, err
		} else {
			return map[string]string{}, map[string]bool{}, models.IntegrationV1{}, 500, err
		}
	}
	return keys, properties, slackIntegration, 0, nil
}

func (it IntegrationsHandler) handleUpdateSlackIntegration(integrationType string, body models.CreateIntegrationSchema) (models.Integration, int, error) {
	keys, properties, errorCode, err := it.getSlackIntegrationDetails("slack", body)
	if err != nil {
		return models.Integration{}, errorCode, err
	}
	slackIntegration, err := updateSlackIntegration(keys["auth_token"], keys["channel_id"], properties[PoisonMAlert], properties[SchemaVAlert], properties[DisconEAlert], body.UIUrl)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid auth token") || strings.Contains(err.Error(), "Invalid channel ID") {
			return models.Integration{}, configuration.SHOWABLE_ERROR_STATUS_CODE, err
		} else {
			return models.Integration{}, 500, err
		}
	}
	return slackIntegration, 0, nil
}
func createSlackIntegration(keys map[string]string, properties map[string]bool, uiUrl string) (models.IntegrationV1, error) {
	var slackIntegration models.IntegrationV1
	// exist, slackIntegration, err := db.GetIntegration("slack")
	exist := false
	var err error
	if !exist {
		err := testSlackIntegration(keys["auth_token"], keys["channel_id"], "Slack integration with Memphis was added successfully")
		if err != nil {
			return slackIntegration, err
		}
		integrationRes, insertErr := db.InsertNewIntegrationPg("slack", keys, properties)
		if insertErr != nil {
			return slackIntegration, insertErr
		}
		slackIntegration = integrationRes
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
		update := models.SdkClientsUpdates{
			Type:   sendNotificationType,
			Update: properties[SchemaVAlert],
		}
		serv.SendUpdateToClients(update)
		slackIntegration.Keys["auth_token"] = hideSlackAuthToken(keys["auth_token"])
		return slackIntegration, nil
	} else if err != nil {
		return slackIntegration, err
	}
	return slackIntegration, errors.New("Slack integration already exists")
}

func updateSlackIntegration(authToken string, channelID string, pmAlert bool, svfAlert bool, disconnectAlert bool, uiUrl string) (models.Integration, error) {
	var slackIntegration models.Integration
	if authToken == "" {
		exist, integrationFromDb, err := db.GetIntegration("slack")
		if err != nil {
			return models.Integration{}, err
		}
		if !exist {
			return models.Integration{}, errors.New("No auth token was provided")
		}
		authToken = integrationFromDb.Keys["auth_token"]
	}
	err := testSlackIntegration(authToken, channelID, "Slack integration with Memphis was updated successfully")
	if err != nil {
		return slackIntegration, err
	}
	keys, properties := createIntegrationsKeysAndProperties("slack", authToken, channelID, pmAlert, svfAlert, disconnectAlert, "", "", "", "")
	slackIntegration, err = db.UpdateIntegration("slack", keys, properties)
	if err != nil {
		return models.Integration{}, err
	}
	integrationToUpdate := models.CreateIntegrationSchema{
		Name:       "slack",
		Keys:       keys,
		Properties: properties,
		UIUrl:      uiUrl,
	}
	msg, err := json.Marshal(integrationToUpdate)
	if err != nil {
		return models.Integration{}, err
	}
	err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		return models.Integration{}, err
	}
	update := models.SdkClientsUpdates{
		Type:   sendNotificationType,
		Update: svfAlert,
	}
	serv.SendUpdateToClients(update)
	keys["auth_token"] = hideSlackAuthToken(keys["auth_token"])
	slackIntegration.Keys = keys
	slackIntegration.Properties = properties
	return models.Integration{}, nil
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

func hideSlackAuthToken(authToken string) string {
	if authToken != "" {
		authToken = "xoxb-****"
		return authToken
	}
	return authToken
}
