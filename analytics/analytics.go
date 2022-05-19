package analytics

import (
	"context"
	"memphis-control-plane/config"
	"memphis-control-plane/db"
	"memphis-control-plane/models"

	"github.com/lightstep/otel-launcher-go/launcher"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

var configuration = config.GetConfig()
var ls launcher.Launcher
var loginsCounter metric.Int64Counter
var installationsCounter metric.Int64Counter
var deploymentId string

func getSystemKey(key string) (models.SystemKey, error) {
	systemKeysCollection := db.GetCollection("system_keys")

	filter := bson.M{"key": key}
	var systemKey models.SystemKey
	err := systemKeysCollection.FindOne(context.TODO(), filter).Decode(&systemKey)
	if err == mongo.ErrNoDocuments {
		return systemKey, nil
	} else if err != nil {
		return systemKey, err
	}
	return systemKey, nil
}

func InitializeAnalytics() error {
	deployment, err := getSystemKey("deployment_id")
	if err != nil {
		return err
	}
	deploymentId = deployment.Value

	analytics, err := getSystemKey("analytics")
	if err != nil {
		return err
	}

	if analytics.Value == "true" {
		ls = launcher.ConfigureOpentelemetry(
			launcher.WithServiceName("memphis"),
			launcher.WithAccessToken(configuration.ANALYTICS_TOKEN),
		)
	}

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

	return nil
}

func IncrementInstallationsCounter() {
	installationsCounter.Add(context.TODO(), 1)
}

func IncrementLoginsCounter() {
	loginsCounter.Add(context.TODO(), 1, attribute.String("deployment_id", deploymentId))
}

func Close() {
	analytics, _ := getSystemKey("analytics")
	if analytics.Value == "true" {
		ls.Shutdown()
	}
}
