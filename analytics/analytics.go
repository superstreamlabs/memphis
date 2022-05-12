package analytics

import (
	"context"
	"memphis-control-plane/config"
	"memphis-control-plane/db"
	"memphis-control-plane/logger"
	"memphis-control-plane/models"

	"github.com/lightstep/otel-launcher-go/launcher"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var configuration = config.GetConfig()
var ls launcher.Launcher

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
	deploymentId, err := getSystemKey("deployment_id")
	if err != nil {
		return err
	}

	analytics, err := getSystemKey("analytics")
	if err != nil {
		return err
	}

	if analytics.Value == "true" {
		ls = launcher.ConfigureOpentelemetry(
			launcher.WithServiceName("memphis-"+deploymentId.Value),
			launcher.WithAccessToken(configuration.ANALYTICS_TOKEN),
		)
		logger.Info("Analytics initialized")
	}

	return nil
}

func StartEvent(eventName string) trace.Span {
	tracer := otel.Tracer("example")
	_, span := tracer.Start(context.TODO(), eventName)
	return span
}

func Close() {
	analytics, _ := getSystemKey("analytics")
	if analytics.Value == "true" {
		ls.Shutdown()
	}
}
