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
// limitations under the License.package server
package server

import (
	"context"
	"encoding/json"
	"memphis-broker/analytics"
	"memphis-broker/models"
	"strconv"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func killRelevantConnections(zombieConnections []primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateMany(context.TODO(),
		bson.M{"_id": bson.M{"$in": zombieConnections}},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		serv.Errorf("killRelevantConnections: " + err.Error())
		return err
	}

	return nil
}

func killProducersByConnections(connectionIds []primitive.ObjectID) error {
	_, err := producersCollection.UpdateMany(context.TODO(),
		bson.M{"connection_id": bson.M{"$in": connectionIds}},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		serv.Errorf("killProducersByConnections: " + err.Error())
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
		serv.Errorf("killConsumersByConnections: " + err.Error())
		return err
	}

	return nil
}

func (srv *Server) removeStaleStations() {
	var stations []models.Station
	cursor, err := stationsCollection.Find(nil, bson.M{"is_deleted": false})
	if err != nil {
		srv.Errorf("removeRedundantStations: " + err.Error())
	}

	if err = cursor.All(nil, &stations); err != nil {
		srv.Errorf("removeRedundantStations: " + err.Error())
	}

	for _, s := range stations {
		go func(srv *Server, s models.Station) {
			stationName, _ := StationNameFromStr(s.Name)
			_, err = srv.memphisStreamInfo(stationName.Intern())
			if IsNatsErr(err, JSStreamNotFoundErr) {
				srv.Warnf("removeRedundantStations: Found zombie station to delete: " + s.Name)
				_, err := stationsCollection.UpdateMany(nil,
					bson.M{"name": s.Name, "is_deleted": false},
					bson.M{"$set": bson.M{"is_deleted": true}})
				if err != nil {
					srv.Errorf("removeRedundantStations: " + err.Error())
				}
			}
		}(srv, s)
	}
}

func getActiveConnections() ([]models.Connection, error) {
	var connections []models.Connection
	cursor, err := connectionsCollection.Find(context.TODO(), bson.M{"is_active": true})
	if err != nil {
		return connections, err
	}
	if err = cursor.All(context.TODO(), &connections); err != nil {
		return connections, err
	}

	return connections, nil
}

// TODO to be deleted
func updateActiveProducersAndConsumers() {
	producersCount, err := producersCollection.CountDocuments(context.TODO(), bson.M{"is_active": true})
	if err != nil {
		serv.Warnf("updateActiveProducersAndConsumers: " + err.Error())
		return
	}

	consumersCount, err := consumersCollection.CountDocuments(context.TODO(), bson.M{"is_active": true})
	if err != nil {
		serv.Warnf("updateActiveProducersAndConsumers: " + err.Error())
		return
	}

	if producersCount > 0 || consumersCount > 0 {
		shouldSendAnalytics, _ := shouldSendAnalytics()
		if shouldSendAnalytics {
			param1 := analytics.EventParam{
				Name:  "active-producers",
				Value: strconv.Itoa(int(producersCount)),
			}
			param2 := analytics.EventParam{
				Name:  "active-consumers",
				Value: strconv.Itoa(int(consumersCount)),
			}
			analyticsParams := []analytics.EventParam{param1, param2}
			analytics.SendEventWithParams("", analyticsParams, "data-sent")
		}
	}
}

func aggregateClientConnections(s *Server) (map[string]string, error) {
	connectionIds := make(map[string]string)
	var lock sync.Mutex
	replySubject := CONN_STATUS_SUBJ + "_reply_" + s.memphis.nuid.Next()
	sub, err := s.subscribeOnGlobalAcc(replySubject, replySubject+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			var incomingConnIds map[string]string
			err := json.Unmarshal(msg, &incomingConnIds)
			if err != nil {
				s.Errorf("aggregateClientConnections: " + err.Error())
				return
			}

			for k := range incomingConnIds {
				lock.Lock()
				connectionIds[k] = ""
				lock.Unlock()
			}
		}(copyBytes(msg))
	})
	if err != nil {
		return nil, err
	}

	// send message to all brokers to get their connections
	s.sendInternalAccountMsgWithReply(s.GlobalAccount(), CONN_STATUS_SUBJ, replySubject, nil, _EMPTY_, true)
	timeout := time.After(50 * time.Second)
	<-timeout
	s.unsubscribeOnGlobalAcc(sub)
	return connectionIds, nil
}

func killFunc(s *Server) {
	connections, err := getActiveConnections()
	if err != nil {
		serv.Errorf("killFunc: getActiveConnections: " + err.Error())
		return
	}

	if len(connections) > 0 {
		var zombieConnections []primitive.ObjectID
		clientConnectionIds, err := aggregateClientConnections(s)
		if err != nil {
			serv.Errorf("killFunc: aggregateClientConnections: " + err.Error())
			return
		}
		for _, conn := range connections {
			if _, exist := clientConnectionIds[(conn.ID).Hex()]; exist { // existence check
				continue
			} else {
				zombieConnections = append(zombieConnections, conn.ID)
			}
		}

		if len(zombieConnections) > 0 {
			serv.Warnf("Zombie connections found, killing")
			err := killRelevantConnections(zombieConnections)
			if err != nil {
				serv.Errorf("killFunc: killRelevantConnections: " + err.Error())
			} else {
				err = killProducersByConnections(zombieConnections)
				if err != nil {
					serv.Errorf("killFunc: killProducersByConnections: " + err.Error())
				}

				err = killConsumersByConnections(zombieConnections)
				if err != nil {
					serv.Errorf("killFunc: killConsumersByConnections: " + err.Error())
				}
			}
		}
	}
	s.removeStaleStations()
}

func (s *Server) KillZombieResources() {
	count := 0
	for range time.Tick(time.Second * 20) {
		if s.JetStreamIsLeader() {
			break
		} else if count > 3 {
			return
		}
		count++
	}

	for range time.Tick(time.Second * 60) {
		s.Debugf("Killing Zombie resources iteration")
		killFunc(s)
		updateActiveProducersAndConsumers() // TODO to be deleted
	}
}
