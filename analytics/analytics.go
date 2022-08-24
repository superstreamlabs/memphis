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

	"github.com/lightstep/otel-launcher-go/launcher"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"

	"go.opentelemetry.io/otel/metric"

	"go.opentelemetry.io/otel/metric/global"
)

var configuration = conf.GetConfig()
var systemKeysCollection *mongo.Collection
var ls launcher.Launcher
var loginsCounter metric.Int64Counter
var installationsCounter metric.Int64Counter
var nextStepsCounter metric.Int64Counter
var stationsCounter metric.Int64Counter
var producersCounter metric.Int64Counter
var consumersCounter metric.Int64Counter
var disableAnalyticsCounter metric.Int64Counter
var deploymentId string
var analyticsFlag string

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

	ls = launcher.ConfigureOpentelemetry(
		launcher.WithServiceName("memphis"),
		launcher.WithAccessToken(configuration.ANALYTICS_TOKEN),
	)

	var Meter = global.GetMeterProvider().Meter("memphis")
	installationsCounter, err = Meter.NewInt64Counter(
		"Installations",
		metric.WithUnit("0"),
		metric.WithDescription("Counting the number of installations of Memphis"),
	)
	if err != nil {
		return err
	}

	nextStepsCounter, err = Meter.NewInt64Counter(
		"NextSteps",
		metric.WithUnit("0"),
		metric.WithDescription("Counting the number of users complete the next steps wizard in the UI"),
	)
	if err != nil {
		return err
	}

	loginsCounter, err = Meter.NewInt64Counter(
		"Logins",
		metric.WithUnit("0"),
		metric.WithDescription("Counting the number of logins to Memphis"),
	)
	if err != nil {
		return err
	}

	stationsCounter, err = Meter.NewInt64Counter(
		"Stations",
		metric.WithUnit("0"),
		metric.WithDescription("Counting the number of stations"),
	)
	if err != nil {
		return err
	}

	producersCounter, err = Meter.NewInt64Counter(
		"Producers",
		metric.WithUnit("0"),
		metric.WithDescription("Counting the number of producers"),
	)
	if err != nil {
		return err
	}

	consumersCounter, err = Meter.NewInt64Counter(
		"Consumers",
		metric.WithUnit("0"),
		metric.WithDescription("Counting the number of consumers"),
	)
	if err != nil {
		return err
	}

	disableAnalyticsCounter, err = Meter.NewInt64Counter(
		"DisableAnalytics",
		metric.WithUnit("0"),
		metric.WithDescription("Counting the number of disable analytics events"),
	)
	if err != nil {
		return err
	}

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

func IncrementInstallationsCounter() {
	installationsCounter.Add(context.TODO(), 1, attribute.String("deployment_id", deploymentId))
}

func IncrementNextStepsCounter() {
	nextStepsCounter.Add(context.TODO(), 1, attribute.String("deployment_id", deploymentId))
}

func IncrementLoginsCounter() {
	loginsCounter.Add(context.TODO(), 1, attribute.String("deployment_id", deploymentId))
}

func IncrementStationsCounter() {
	stationsCounter.Add(context.TODO(), 1, attribute.String("deployment_id", deploymentId))
}

func IncrementProducersCounter() {
	producersCounter.Add(context.TODO(), 1, attribute.String("deployment_id", deploymentId))
}

func IncrementConsumersCounter() {
	consumersCounter.Add(context.TODO(), 1, attribute.String("deployment_id", deploymentId))
}

func IncrementDisableAnalyticsCounter() {
	disableAnalyticsCounter.Add(context.TODO(), 1, attribute.String("deployment_id", deploymentId))
}

func Close() {
	analytics, _ := getSystemKey("analytics")
	if analytics.Value == "true" {
		ls.Shutdown()
	}
}
