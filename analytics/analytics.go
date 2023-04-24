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
package analytics

import (
	"memphis/conf"
	"memphis/db"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/posthog/posthog-go"
)

type EventParam struct {
	Name  string `json:"name"`
	Value string `json:"value" binding:"required"`
}

const MEMPHIS_GLOBAL_ACCOUNT = "memphis"

var configuration = conf.GetConfig()
var deploymentId string
var memphisVersion string
var AnalyticsClient posthog.Client

func InitializeAnalytics(analyticsToken, memphisV string) error {
	memphisVersion = memphisV
	exist, deployment, err := db.GetSystemKey("deployment_id", MEMPHIS_GLOBAL_ACCOUNT)
	if err != nil {
		return err
	} else if !exist {
		uid, err := uuid.NewV4()
		if err != nil {
			return err
		}
		deploymentId = uid.String()
		err = db.InsertSystemKey("deployment_id", deploymentId, MEMPHIS_GLOBAL_ACCOUNT)
		if err != nil {
			return err
		}
	} else {
		deploymentId = deployment.Value
	}

	exist, _, err = db.GetSystemKey("analytics", MEMPHIS_GLOBAL_ACCOUNT)
	if err != nil {
		return err
	} else if !exist {
		value := ""
		if configuration.ANALYTICS == "true" {
			value = "true"
		} else {
			value = "false"
		}

		err = db.InsertSystemKey("analytics", value, MEMPHIS_GLOBAL_ACCOUNT)
		if err != nil {
			return err
		}
	}

	client, err := posthog.NewWithConfig(analyticsToken, posthog.Config{Endpoint: "https://app.posthog.com"})
	if err != nil {
		return err
	}

	AnalyticsClient = client
	return nil
}

func Close() {
	_, analytics, _ := db.GetSystemKey("analytics", MEMPHIS_GLOBAL_ACCOUNT)
	if analytics.Value == "true" {
		AnalyticsClient.Close()
	}
}

func SendEvent(userId, eventName string) {
	var distinctId string
	if configuration.DEV_ENV != "" {
		distinctId = "dev"
		// } else if configuration.SANDBOX_ENV == "true" {
		// 	distinctId = "sandbox" + "-" + userId
	} else {
		distinctId = deploymentId + "-" + userId
	}
	distinctId = strings.TrimSuffix(distinctId, "-")

	p := posthog.NewProperties()
	p.Set("memphis-version", memphisVersion)

	go AnalyticsClient.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      eventName,
		Properties: p,
	})
}

func SendEventWithParams(userId string, params []EventParam, eventName string) {
	var distinctId string
	if configuration.DEV_ENV != "" {
		distinctId = "dev"
		// } else if configuration.SANDBOX_ENV == "true" {
		// 	distinctId = "sandbox" + "-" + userId
	} else {
		distinctId = deploymentId + "-" + userId
	}
	distinctId = strings.TrimSuffix(distinctId, "-")

	p := posthog.NewProperties()
	for _, param := range params {
		p.Set(param.Name, param.Value)
	}
	p.Set("memphis-version", memphisVersion)

	go AnalyticsClient.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      eventName,
		Properties: p,
	})
}

func SendErrEvent(origin, errMsg string) {
	distinctId := deploymentId
	if configuration.DEV_ENV != "" {
		distinctId = "dev"
	}
	// } else if configuration.SANDBOX_ENV == "true" {
	// 	distinctId = "sandbox"
	// }

	p := posthog.NewProperties()
	p.Set("err_log", errMsg)
	p.Set("err_source", origin)
	p.Set("memphis-version", memphisVersion)
	AnalyticsClient.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      "error",
		Properties: p,
	})
}
