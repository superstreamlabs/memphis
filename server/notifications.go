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
	"errors"

	"github.com/memphisdev/memphis/models"
)

func SendNotification(tenantName string, title string, message string, msgType string) error {
	for k, f := range NotificationFunctionsMap {
		switch k {
		case "slack":
			if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(tenantName); ok {
				if slackIntegration, ok := tenantInetgrations["slack"].(models.SlackIntegration); ok {
					if slackIntegration.Properties[msgType] {
						err := f.(func(models.SlackIntegration, string, string) error)(slackIntegration, title, message)
						if err != nil {
							return err
						}
					}
				}
			}
		case "discord":
			if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(tenantName); ok {
				if discordIntegration, ok := tenantInetgrations["discord"].(models.DiscordIntegration); ok {
					if discordIntegration.Properties[msgType] {
						err := f.(func(models.DiscordIntegration, string, string) error)(discordIntegration, title, message)
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

func shouldSendNotification(tenantName string, alertType string) bool {
	if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(tenantName); ok {
		if slackIntegration, ok := tenantInetgrations["slack"].(models.SlackIntegration); ok {
			if slackIntegration.Properties[alertType] {
				return true
			}
		}

		if discordIntegration, ok := tenantInetgrations["discord"].(models.DiscordIntegration); ok {
			if discordIntegration.Properties[alertType] {
				return true
			}
		}
	}
	return false
}
