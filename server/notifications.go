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
	"fmt"
	"github.com/memphisdev/memphis/models"
	"time"
)

const (
	slackIntegrationName = "slack"
)

type NotificationMsg struct {
	TenantName string    `json:"tenantName"`
	Title      string    `json:"title"`
	Message    string    `json:"message"`
	MsgType    string    `json:"msgType"`
	Time       time.Time `json:"time"`
}

type NotificationMsgWithReply struct {
	NotificationMsg *NotificationMsg
	ReplySubject    string
}

func (s *Server) SendNotification(tenantName string, title string, message string, msgType string) error {
	for k := range NotificationFunctionsMap {
		switch k {
		case slackIntegrationName:
			if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(tenantName); ok {
				if slackIntegration, ok := tenantInetgrations[slackIntegrationName].(models.SlackIntegration); ok {
					if slackIntegration.Properties[msgType] {
						// TODO: if the stream doesn't exist save the messages in buffer
						if !NOTIFICATIONS_BUFFER_STREAM_CREATED {
							return nil
						}

						// TODO: do we need msg-id here? if yes - what's the best way to generate it? hash title?
						if tenantName == "" {
							tenantName = serv.MemphisGlobalAccountString()
						}
						notificationMsg := NotificationMsg{
							TenantName: tenantName,
							Title:      title,
							Message:    message,
							MsgType:    msgType,
							Time:       time.Now(),
						}

						err := saveSlackNotificationToQueue(s, notificationsStreamName, tenantName, &notificationMsg)
						if err != nil {
							return err
						}
					}
				}
			}
		default:
			return errors.New("failed sending notification: unsupported integration")
		}
	}
	return nil
}

func saveSlackNotificationToQueue(s *Server, subject, tenantName string, notificationMsg *NotificationMsg) error {
	msg, err := json.Marshal(notificationMsg)
	if err != nil {
		return err
	}
	err = s.sendInternalAccountMsgWithEcho(s.MemphisGlobalAccount(), subject, msg)
	if err != nil {
		return fmt.Errorf("SendNotification (tenant %s): %w", tenantName, err)
	}

	return nil
}

func shouldSendNotification(tenantName string, alertType string) bool {
	if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(tenantName); ok {
		if slackIntegration, ok := tenantInetgrations["slack"].(models.SlackIntegration); ok {
			if slackIntegration.Properties[alertType] {
				return true
			}
		}
	}
	return false
}
