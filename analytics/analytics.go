// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package analytics
package analytics

import (
	"context"
	"memphis-broker/conf"
	"memphis-broker/db"
	"memphis-broker/models"

	"github.com/posthog/posthog-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type EventParam struct {
	Name  string `json:"name"`
	Value string `json:"value" binding:"required"`
}

var configuration = conf.GetConfig()
var systemKeysCollection *mongo.Collection
var deploymentId string
var analyticsFlag string
var AnalyticsClient posthog.Client

func InitializeAnalytics(c *mongo.Client) error {
	systemKeysCollection = db.GetCollection("system_keys", c)
	deployment, err := getSystemKey("deployment_id")
	if err == mongo.ErrNoDocuments {
		deploymentId = primitive.NewObjectID().Hex()
		deploymentKey := models.SystemKey{
			ID:    primitive.NewObjectID(),
			Key:   "deployment_id",
			Value: deploymentId,
		}

		_, err = systemKeysCollection.InsertOne(context.TODO(), deploymentKey)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		deploymentId = deployment.Value
	}

	analytics, err := getSystemKey("analytics")
	if err == mongo.ErrNoDocuments {
		var analyticsKey models.SystemKey
		if configuration.ANALYTICS == "true" {
			analyticsKey = models.SystemKey{
				ID:    primitive.NewObjectID(),
				Key:   "analytics",
				Value: "true",
			}
		} else {
			analyticsKey = models.SystemKey{
				ID:    primitive.NewObjectID(),
				Key:   "analytics",
				Value: "false",
			}
		}

		_, err = systemKeysCollection.InsertOne(context.TODO(), analyticsKey)
		if err != nil {
			return err
		}
		analyticsFlag = configuration.ANALYTICS
	} else if err != nil {
		return err
	} else {
		analyticsFlag = analytics.Value
	}

	client, err := posthog.NewWithConfig(configuration.ANALYTICS_TOKEN, posthog.Config{Endpoint: "https://app.posthog.com"})
	if err != nil {
		return err
	}

	AnalyticsClient = client
	return nil
}

func getSystemKey(key string) (models.SystemKey, error) {
	filter := bson.M{"key": key}
	var systemKey models.SystemKey
	err := systemKeysCollection.FindOne(context.TODO(), filter).Decode(&systemKey)
	if err != nil {
		return systemKey, err
	}
	return systemKey, nil
}

func Close() {
	analytics, _ := getSystemKey("analytics")
	if analytics.Value == "true" {
		AnalyticsClient.Close()
	}
}

func SendEvent(userId, eventName string) {
	var distinctId string
	if configuration.DEV_ENV != "" {
		distinctId = "dev"
	} else if configuration.SANDBOX_ENV == "true" {
		distinctId = "sandbox" + "-" + userId
	} else {
		distinctId = deploymentId + "-" + userId
	}

	go AnalyticsClient.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      eventName,
	})
}

func SendEventWithParams(userId string, params []EventParam, eventName string) {
	var distinctId string
	if configuration.DEV_ENV != "" {
		distinctId = "dev"
	} else if configuration.SANDBOX_ENV == "true" {
		distinctId = "sandbox" + "-" + userId
	} else {
		distinctId = deploymentId + "-" + userId
	}

	p := posthog.NewProperties()
	for _, param := range params {
		p.Set(param.Name, param.Value)
	}

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
	} else if configuration.SANDBOX_ENV == "true" {
		distinctId = "sandbox"
	}

	p := posthog.NewProperties()
	p.Set("err_log", errMsg)
	p.Set("err_source", origin)
	AnalyticsClient.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      "error",
		Properties: p,
	})
}
