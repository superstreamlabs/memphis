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
		serv.Errorf("killRelevantConnections error: " + err.Error())
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

func (srv *Server) removeRedundantStations() {
	var stations []models.Station
	cursor, err := stationsCollection.Find(nil, bson.M{"is_deleted": false})
	if err != nil {
		srv.Errorf("removeRedundantStations error: " + err.Error())
	}

	if err = cursor.All(nil, &stations); err != nil {
		srv.Errorf("removeRedundantStations error: " + err.Error())
	}

	for _, s := range stations {
		go func(srv *Server, s models.Station) {
			stationName, _ := StationNameFromStr(s.Name)
			_, err = srv.memphisStreamInfo(stationName.Intern())
			if IsNatsErr(err, JSStreamNotFoundErr) {
				srv.Warnf("Found zombie station to delete: " + s.Name)
				_, err := stationsCollection.UpdateMany(nil,
					bson.M{"name": s.Name, "is_deleted": false},
					bson.M{"$set": bson.M{"is_deleted": true}})
				if err != nil {
					srv.Errorf("removeRedundantStations error: " + err.Error())
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
		serv.Warnf("updateActiveProducersAndConsumers error: " + err.Error())
		return
	}
	consumersCount, err := consumersCollection.CountDocuments(context.TODO(), bson.M{"is_active": true})
	if err != nil {
		serv.Warnf("updateActiveProducersAndConsumers error: " + err.Error())
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

func killFunc(s *Server) {
	connections, err := getActiveConnections()
	if err != nil {
		serv.Errorf("killFunc error: " + err.Error())
		return
	}

	var zombieConnections []primitive.ObjectID
	var lock sync.Mutex
	wg := sync.WaitGroup{}
	wg.Add(len(connections))
	for _, conn := range connections {
		go func(s *Server, conn models.Connection, wg *sync.WaitGroup, lock *sync.Mutex) {
			respCh := make(chan []byte)
			msg := (conn.ID).Hex()
			reply := CONN_STATUS_SUBJ + "_reply" + s.memphis.nuid.Next()

			sub, err  := s.subscribeOnGlobalAcc(reply, reply+"_sid", func(_ *client, subject, reply string, msg []byte) {
				go func(msg []byte) { respCh <- msg }(copyBytes(msg))
			})
			if err != nil {
				s.Errorf("killFunc error: " + err.Error())
				wg.Done()
				return
			}

			s.sendInternalAccountMsgWithReply(s.GlobalAccount(), CONN_STATUS_SUBJ, reply, nil, msg, true)
			timeout := time.After(30 * time.Second)
			select {
			case <-respCh:
				wg.Done()
				return
			case <-timeout:
				lock.Lock()
				zombieConnections = append(zombieConnections, conn.ID)
				lock.Unlock()
			}
			sub.close()
			wg.Done()
		}(s, conn, &wg, &lock)
	}
	wg.Wait()

	if len(zombieConnections) > 0 {
		serv.Warnf("Zombie connections found, killing")
		err := killRelevantConnections(zombieConnections)
		if err != nil {
			serv.Errorf("killFunc error: " + err.Error())
		} else {
			err = killProducersByConnections(zombieConnections)
			if err != nil {
				serv.Errorf("killFunc error: " + err.Error())
			}

			err = killConsumersByConnections(zombieConnections)
			if err != nil {
				serv.Errorf("killFunc error: " + err.Error())
			}
		}
	}

	err = removeOldPoisonMsgs()
	if err != nil {
		serv.Errorf("killFunc error: " + err.Error())
	}

	s.removeRedundantStations()
}

func (s *Server) KillZombieResources() {
	s.Debugf("Killing Zombie resources iteration")
	killFunc(s)

	for range time.Tick(time.Second * 30) {
		s.Debugf("Killing Zombie resources iteration")
		killFunc(s)
		updateActiveProducersAndConsumers() // TODO to be deleted
	}
}
