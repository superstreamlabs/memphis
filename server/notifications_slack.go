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
	"strings"

	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"

	"github.com/slack-go/slack"
)

func IsSlackEnabled(tenantName string) (bool, error) {
	exist, _, err := db.GetIntegration("slack", tenantName)
	if err != nil {
		return false, err
	}

	if !exist {
		return false, nil
	}

	return true, nil
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

func cacheDetailsSlack(keys map[string]interface{}, properties map[string]bool, tenantName string) {
	var authToken, channelID string
	var poisonMessageAlert, schemaValidationFailAlert, disconnectionEventsAlert bool
	slackIntegration := models.SlackIntegration{}
	slackIntegration.Keys = make(map[string]string)
	slackIntegration.Properties = make(map[string]bool)
	if keys == nil {
		deleteIntegrationFromTenant(tenantName, "slack", IntegrationsConcurrentCache)
		return
	}
	if properties == nil {
		poisonMessageAlert = false
		schemaValidationFailAlert = false
		disconnectionEventsAlert = false
	}
	authToken, ok := keys["auth_token"].(string)
	if !ok {
		deleteIntegrationFromTenant(tenantName, "slack", IntegrationsConcurrentCache)
		return
	}
	channelID, ok = keys["channel_id"].(string)
	if !ok {
		deleteIntegrationFromTenant(tenantName, "slack", IntegrationsConcurrentCache)
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
	if _, ok := IntegrationsConcurrentCache.Load(tenantName); !ok {
		IntegrationsConcurrentCache.Add(tenantName, map[string]interface{}{"slack": slackIntegration})
	} else {
		err := addIntegrationToTenant(tenantName, "slack", IntegrationsConcurrentCache, slackIntegration)
		if err != nil {
			serv.Errorf("cacheDetailsSlack: " + err.Error())
			return
		}
	}
}

func (it IntegrationsHandler) getSlackIntegrationDetails(body models.CreateIntegrationSchema) (map[string]interface{}, map[string]bool, int, error) {
	var authToken, channelID, uiUrl string
	var pmAlert, svfAlert, disconnectAlert bool
	authToken, ok := body.Keys["auth_token"].(string)
	if !ok {
		return map[string]interface{}{}, map[string]bool{}, SHOWABLE_ERROR_STATUS_CODE, errors.New("must provide auth token for slack integration")
	}
	channelID, ok = body.Keys["channel_id"].(string)
	if !ok {
		if !ok {
			return map[string]interface{}{}, map[string]bool{}, SHOWABLE_ERROR_STATUS_CODE, errors.New("must provide channel ID for slack integration")
		}
	}
	uiUrl = body.UIUrl
	if uiUrl == "" {
		return map[string]interface{}{}, map[string]bool{}, 500, errors.New("must provide UI url for slack integration")
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

	keys, properties := createIntegrationsKeysAndProperties("slack", authToken, channelID, pmAlert, svfAlert, disconnectAlert, "", "", "", "", "", "", map[string]interface{}{}, "", "", "", "")
	return keys, properties, 0, nil
}

func (it IntegrationsHandler) handleCreateSlackIntegration(tenantName string, body models.CreateIntegrationSchema) (map[string]interface{}, map[string]bool, models.Integration, int, error) {
	keys, properties, errorCode, err := it.getSlackIntegrationDetails(body)
	if err != nil {
		return keys, properties, models.Integration{}, errorCode, err
	}
	if it.S.opts.UiHost == "" {
		EditClusterCompHost("ui_host", body.UIUrl)
	}
	slackIntegration, err := createSlackIntegration(tenantName, keys, properties, body.UIUrl)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "invalid auth token") || strings.Contains(errMsg, "invalid channel") || strings.Contains(errMsg, "already exists") {
			return map[string]interface{}{}, map[string]bool{}, models.Integration{}, SHOWABLE_ERROR_STATUS_CODE, err
		} else {
			return map[string]interface{}{}, map[string]bool{}, models.Integration{}, 500, err
		}
	}
	return keys, properties, slackIntegration, 0, nil
}

func (it IntegrationsHandler) handleUpdateSlackIntegration(tenantName, integrationType string, body models.CreateIntegrationSchema) (models.Integration, int, error) {
	keys, properties, errorCode, err := it.getSlackIntegrationDetails(body)
	if err != nil {
		return models.Integration{}, errorCode, err
	}
	slackIntegration, err := updateSlackIntegration(tenantName, keys["auth_token"].(string), keys["channel_id"].(string), properties[PoisonMAlert], properties[SchemaVAlert], properties[DisconEAlert], body.UIUrl)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "invalid auth token") || strings.Contains(errMsg, "invalid channel") {
			return models.Integration{}, SHOWABLE_ERROR_STATUS_CODE, err
		} else {
			return models.Integration{}, 500, err
		}
	}
	return slackIntegration, 0, nil
}
func createSlackIntegration(tenantName string, keys map[string]interface{}, properties map[string]bool, uiUrl string) (models.Integration, error) {
	var slackIntegration models.Integration
	exist, slackIntegration, err := db.GetIntegration("slack", tenantName)
	if err != nil {
		return slackIntegration, err
	} else if !exist {
		err := testSlackIntegration(keys["auth_token"].(string), keys["channel_id"].(string), "Slack integration with Memphis was added successfully")
		if err != nil {
			return slackIntegration, err
		}
		stringMapKeys := GetKeysAsStringMap(keys)
		cloneKeys := copyMaps(stringMapKeys)
		encryptedValue, err := EncryptAES([]byte(keys["auth_token"].(string)))
		if err != nil {
			return models.Integration{}, err
		}
		cloneKeys["auth_token"] = encryptedValue
		interfaceMapKeys := copyStringMapToInterfaceMap(cloneKeys)
		integrationRes, insertErr := db.InsertNewIntegration(tenantName, "slack", interfaceMapKeys, properties)
		if insertErr != nil {
			return slackIntegration, insertErr
		}
		slackIntegration = integrationRes
		integrationToUpdate := models.CreateIntegration{
			Name:       "slack",
			Keys:       keys,
			Properties: properties,
			UIUrl:      uiUrl,
			TenantName: tenantName,
		}
		msg, err := json.Marshal(integrationToUpdate)
		if err != nil {
			return slackIntegration, err
		}
		err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
		if err != nil {
			return slackIntegration, err
		}
		update := models.SdkClientsUpdates{
			Type:   sendNotificationType,
			Update: properties[SchemaVAlert],
		}
		serv.SendUpdateToClients(update)
		slackIntegration.Keys["auth_token"] = hideSlackAuthToken(keys["auth_token"].(string))
		return slackIntegration, nil
	}
	return slackIntegration, errors.New("slack integration already exists")
}

func updateSlackIntegration(tenantName string, authToken string, channelID string, pmAlert bool, svfAlert bool, disconnectAlert bool, uiUrl string) (models.Integration, error) {
	var slackIntegration models.Integration
	if authToken == "" {
		exist, integrationFromDb, err := db.GetIntegration("slack", tenantName)
		if err != nil {
			return models.Integration{}, err
		}
		if !exist {
			return models.Integration{}, errors.New("no auth token was provided")
		}
		key := getAESKey()
		token, err := DecryptAES(key, integrationFromDb.Keys["auth_token"].(string))
		if err != nil {
			return models.Integration{}, err
		}
		authToken = token
	}
	err := testSlackIntegration(authToken, channelID, "Slack integration with Memphis was updated successfully")
	if err != nil {
		return slackIntegration, err
	}
	keys, properties := createIntegrationsKeysAndProperties("slack", authToken, channelID, pmAlert, svfAlert, disconnectAlert, "", "", "", "", "", "", map[string]interface{}{}, "", "", "", "")
	stringMapKeys := GetKeysAsStringMap(keys)
	cloneKeys := copyMaps(stringMapKeys)
	encryptedValue, err := EncryptAES([]byte(authToken))
	if err != nil {
		return models.Integration{}, err
	}
	cloneKeys["auth_token"] = encryptedValue
	interfaceMapKeys := copyStringMapToInterfaceMap(cloneKeys)
	slackIntegration, err = db.UpdateIntegration(tenantName, "slack", interfaceMapKeys, properties)
	if err != nil {
		return models.Integration{}, err
	}

	integrationToUpdate := models.CreateIntegration{
		Name:       "slack",
		Keys:       keys,
		Properties: properties,
		UIUrl:      uiUrl,
		TenantName: tenantName,
	}
	msg, err := json.Marshal(integrationToUpdate)
	if err != nil {
		return models.Integration{}, err
	}
	err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		return models.Integration{}, err
	}
	update := models.SdkClientsUpdates{
		Type:   sendNotificationType,
		Update: svfAlert,
	}
	serv.SendUpdateToClients(update)
	keys["auth_token"] = hideSlackAuthToken(cloneKeys["auth_token"])
	slackIntegration.Keys = keys
	slackIntegration.Properties = properties
	return slackIntegration, nil
}

func testSlackIntegration(authToken string, channelID string, message string) error {
	slackClientTemp := slack.New(authToken)
	_, err := slackClientTemp.AuthTest()
	if err != nil {
		return errors.New("invalid auth token")
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
		return errors.New("invalid channel ID")
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
