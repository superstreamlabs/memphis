// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
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

var configuration = conf.GetConfig()
var systemKeysCollection *mongo.Collection
var deploymentId string
var analyticsFlag string
var AnalyticsClient posthog.Client

func InitializeAnalytics(c *mongo.Client) error {
	systemKeysCollection = db.GetCollection("system_keys", c)
	deployment, err := getSystemKey("deployment_id")
	if err == mongo.ErrNoDocuments {
		deploymentId := primitive.NewObjectID().Hex()
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
	distinctId := deploymentId + "-" + userId
	AnalyticsClient.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      eventName,
	})
}
