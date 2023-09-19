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
	"strings"

	"github.com/memphisdev/memphis/analytics"
	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"
	"github.com/memphisdev/memphis/utils"

	"github.com/gin-gonic/gin"
)

const sendNotificationType = "send_notification"

type IntegrationsHandler struct{ S *Server }

func (it IntegrationsHandler) DisconnectIntegration(c *gin.Context) {
	var body models.DisconnectIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("DisconnectIntegration at getUserDetailsFromMiddleware: Integration %v: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	exist, _, err := db.GetTenantByName(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant:%v]DisconnectIntegration at GetTenantByName: %v", user.TenantName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("[tenant: %v]DisconnectIntegration : tenant %v does not exist", user.TenantName, user.TenantName)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	integrationType := strings.ToLower(body.Name)
	err = db.DeleteIntegration(integrationType, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v]DisconnectIntegration at db.DeleteIntegration: Integration %v: %v", user.TenantName, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	integrationUpdate := models.Integration{
		Name:       strings.ToLower(body.Name),
		Keys:       nil,
		Properties: nil,
		TenantName: user.TenantName,
	}

	msg, err := json.Marshal(integrationUpdate)
	if err != nil {
		serv.Errorf("[tenant: %v]DisconnectIntegration at json.Marshal: Integration %v: %v", user.TenantName, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		serv.Errorf("[tenant: %v]DisconnectIntegration at sendInternalAccountMsgWithReply: Integration %v: %v", user.TenantName, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	switch body.Name {
	case "slack":
		update := models.SdkClientsUpdates{
			Type:   sendNotificationType,
			Update: false,
		}
		serv.SendUpdateToClients(update)
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-disconnect-integration-"+integrationType)
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
		analyticsParams := map[string]interface{}{"request-content": body.RequestContent}
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-request-integration")
	}

	c.IndentedJSON(200, gin.H{})
}
