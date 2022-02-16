package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strech-server/config"
	"strech-server/logger"
	"time"
)

var configuration = config.GetConfig()

const (
	dbOperationTimeout = 5
)

func getConnection() (*mongo.Client, context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), dbOperationTimeout*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(configuration.MONGO_URL))
	if err != nil {
		logger.Error("Failed to create Mongo DB client: " + err.Error())
		panic("Failed to create Mongo DB client: " + err.Error())
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error("Failed to create Mongo DB client: " + err.Error())
		panic("Failed to create Mongo DB client: " + err.Error())
	}
	logger.Info("Connected to MongoDB")

	return client, ctx, cancel
}

func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database(configuration.DB_NAME).Collection(collectionName)

	return collection
}

func Close(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

var Client, Ctx, Cancel = getConnection()
