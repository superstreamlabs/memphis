// Copyright 2021-2022 The Memphis Authors
// Licensed under the GNU General Public License v3.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package analytics

import (
	"context"
	"memphis-broker/config"
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

var configuration = config.GetConfig()
var systemKeysCollection = db.GetCollection("system_keys")
var ls launcher.Launcher
var loginsCounter metric.Int64Counter
var installationsCounter metric.Int64Counter
var stationsCounter metric.Int64Counter
var producersCounter metric.Int64Counter
var consumersCounter metric.Int64Counter
var deploymentId string
var analyticsFlag string

func getSystemKey(key string) (models.SystemKey, error) {
	filter := bson.M{"key": key}
	var systemKey models.SystemKey
	err := systemKeysCollection.FindOne(context.TODO(), filter).Decode(&systemKey)
	if err != nil {
		return systemKey, err
	}
	return systemKey, nil
}

func InitializeAnalytics() error {
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

	if analyticsFlag == "true" {
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

		loginsCounter, err = Meter.NewInt64Counter(
			"Logins",
			metric.WithUnit("0"),
			metric.WithDescription("Counting the number of logins to Memphis"),
		)

		stationsCounter, err = Meter.NewInt64Counter(
			"Stations",
			metric.WithUnit("0"),
			metric.WithDescription("Counting the number of stations"),
		)

		producersCounter, err = Meter.NewInt64Counter(
			"Producers",
			metric.WithUnit("0"),
			metric.WithDescription("Counting the number of producers"),
		)

		consumersCounter, err = Meter.NewInt64Counter(
			"Consumers",
			metric.WithUnit("0"),
			metric.WithDescription("Counting the number of consumers"),
		)
	}

	return nil
}

func IncrementInstallationsCounter() {
	installationsCounter.Add(context.TODO(), 1)
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

func Close() {
	analytics, _ := getSystemKey("analytics")
	if analytics.Value == "true" {
		ls.Shutdown()
	}
}
