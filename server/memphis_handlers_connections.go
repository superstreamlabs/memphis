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
	connectConfigUpdatesSubjectTemplate = CONFIGURATIONS_UPDATES_SUBJ + ".init.%s"
)

func updateNewClientWithConfig(c *client, connId string) {
	// TODO more configurations logic here

	slackEnabled, err := IsSlackEnabled()
	if err != nil {
		c.Errorf("updateNewClientWithConfig: " + err.Error())
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
		s.Errorf("sendConnectUpdate: " + err.Error())
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
		objID                 string
	)
	switch len(splittedMemphisInfo) {
	case 2:
		// normal Memphis SDK carry on to the rest of the function
		isNativeMemphisClient = true
		username = strings.ToLower(splittedMemphisInfo[1])
	case 1:
		// NATS SDK, means we extract username from the token field
		isNativeMemphisClient = false
		splittedToken := strings.Split(client.opts.Token, connectItemSep)
		if len(splittedToken) != 2 {
			client.Warnf("handleConnectMessage: missing username or token")
			return errors.New("missing username or token")
		}
		username = strings.ToLower(splittedToken[0])
	default:
		client.Warnf("handleConnectMessage: missing username or connectionId")
		return errors.New("missing username or connectionId")
	}

	exist, user, err := db.GetUserByUsername(username)
	if err != nil {
		errMsg := "User " + username + ": " + err.Error()
		client.Errorf("handleConnectMessage: " + errMsg)
		return err
	}
	if !exist {
		errMsg := "User " + username + " does not exist"
		client.Warnf("handleConnectMessage: " + errMsg)
		return errors.New(errMsg)
	}
	if user.UserType != "root" && user.UserType != "application" {
		client.Warnf("handleConnectMessage: Please use a user of type Root/Application and not Management")
		return errors.New("Please use a user of type Root/Application and not Management")
	}

	if isNativeMemphisClient {
		objID := splittedMemphisInfo[0]

		exist, _, err = db.GetConnectionByID(objID)
		if err != nil {
			errMsg := "User " + username + ": " + err.Error()
			client.Errorf("handleConnectMessage: " + errMsg)
			return err
		}

		if exist {
			err = connectionsHandler.ReliveConnection(objID)
			if err != nil {
				errMsg := "User " + username + ": " + err.Error()
				client.Errorf("handleConnectMessage: " + errMsg)
				return err
			}
			err = producersHandler.ReliveProducers(objID)
			if err != nil {
				errMsg := "User " + username + ": " + err.Error()
				client.Errorf("handleConnectMessage: " + errMsg)
				return err
			}
			err = consumersHandler.ReliveConsumers(objID)
			if err != nil {

				errMsg := "User " + username + ": " + err.Error()
				client.Errorf("handleConnectMessage: " + errMsg)
				return err
			}
		} else {
			err := connectionsHandler.CreateConnection(username, client.RemoteAddress().String(), objID)
			if err != nil {
				errMsg := "User " + username + ": " + err.Error()
				client.Errorf("handleConnectMessage: " + errMsg)
				return err
			}
		}
		updateNewClientWithConfig(client, objID)
	}

	client.memphisInfo = memphisClientInfo{username: username, connectionId: objID, isNative: isNativeMemphisClient}
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if !exist && shouldSendAnalytics { // exist indicates it is a reconnect
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
		analytics.SendEventWithParams(username, analyticsParams, event)
	}

	return nil
}

func (ch ConnectionsHandler) CreateConnection(username, clientAddress string, connectionId string) error {
	username = strings.ToLower(username)
	exist, user, err := db.GetUserByUsername(username)
	if err != nil {
		errMsg := "User " + username + ": " + err.Error()
		serv.Errorf("CreateConnection error: " + errMsg)
		return err
	}
	if !exist {
		errMsg := "User " + username + " does not exist"
		return errors.New(errMsg)
	}

	newConnection := models.Connection{
		ID:            connectionId,
		CreatedBy:     user.ID,
		IsActive:      true,
		CreationDate:  time.Now(),
		ClientAddress: clientAddress,
	}

	err = db.InsertConnection(newConnection)
	if err != nil {
		errMsg := "User " + username + ": " + err.Error()
		serv.Errorf("CreateConnection error: " + errMsg)
		return err
	}
	return nil
}

func (ch ConnectionsHandler) ReliveConnection(connectionId string) error {
	err := db.UpdateConnection(connectionId, true)
	if err != nil {
		serv.Errorf("ReliveConnection error: " + err.Error())
		return err
	}
	return nil
}

func (mci *memphisClientInfo) updateDisconnection() error {
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
	err = SendNotification("Disconnection events", msg, DisconEAlert)
	if err != nil {
		return err
	}

	serv.Noticef("Client has been disconnected from Memphis")
	return nil
}
