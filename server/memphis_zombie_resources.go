// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"context"
	"encoding/json"
	"memphis/analytics"
	"memphis/models"
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

func updateSystemLiveness() {
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		stationsHandler := StationsHandler{S: serv}
		stations, totalMessages, totalDlsMsgs, err := stationsHandler.GetAllStationsDetails()
		if err != nil {
			serv.Warnf("updateSystemLiveness: " + err.Error())
			return
		}

		producersCount, err := producersCollection.CountDocuments(context.TODO(), bson.M{"is_active": true})
		if err != nil {
			serv.Warnf("updateSystemLiveness: " + err.Error())
			return
		}

		consumersCount, err := consumersCollection.CountDocuments(context.TODO(), bson.M{"is_active": true})
		if err != nil {
			serv.Warnf("updateSystemLiveness: " + err.Error())
			return
		}

		param1 := analytics.EventParam{
			Name:  "total-messages",
			Value: strconv.Itoa(int(totalMessages)),
		}
		param2 := analytics.EventParam{
			Name:  "total-dls-messages",
			Value: strconv.Itoa(int(totalDlsMsgs)),
		}
		param3 := analytics.EventParam{
			Name:  "total-stations",
			Value: strconv.Itoa(len(stations)),
		}
		param4 := analytics.EventParam{
			Name:  "active-producers",
			Value: strconv.Itoa(int(producersCount)),
		}
		param5 := analytics.EventParam{
			Name:  "active-consumers",
			Value: strconv.Itoa(int(consumersCount)),
		}
		analyticsParams := []analytics.EventParam{param1, param2, param3, param4, param5}
		analytics.SendEventWithParams("", analyticsParams, "system-is-up")
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
			}
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

	s.removeStaleStations()
}

func (s *Server) KillZombieResources() {
	if s.JetStreamIsClustered() {
		count := 0
		for range time.Tick(time.Second * 20) {
			if s.JetStreamIsLeader() {
				break
			} else if count > 3 {
				return
			}
			count++
		}
	}

	count := 0
	firstIteration := true
	for range time.Tick(time.Minute * 1) {
		s.Debugf("Killing Zombie resources iteration")
		killFunc(s)

		if firstIteration || count == 1*60 { // once in 1 hour
			updateSystemLiveness()
			count = 0

		}
		firstIteration = false
		count++
	}
}
