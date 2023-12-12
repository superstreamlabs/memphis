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
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"
)

type Author struct {
	Name string `json:"name"`
}

type Embed struct {
	Color       string `json:"color"`
	Author      Author `json:"author"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type DiscordMessage struct {
	Embeds []Embed `json:"embeds"`
}

func IsDiscordEnabled(tenantName string) (bool, error) {
	exist, _, err := db.GetIntegration("discord", tenantName)
	if err != nil {
		return false, err
	}

	if !exist {
		return false, nil
	}

	return true, nil
}

func sendMessageToDiscordChannel(integration models.DiscordIntegration, title string, message string) error {
	payload := DiscordMessage{
		Embeds: []Embed{
			{
				Color: "6641663",
				Author: Author{
					Name: "Memphis",
				},
				Title:       title,
				Description: message,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		integration.Keys["webhook_url"],
		"application/json",
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func cacheDetailsDiscord(keys map[string]interface{}, properties map[string]bool, tenantName string) {
	var webhookUrl string
	var poisonMessageAlert, schemaValidationFailAlert, disconnectionEventsAlert bool
	discordIntegration := models.DiscordIntegration{}
	discordIntegration.Keys = make(map[string]string)
	discordIntegration.Properties = make(map[string]bool)
	if keys == nil {
		deleteIntegrationFromTenant(tenantName, "discord", IntegrationsConcurrentCache)
		return
	}
	if properties == nil {
		poisonMessageAlert = false
		schemaValidationFailAlert = false
		disconnectionEventsAlert = false
	}
	webhookUrl, ok := keys["webhook_url"].(string)
	if !ok {
		deleteIntegrationFromTenant(tenantName, "discord", IntegrationsConcurrentCache)
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
	if discordIntegration.Keys["webhook_url"] != webhookUrl {
		discordIntegration.Keys["webhook_url"] = webhookUrl
	}

	discordIntegration.Properties[PoisonMAlert] = poisonMessageAlert
	discordIntegration.Properties[SchemaVAlert] = schemaValidationFailAlert
	discordIntegration.Properties[DisconEAlert] = disconnectionEventsAlert
	discordIntegration.Name = "discord"
	if _, ok := IntegrationsConcurrentCache.Load(tenantName); !ok {
		IntegrationsConcurrentCache.Add(tenantName, map[string]interface{}{"discord": discordIntegration})
	} else {
		err := addIntegrationToTenant(tenantName, "discord", IntegrationsConcurrentCache, discordIntegration)
		if err != nil {
			serv.Errorf("cacheDetailsDiscord: " + err.Error())
			return
		}
	}
}

func (it IntegrationsHandler) getDiscordIntegrationDetails(body models.CreateIntegrationSchema) (map[string]interface{}, map[string]bool, int, error) {
	var webhookUrl, uiUrl string
	var pmAlert, svfAlert, disconnectAlert bool
	webhookUrl, ok := body.Keys["webhook_url"].(string)
	if !ok {
		return map[string]interface{}{}, map[string]bool{}, SHOWABLE_ERROR_STATUS_CODE, errors.New("must provide webhook url for discord integration")
	}
	uiUrl = body.UIUrl
	if uiUrl == "" {
		return map[string]interface{}{}, map[string]bool{}, 500, errors.New("must provide UI url for discord integration")
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

	keys, properties := createIntegrationsKeysAndProperties("discord", "", "", webhookUrl, pmAlert, svfAlert, disconnectAlert, "", "", "", "", "", "", map[string]interface{}{}, "", "", "", "")
	return keys, properties, 0, nil
}

func (it IntegrationsHandler) handleCreateDiscordIntegration(tenantName string, body models.CreateIntegrationSchema) (map[string]interface{}, map[string]bool, models.Integration, int, error) {
	keys, properties, errorCode, err := it.getDiscordIntegrationDetails(body)
	if err != nil {
		return keys, properties, models.Integration{}, errorCode, err
	}
	if it.S.opts.UiHost == "" {
		EditClusterCompHost("ui_host", body.UIUrl)
	}
	discordIntegration, err := createDiscordIntegration(tenantName, keys, properties, body.UIUrl)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "invalid webhook url") {
			return map[string]interface{}{}, map[string]bool{}, models.Integration{}, SHOWABLE_ERROR_STATUS_CODE, err
		} else {
			return map[string]interface{}{}, map[string]bool{}, models.Integration{}, 500, err
		}
	}
	return keys, properties, discordIntegration, 0, nil
}

func (it IntegrationsHandler) handleUpdateDiscordIntegration(tenantName, integrationType string, body models.CreateIntegrationSchema) (models.Integration, int, error) {
	keys, properties, errorCode, err := it.getDiscordIntegrationDetails(body)
	if err != nil {
		return models.Integration{}, errorCode, err
	}
	discordIntegration, err := updateDiscordIntegration(tenantName, keys["webhook_url"].(string), properties[PoisonMAlert], properties[SchemaVAlert], properties[DisconEAlert], body.UIUrl)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "invalid webhook url") {
			return models.Integration{}, SHOWABLE_ERROR_STATUS_CODE, err
		} else {
			return models.Integration{}, 500, err
		}
	}
	return discordIntegration, 0, nil
}

func createDiscordIntegration(tenantName string, keys map[string]interface{}, properties map[string]bool, uiUrl string) (models.Integration, error) {
	var discordIntegration models.Integration
	exist, discordIntegration, err := db.GetIntegration("discord", tenantName)
	if err != nil {
		return discordIntegration, err
	} else if !exist {
		err := testDiscordIntegration(keys["webhook_url"].(string))
		if err != nil {
			return discordIntegration, err
		}
		stringMapKeys := GetKeysAsStringMap(keys)
		cloneKeys := copyMaps(stringMapKeys)
		encryptedValue, err := EncryptAES([]byte(keys["webhook_url"].(string)))
		if err != nil {
			return models.Integration{}, err
		}
		cloneKeys["webhook_url"] = encryptedValue
		interfaceMapKeys := copyStringMapToInterfaceMap(cloneKeys)
		integrationRes, insertErr := db.InsertNewIntegration(tenantName, "discord", interfaceMapKeys, properties)
		if insertErr != nil {
			return discordIntegration, insertErr
		}
		discordIntegration = integrationRes
		integrationToUpdate := models.CreateIntegration{
			Name:       "discord",
			Keys:       keys,
			Properties: properties,
			UIUrl:      uiUrl,
			TenantName: tenantName,
			IsValid:    integrationRes.IsValid,
		}
		msg, err := json.Marshal(integrationToUpdate)
		if err != nil {
			return discordIntegration, err
		}
		err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
		if err != nil {
			return discordIntegration, err
		}
		update := models.SdkClientsUpdates{
			Type:   sendNotificationType,
			Update: properties[SchemaVAlert],
		}
		serv.SendUpdateToClients(update)
		discordIntegration.Keys["webhook_url"] = hideDiscordWebhookUrl(keys["webhook_url"].(string))
		return discordIntegration, nil
	}
	return discordIntegration, errors.New("discord integration already exists")
}

func updateDiscordIntegration(tenantName string, webhookUrl string, pmAlert bool, svfAlert bool, disconnectAlert bool, uiUrl string) (models.Integration, error) {
	var discordIntegration models.Integration
	if webhookUrl == "" {
		exist, integrationFromDb, err := db.GetIntegration("discord", tenantName)
		if err != nil {
			return models.Integration{}, err
		}
		if !exist {
			return models.Integration{}, errors.New("no webhook url was provided")
		}
		key := getAESKey()
		url, err := DecryptAES(key, integrationFromDb.Keys["webhook_url"].(string))
		if err != nil {
			return models.Integration{}, err
		}
		webhookUrl = url
	}
	err := testDiscordIntegration(webhookUrl)
	if err != nil {
		return discordIntegration, err
	}
	keys, properties := createIntegrationsKeysAndProperties("discord", "", "", webhookUrl, pmAlert, svfAlert, disconnectAlert, "", "", "", "", "", "", map[string]interface{}{}, "", "", "", "")
	stringMapKeys := GetKeysAsStringMap(keys)
	cloneKeys := copyMaps(stringMapKeys)
	encryptedValue, err := EncryptAES([]byte(webhookUrl))
	if err != nil {
		return models.Integration{}, err
	}
	cloneKeys["webhook_url"] = encryptedValue
	interfaceMapKeys := copyStringMapToInterfaceMap(cloneKeys)
	discordIntegration, err = db.UpdateIntegration(tenantName, "discord", interfaceMapKeys, properties)
	if err != nil {
		return models.Integration{}, err
	}

	integrationToUpdate := models.CreateIntegration{
		Name:       "discord",
		Keys:       keys,
		Properties: properties,
		UIUrl:      uiUrl,
		TenantName: tenantName,
		IsValid:    discordIntegration.IsValid,
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
	keys["webhook_url"] = hideDiscordWebhookUrl(cloneKeys["webhook_url"])
	discordIntegration.Keys = keys
	discordIntegration.Properties = properties
	return discordIntegration, nil
}

func testDiscordIntegration(webhookUrl string) error {
	resp, err := http.Head(webhookUrl)
	if err != nil {
		return errors.New("invalid webhook url")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("invalid webhook url")
	}
	return nil
}

func hideDiscordWebhookUrl(webhookUrl string) string {
	if webhookUrl != "" {
		webhookUrl = "https://discord.com/api/webhooks/****"
		return webhookUrl
	}
	return webhookUrl
}
