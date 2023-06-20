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
	"encoding/json"
	"memphis/analytics"
	"memphis/db"
	"memphis/models"
	"strconv"
	"sync"
	"time"
)

func (srv *Server) removeStaleStations() {
	stations, err := db.GetActiveStations()
	if err != nil {
		srv.Errorf("removeStaleStations: %v", err.Error())
	}
	for _, s := range stations {
		go func(srv *Server, s models.Station) {
			stationName, _ := StationNameFromStr(s.Name)
			_, err = srv.memphisStreamInfo(s.TenantName, stationName.Intern())
			if IsNatsErr(err, JSStreamNotFoundErr) {
				srv.Warnf("[tenant: %v]removeStaleStations: Found zombie station to delete: %v", s.TenantName, s.Name)
				err := removeStationResources(srv, s, false)
				if err != nil {
					srv.Errorf("[tenant: %v]removeStaleStations at removeStationResources: %v", s.TenantName, err.Error())
				}
				err = db.DeleteStation(s.Name, s.TenantName)
				if err != nil {
					srv.Errorf("[tenant: %v]removeStaleStations at DeleteStation: %v", s.TenantName, err.Error())
				}
			}
		}(srv, s)
	}
}

func updateSystemLiveness() {
	stationsHandler := StationsHandler{S: serv}
	stations, totalMessages, totalDlsMsgs, err := stationsHandler.GetAllStationsDetails(false, "")
	if err != nil {
		serv.Warnf("updateSystemLiveness: %v", err.Error())
		return
	}

	producersCount, err := db.CountAllActiveProudcers()
	if err != nil {
		serv.Warnf("updateSystemLiveness: %v", err.Error())
		return
	}

	consumersCount, err := db.CountAllActiveConsumers()
	if err != nil {
		serv.Warnf("updateSystemLiveness: %v", err.Error())
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
	analytics.SendEventWithParams("", "", analyticsParams, "system-is-up")
}

func aggregateClientConnections(s *Server) (map[string]string, error) {
	connectionIds := make(map[string]string)
	var lock sync.Mutex
	replySubject := CONN_STATUS_SUBJ + "_reply_" + s.memphis.nuid.Next()
	sub, err := s.subscribeOnAcc(s.GlobalAccount(), replySubject, replySubject+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			var incomingConnIds map[string]string
			err := json.Unmarshal(msg, &incomingConnIds)
			if err != nil {
				s.Errorf("aggregateClientConnections: %v", err.Error())
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
	s.unsubscribeOnAcc(s.GlobalAccount(), sub)
	return connectionIds, nil
}

func killFunc(s *Server) {
	connections, err := db.GetActiveConnections()
	if err != nil {
		serv.Errorf("killFunc: GetActiveConnections: %v", err.Error())
		return
	}

	if len(connections) > 0 {
		var zombieConnections []string
		clientConnectionIds, err := aggregateClientConnections(s)
		if err != nil {
			serv.Errorf("killFunc: aggregateClientConnections: %v", err.Error())
			return
		}
		for _, conn := range connections {
			if _, exist := clientConnectionIds[conn.ID]; exist { // existence check
				continue
			} else {
				zombieConnections = append(zombieConnections, conn.ID)
			}
		}

		if len(zombieConnections) > 0 {
			serv.Warnf("Zombie connections found, killing")
			err := db.KillRelevantConnections(zombieConnections)
			if err != nil {
				serv.Errorf("killFunc: killRelevantConnections: %v", err.Error())
			}
			err = db.KillProducersByConnections(zombieConnections)
			if err != nil {
				serv.Errorf("killFunc: killProducersByConnections: %v", err.Error())
			}
			err = db.KillConsumersByConnections(zombieConnections)
			if err != nil {
				serv.Errorf("killFunc: killConsumersByConnections: %v", err.Error())
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
