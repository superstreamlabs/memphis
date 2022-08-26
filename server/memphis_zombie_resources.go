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
package server

import (
	"context"
	"memphis-broker/models"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const pingGrace = 5 * time.Second

func killRelevantConnections() ([]models.Connection, error) {
	duration := -(serv.opts.PingInterval + pingGrace)
	lastAllowedTime := time.Now().Local().Add(duration)

	var connections []models.Connection
	cursor, err := connectionsCollection.Find(context.TODO(), bson.M{"is_active": true, "last_ping": bson.M{"$lt": lastAllowedTime}})
	if err != nil {
		serv.Errorf("killRelevantConnections error: " + err.Error())
		return connections, err
	}

	if err = cursor.All(context.TODO(), &connections); err != nil {
		serv.Errorf("killRelevantConnections error: " + err.Error())
		return connections, err
	}

	_, err = connectionsCollection.UpdateMany(context.TODO(),
		bson.M{"is_active": true, "last_ping": bson.M{"$lt": lastAllowedTime}},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		serv.Errorf("KillConnections error: " + err.Error())
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
		serv.Errorf("killProducersByConnections error: " + err.Error())
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
		serv.Errorf("killConsumersByConnections error: " + err.Error())
		return err
	}

	return nil
}

func removeOldPoisonMsgs() error {
	filter := bson.M{"creation_date": bson.M{"$lte": (time.Now().Add(-(time.Hour * time.Duration(configuration.POISON_MSGS_RETENTION_IN_HOURS))))}}
	_, err := poisonMessagesCollection.DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}

func KillZombieResources(wg *sync.WaitGroup) {
	for range time.Tick(time.Second * 30) {
		connections, err := killRelevantConnections()
		if err != nil {
			serv.Errorf("KillZombieResources error: " + err.Error())
		} else if len(connections) > 0 {
			serv.Warnf("zombie connection found, killing %v", connections)
			var connectionIds []primitive.ObjectID
			for _, con := range connections {
				connectionIds = append(connectionIds, con.ID)
			}

			err = killProducersByConnections(connectionIds)
			if err != nil {
				serv.Errorf("KillZombieResources error: " + err.Error())
			}

			err = killConsumersByConnections(connectionIds)
			if err != nil {
				serv.Errorf("KillZombieResources error: " + err.Error())
			}
		}

		if err != nil {
			serv.Errorf("KillZombieResources error: " + err.Error())
		}

		err = removeOldPoisonMsgs()
		if err != nil {
			serv.Errorf("KillZombieResources error: " + err.Error())
		}
	}

	defer wg.Done()
}
