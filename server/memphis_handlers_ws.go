// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// HTTP://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server
package server

import (
	"encoding/json"
	"errors"
	"memphis-broker/models"
	"time"
)

const (
	memphisWS_SubscribeMsg                 = "SUB"
	memphisWS_UnsubscribeMsg               = "UNSUB"
	memphisWS_subsSubject                  = "$memphis_ws_subs"
	memphisWS_sub_StationOverviewData      = "$memphis_sub__station_overview_data"
	memphisWS_template_StationOverviewData = "station_overview_data.%s"
)

type memphisWSSubscription struct {
	refCount   int
	updateFunc func() map[string]any
}

func (s *Server) initWS() {
	ws := &s.memphis.ws
	ws.subscriptions = make(map[string]memphisWSSubscription)
	handlers := Handlers{
		Producers:  ProducersHandler{S: s},
		Consumers:  ConsumersHandler{S: s},
		AuditLogs:  AuditLogsHandler{},
		Stations:   StationsHandler{S: s},
		Monitoring: MonitoringHandler{S: s},
		PoisonMsgs: PoisonMessagesHandler{S: s},
		Schemas:    SchemasHandler{S: s},
	}

	s.queueSubscribe(memphisWS_subsSubject,
		memphisWS_subsSubject,
		s.createWSRegistrationHandler(&handlers))

	go memphisWSLoop(s, ws.subscriptions, ws.quitCh)
}

func memphisWSLoop(s *Server, subs map[string]memphisWSSubscription, quitCh chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			for k, v := range subs {
				update := v.updateFunc()
				updateRaw, err := json.Marshal(update)
				if err != nil {
					s.Errorf(err.Error())
				}
				s.respondOnGlobalAcc(k, updateRaw)
			}
		case <-quitCh:
			ticker.Stop()
			return
		}
	}
}

func (s *Server) createWSRegistrationHandler(h *Handlers) simplifiedMsgHandler {
	return func(c *client, subj, reply string, msg []byte) {
		subscriptions := s.memphis.ws.subscriptions
		switch string(msg) {
		case memphisWS_SubscribeMsg:
			if sub, ok := subscriptions[subj]; ok {
				sub.refCount += 1
				subscriptions[subj] = sub
				return
			}
			subscriptions[subj] = memphisWSParseSubscriptionSubject(s, h, subj)

		case memphisWS_UnsubscribeMsg:
			sub := subscriptions[subj]
			sub.refCount -= 1
			if sub.refCount <= 0 {
				delete(subscriptions, subj)
				return
			}
			subscriptions[subj] = sub
		default:
			s.Errorf("memphis websocket: invalid sub/unsub operation")
		}
	}
}

func memphisWSParseSubscriptionSubject(s *Server, h *Handlers, subj string) memphisWSSubscription {
	subjectHead := tokenAt(subj, 1)
	switch subjectHead {
	case memphisWS_sub_StationOverviewData:
		stationName := tokenAt(subj, 2)
		if stationName == _EMPTY_ {
			s.Errorf("memphis websocket invalid station name")
		}

		return memphisWSSubscription{
			refCount: 1,
			updateFunc: func() map[string]any {
				result, err := memphisWSGetStationOverviewData(s, h, stationName)
				if err != nil {
					s.Errorf(err.Error())
				}
				return result
			},
		}
	default:
		s.Errorf("memphis websocket: invalid subject")
	}
	return memphisWSSubscription{}
}

func memphisWSGetStationOverviewData(s *Server, h *Handlers, stationName string) (map[string]any, error) {
	sn, err := StationNameFromStr(stationName)
	if err != nil {
		return map[string]any{}, err
	}

	exist, station, err := IsStationExist(sn)
	if err != nil {
		return map[string]any{}, err
	}
	if !exist {
		return map[string]any{}, errors.New("Station does not exist")
	}

	connectedProducers, disconnectedProducers, deletedProducers, err := h.Producers.GetProducersByStation(station)
	if err != nil {
		return map[string]any{}, err
	}
	connectedCgs, disconnectedCgs, deletedCgs, err := h.Consumers.GetCgsByStation(sn, station)
	if err != nil {
		return map[string]any{}, err
	}
	auditLogs, err := h.AuditLogs.GetAuditLogsByStation(station)
	if err != nil {
		return map[string]any{}, err
	}
	totalMessages, err := h.Stations.GetTotalMessages(station.Name)
	if err != nil {
		return map[string]any{}, err
	}
	avgMsgSize, err := h.Stations.GetAvgMsgSize(station)
	if err != nil {
		return map[string]any{}, err
	}

	messagesToFetch := 1000
	messages, err := h.Stations.GetMessages(station, messagesToFetch)
	if err != nil {
		return map[string]any{}, err
	}

	poisonMessages, err := h.PoisonMsgs.GetPoisonMsgsByStation(station)
	if err != nil {
		return map[string]any{}, err
	}

	tags, err := h.Tags.GetTagsByStation(station.ID)
	if err != nil {
		return map[string]any{}, err
	}
	leader, followers, err := h.Stations.GetLeaderAndFollowers(station)
	if err != nil {
		return map[string]any{}, err
	}

	schema, err := h.Schemas.GetSchemaByStationName(sn)

	if err != nil && err != ErrNoSchema {
		return map[string]any{}, err
	}

	var response map[string]any

	if err == ErrNoSchema {
		response = map[string]any{
			"connected_producers":    connectedProducers,
			"disconnected_producers": disconnectedProducers,
			"deleted_producers":      deletedProducers,
			"connected_cgs":          connectedCgs,
			"disconnected_cgs":       disconnectedCgs,
			"deleted_cgs":            deletedCgs,
			"total_messages":         totalMessages,
			"average_message_size":   avgMsgSize,
			"audit_logs":             auditLogs,
			"messages":               messages,
			"poison_messages":        poisonMessages,
			"tags":                   tags,
			"leader":                 leader,
			"followers":              followers,
			"schema":                 struct{}{},
		}
		return response, nil
	}

	schemaVersion, err := h.Schemas.GetSchemaVersion(station.Schema.VersionNumber, schema.ID)
	if err != nil {
		return map[string]any{}, err
	}
	updatesAvailable := !schemaVersion.Active
	schemaDetails := models.StationOverviewSchemaDetails{SchemaName: schema.Name, VersionNumber: station.Schema.VersionNumber, UpdatesAvailable: updatesAvailable}

	response = map[string]any{
		"connected_producers":    connectedProducers,
		"disconnected_producers": disconnectedProducers,
		"deleted_producers":      deletedProducers,
		"connected_cgs":          connectedCgs,
		"disconnected_cgs":       disconnectedCgs,
		"deleted_cgs":            deletedCgs,
		"total_messages":         totalMessages,
		"average_message_size":   avgMsgSize,
		"audit_logs":             auditLogs,
		"messages":               messages,
		"poison_messages":        poisonMessages,
		"tags":                   tags,
		"leader":                 leader,
		"followers":              followers,
		"schema":                 schemaDetails,
	}

	return response, nil
}
