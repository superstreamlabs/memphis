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
	"fmt"
	"memphis/analytics"
	"memphis/db"
	"memphis/models"

	"errors"
	"strings"
	"time"
)

type ConnectionsHandler struct{}

var connectionsHandler ConnectionsHandler
var producersHandler ProducersHandler
var consumersHandler ConsumersHandler

const (
	connectItemSep                      = "::"
	userNameItemSep                     = "$"
	connectConfigUpdatesSubjectTemplate = "$memphis_configurations_updates.init.%s"
)

func updateNewClientWithConfig(c *client, connId string) {
	// TODO more configurations logic here

	slackEnabled, err := IsSlackEnabled(c.acc.GetName())
	if err != nil {
		c.Errorf("updateNewClientWithConfig: %v", err.Error())
	}

	config := models.GlobalConfigurationsUpdate{
		Notifications: slackEnabled,
	}

	sendConnectUpdate(c, config, connId)
}

func sendConnectUpdate(c *client, ccu models.GlobalConfigurationsUpdate, connId string) {
	s := c.srv
	rawMsg, err := json.Marshal(ccu)
	if err != nil {
		s.Errorf("sendConnectUpdate: %v", err.Error())
		return
	}
	subject := fmt.Sprintf(connectConfigUpdatesSubjectTemplate, connId)
	s.sendInternalAccountMsg(c.acc, subject, rawMsg)
}

func handleConnectMessage(client *client) error {
	splittedMemphisInfo := strings.Split(client.opts.Name, connectItemSep)

	var (
		isNativeMemphisClient bool
		username              string
		connectionId          string
		err                   error
	)
	switch len(splittedMemphisInfo) {
	case 2:
		// normal Memphis SDK carry on to the rest of the function
		isNativeMemphisClient = true
		username = strings.ToLower(splittedMemphisInfo[1])
	case 1:
		// NATS SDK, means we extract username from the token field
		isNativeMemphisClient = false
		var splittedToken []string
		if configuration.USER_PASS_BASED_AUTH {
			splittedToken = strings.Split(client.opts.Username, userNameItemSep)
			username = strings.ToLower(splittedToken[0])
		} else {
			splittedToken := strings.Split(client.opts.Token, connectItemSep)
			if len(splittedToken) != 2 {
				client.Warnf("handleConnectMessage: missing username or token")
				return errors.New("missing username or token")
			}
			username, tenantId, err := getUserAndTenantIdFromString(strings.ToLower(splittedToken[0]))
			if err != nil {
				client.Errorf("[tenant Id: %v]handleConnectMessage: User %v : %v", tenantId, username, err.Error())
				return err
			}
		}
	default:
		client.Warnf("handleConnectMessage: missing username or connectionId")
		return errors.New("missing username or connectionId")
	}
	exist, user, err := db.GetUserByUsername(username, client.acc.GetName())
	if err != nil {
		client.Errorf("handleConnectMessage: User %v : %v", username, err.Error())
		return err
	}
	if !exist {
		errMsg := fmt.Sprintf("handleConnectMessage:  User %v does not exist", username)
		client.Warnf(errMsg)
		return errors.New(errMsg)
	}
	if user.UserType != "root" && user.UserType != "application" {
		client.Warnf("[tenant: %v][user: %v] handleConnectMessage: Please use a user of type Root/Application and not Management", user.TenantName, user.Username)
		return errors.New("please use a user of type Root/Application and not Management")
	}

	if isNativeMemphisClient {
		connectionId = splittedMemphisInfo[0]
		exist, err := connectionsHandler.CreateConnection(user.ID, client.RemoteAddress().String(), connectionId, user.Username, client.Account().GetName())
		if err != nil {
			client.Errorf("[tenant: %v][user: %v]handleConnectMessage at CreateConnection: %v", user.TenantName, username, err.Error())
			return err
		}
		if exist {
			err = connectionsHandler.ReliveConectionResources(connectionId)
			if err != nil {
				client.Errorf("[tenant: %v][user: %v]handleConnectMessage at ReliveConectionResources: %v", user.TenantName, username, err.Error())
				return err
			}

		} else {
			go func() {
				shouldSendAnalytics, _ := shouldSendAnalytics()
				if shouldSendAnalytics { // exist indicates it is a reconnect
					splitted := strings.Split(client.opts.Lang, ".")
					sdkName := splitted[len(splitted)-1]
					param := analytics.EventParam{
						Name:  "sdk",
						Value: sdkName,
					}
					analyticsParams := []analytics.EventParam{param}
					event := "user-connect-sdk"
					if !isNativeMemphisClient {
						event = "user-connect-nats-sdk"
					}
					analytics.SendEventWithParams(user.TenantName, username, analyticsParams, event)
				}
			}()
		}
		updateNewClientWithConfig(client, connectionId)
	}

	client.memphisInfo = memphisClientInfo{username: username, connectionId: connectionId, isNative: isNativeMemphisClient}
	return nil
}

func (ch ConnectionsHandler) CreateConnection(userId int, clientAddress string, connectionId string, createdByUsername string, tenantName string) (bool, error) {
	createdByUsername = strings.ToLower(createdByUsername)
	newConnection := models.Connection{
		ID:                connectionId,
		CreatedBy:         userId,
		CreatedByUsername: createdByUsername,
		IsActive:          true,
		CreatedAt:         time.Now(),
		ClientAddress:     clientAddress,
		TenantName:        strings.ToLower(tenantName),
	}

	err := db.InsertConnection(newConnection, tenantName)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (ch ConnectionsHandler) ReliveConectionResources(connectionId string) error {
	err := db.ReliveConectionResources(connectionId, true)
	if err != nil {
		serv.Errorf("ReliveConectionResources error: %v", err.Error())
		return err
	}
	return nil
}

func (mci *memphisClientInfo) updateDisconnection(tenantName string) error {
	if mci.connectionId == "" {
		return nil
	}

	err := db.UpdateConnection(mci.connectionId, false)
	if err != nil {
		return err
	}
	producers, err := db.GetProducersByConnectionIDWithStationDetails(mci.connectionId)
	if err != nil {
		return err
	}
	var producerNames, consumerNames string
	if len(producers) > 0 {
		err = db.UpdateProducersConnection(mci.connectionId, false)
		if err != nil {
			return err
		}

		for i := 0; i < len(producers); i++ {
			producerNames = producerNames + "Producer: " + producers[i].Name + " Station: " + producers[i].StationName + "\n"
		}
	}

	consumers, err := db.GetConsumersByConnectionIDWithStationDetails(mci.connectionId)
	if err != nil {
		return err
	}
	if len(consumers) > 0 {
		err = db.UpdateConsumersConnection(mci.connectionId, false)
		if err != nil {
			return err
		}

		for i := 0; i < len(consumers); i++ {
			consumerNames = consumerNames + "Consumer: " + consumers[i].Name + " | Station: " + consumers[i].StationName + "\n"
		}
	}
	msg := ""
	if len(producerNames) > 0 {
		msg = producerNames
	}
	if len(consumerNames) > 0 {
		msg = msg + consumerNames
	}

	if len(consumerNames) > 0 || len(producerNames) > 0 {
		err = SendNotification(tenantName, "Disconnection events", msg, DisconEAlert)
		if err != nil {
			return err
		}
	}

	return nil
}
