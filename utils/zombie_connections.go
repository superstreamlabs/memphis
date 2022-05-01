package utils

import (
	"context"
	"memphis-control-plane/db"
	"memphis-control-plane/logger"
	"memphis-control-plane/models"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var connectionsCollection *mongo.Collection = db.GetCollection("connections")
var producersCollection *mongo.Collection = db.GetCollection("producers")
var consumersCollection *mongo.Collection = db.GetCollection("consumers")

func killRelevantConnections() ([]models.Connection, error) {
	timeWithoutPing := primitive.NewDateTimeFromTime(time.Now().Add(time.Duration(-configuration.PING_INTERVAL_MS-5000) * time.Millisecond))

	var connections []models.Connection
	cursor, err := connectionsCollection.Find(context.TODO(), bson.M{"is_active": true, "last_ping": bson.M{"$lt": timeWithoutPing}})
	if err != nil {
		logger.Error("killRelevantConnections error: " + err.Error())
		return connections, err
	}

	if err = cursor.All(context.TODO(), &connections); err != nil {
		logger.Error("killRelevantConnections error: " + err.Error())
		return connections, err
	}

	_, err = connectionsCollection.UpdateMany(context.TODO(),
		bson.M{"is_active": true, "last_ping": bson.M{"$lt": timeWithoutPing}},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		logger.Error("KillConnections error: " + err.Error())
		return connections, err
	}

	var connectionIds []primitive.ObjectID
	for _, con := range connections {
		connectionIds = append(connectionIds, con.ID)
	}

	return connections, nil
}

func killProducersByConnections(connectionIds []primitive.ObjectID) error {
	_, err := producersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": bson.M{"$in": connectionIds}},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		logger.Error("killProducersByConnections error: " + err.Error())
		return err
	}

	return nil
}

func killConsumersByConnections(connectionIds []primitive.ObjectID) error {
	_, err := consumersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": bson.M{"$in": connectionIds}},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		logger.Error("killConsumersByConnections error: " + err.Error())
		return err
	}

	return nil
}

func KillZombieConnections(wg *sync.WaitGroup) {
	for range time.Tick(time.Second * 30) {
		connections, err := killRelevantConnections()
		if err != nil {
			logger.Error("KillZombieConnections error: " + err.Error())
		} else if len(connections) > 0 {
			var connectionIds []primitive.ObjectID
			for _, con := range connections {
				connectionIds = append(connectionIds, con.ID)
			}

			err = killProducersByConnections(connectionIds)
			if err != nil {
				logger.Error("KillZombieConnections error: " + err.Error())
			}

			err = killConsumersByConnections(connectionIds)
			if err != nil {
				logger.Error("KillZombieConnections error: " + err.Error())
			}
		}
	}

	defer wg.Done()
}
