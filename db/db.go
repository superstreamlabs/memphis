package db

import (
	"context"
	"memphis-control-plane/config"
	"memphis-control-plane/logger"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var configuration = config.GetConfig()

const (
	dbOperationTimeout = 20
)

func initializeDbConnection() (*mongo.Client, context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.TODO(), dbOperationTimeout*time.Second)

	var clientOptions *options.ClientOptions
	if configuration.DOCKER_ENV != "" {
		clientOptions = options.Client().ApplyURI(configuration.MONGO_URL).SetConnectTimeout(dbOperationTimeout * time.Second)
	} else {
		auth := options.Credential{
			AuthSource: configuration.DB_NAME,
			Username:   configuration.MONGO_USER,
			Password:   configuration.MONGO_PASS,
		}
		clientOptions = options.Client().ApplyURI(configuration.MONGO_URL).SetAuth(auth).SetConnectTimeout(dbOperationTimeout * time.Second)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Error("Failed to create Mongodb client: " + err.Error())
		panic("Failed to create Mongodb client: " + err.Error())
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error("Failed to create Mongo DB client: " + err.Error())
		panic("Failed to create Mongo DB client: " + err.Error())
	}

	return client, ctx, cancel
}

func GetCollection(collectionName string) *mongo.Collection {
	var collection *mongo.Collection = Client.Database(configuration.DB_NAME).Collection(collectionName)
	return collection
}

func Close() {
	defer Cancel()
	defer func() {
		if err := Client.Disconnect(Ctx); err != nil {
			logger.Error("Failed to close Mongodb client: " + err.Error())
			panic("Failed to close Mongodb client: " + err.Error())
		}
	}()
}

var Client, Ctx, Cancel = initializeDbConnection()
