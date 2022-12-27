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
	"fmt"
	"memphis-broker/models"
	"strings"
	"time"
)

const (
	memphisWS_SubscribeMsg              = "SUB"
	memphisWS_UnsubscribeMsg            = "UNSUB"
	memphisWS_Subj_Subs                 = "$memphis_ws_subs.>"
	memphisWs_Cgroup_Subs               = "$memphis_ws_subs_cg"
	memphisWS_TemplSubj_Publish         = "$memphis_ws_pubs.%s"
	memphisWS_Subj_MainOverviewData     = "main_overview_data"
	memphisWS_Subj_StationOverviewData  = "station_overview_data"
	memphisWS_Subj_PoisonMsgJourneyData = "poison_message_journey_data"
	memphisWS_Subj_AllStationsData      = "get_all_stations_data"
	memphisWS_Subj_SysLogsData          = "syslogs_data"
	memphisWS_Subj_SysLogsDataInf       = "syslogs_data.info"
	memphisWS_Subj_SysLogsDataWrn       = "syslogs_data.warn"
	memphisWS_Subj_SysLogsDataErr       = "syslogs_data.err"
	memphisWS_Subj_AllSchemasData       = "get_all_schema_data"
)

type memphisWSReqFiller func() (any, error)

func (s *Server) initWS() {
	ws := &s.memphis.ws
	ws.subscriptions = make(map[string]memphisWSReqFiller)
	handlers := Handlers{
		Producers:  ProducersHandler{S: s},
		Consumers:  ConsumersHandler{S: s},
		AuditLogs:  AuditLogsHandler{},
		Stations:   StationsHandler{S: s},
		Monitoring: MonitoringHandler{S: s},
		PoisonMsgs: PoisonMessagesHandler{S: s},
		Schemas:    SchemasHandler{S: s},
	}

	s.queueSubscribe(memphisWS_Subj_Subs,
		memphisWs_Cgroup_Subs,
		s.createWSRegistrationHandler(&handlers))

	go memphisWSLoop(s, ws.subscriptions, ws.quitCh)
}

func memphisWSLoop(s *Server, subs map[string]memphisWSReqFiller, quitCh chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			for k, updateFiller := range subs {
				replySubj := fmt.Sprintf(memphisWS_TemplSubj_Publish, k+"."+configuration.SERVER_NAME)
				if !s.GlobalAccount().SubscriptionInterest(replySubj) {
					s.Debugf("removing memphis ws subscription %s", replySubj)
					delete(subs, k)
					continue
				}
				update, err := updateFiller()
				if err != nil {
					s.Errorf(err.Error())
					continue
				}
				updateRaw, err := json.Marshal(update)
				if err != nil {
					s.Errorf(err.Error())
					continue
				}

				s.respondOnGlobalAcc(replySubj, updateRaw)
			}
		case <-quitCh:
			ticker.Stop()
			return
		}
	}
}

func tokensFromToEnd(subject string, index uint8) string {
	ti, start := uint8(1), 0
	for i := 0; i < len(subject); i++ {
		if subject[i] == btsep {
			if ti == index {
				return subject[start:]
			}
			start = i + 1
			ti++
		}
	}
	if ti == index {
		return subject[start:]
	}
	return _EMPTY_
}

func (s *Server) createWSRegistrationHandler(h *Handlers) simplifiedMsgHandler {
	return func(c *client, subj, reply string, msg []byte) {
		s.Debugf("memphisWS registration - %s,%s", subj, string(msg))
		subscriptions := s.memphis.ws.subscriptions
		filteredSubj := tokensFromToEnd(subj, 2)
		trimmedMsg := strings.TrimSuffix(string(msg), "\r\n")
		switch trimmedMsg {
		case memphisWS_SubscribeMsg:
			if _, ok := subscriptions[filteredSubj]; !ok {
				reqFiller, err := memphisWSGetReqFillerFromSubj(s, h, filteredSubj)
				if err != nil {
					s.Errorf("memphis websocket: " + err.Error())
					return
				}
				subscriptions[filteredSubj] = reqFiller
			}

		default:
			s.Errorf("memphis websocket: invalid sub/unsub operation")
		}
		if configuration.SERVER_NAME == "" {
			configuration.SERVER_NAME = "broker"
		}

		type brokerName struct {
			Name string `json:"name"`
		}

		broName := brokerName{configuration.SERVER_NAME}
		serverName, err := json.Marshal(broName)

		if err != nil {
			s.Errorf("memphis websocket: " + err.Error())
			return
		}

		s.sendInternalAccountMsg(s.GlobalAccount(), reply, serverName)
	}
}

func unwrapHandlersFunc[T interface{}](f func(*Handlers) (T, error), h *Handlers) func() (any, error) {
	return func() (any, error) {
		return f(h)
	}
}

func memphisWSGetReqFillerFromSubj(s *Server, h *Handlers, subj string) (memphisWSReqFiller, error) {
	subjectHead := tokenAt(subj, 1)
	switch subjectHead {
	case memphisWS_Subj_MainOverviewData:
		return unwrapHandlersFunc(memphisWSGetMainOverviewData, h), nil

	case memphisWS_Subj_StationOverviewData:
		stationName := strings.Join(strings.Split(subj, ".")[1:], ".")
		if stationName == _EMPTY_ {
			return nil, errors.New("invalid station name")
		}
		return func() (any, error) {
			return memphisWSGetStationOverviewData(s, h, stationName)
		}, nil

	case memphisWS_Subj_PoisonMsgJourneyData:
		poisonMsgId := tokenAt(subj, 2)
		if poisonMsgId == _EMPTY_ {
			return nil, errors.New("invalid poison msg id")
		}
		return func() (any, error) {
			return h.Stations.GetDlsMessageJourneyDetails(poisonMsgId)
		}, nil

	case memphisWS_Subj_AllStationsData:
		return unwrapHandlersFunc(memphisWSGetStationsOverviewData, h), nil

	case memphisWS_Subj_SysLogsData:
		logLevel := tokenAt(subj, 2)
		return func() (any, error) {
			return memphisWSGetSystemLogs(h, logLevel)
		}, nil

	case memphisWS_Subj_AllSchemasData:
		return unwrapHandlersFunc(memphisWSGetSchemasOverviewData, h), nil
	default:
		return nil, errors.New("invalid subject")
	}
}

func memphisWSGetMainOverviewData(h *Handlers) (models.MainOverviewData, error) {
	stations, err := h.Stations.GetAllStationsDetails()
	if err != nil {
		return models.MainOverviewData{}, nil
	}
	totalMessages, err := h.Stations.GetTotalMessagesAcrossAllStations()
	if err != nil {
		return models.MainOverviewData{}, err
	}
	systemComponents, err := h.Monitoring.GetSystemComponents()
	if err != nil {
		return models.MainOverviewData{}, err
	}

	return models.MainOverviewData{
		TotalStations:    len(stations),
		TotalMessages:    totalMessages,
		SystemComponents: systemComponents,
		Stations:         stations,
	}, nil
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
		return map[string]any{}, errors.New("Station " + stationName + " does not exist")
	}

	connectedProducers, disconnectedProducers, deletedProducers, err := h.Producers.GetProducersByStation(station)
	if err != nil {
		return map[string]any{}, err
	}
	connectedCgs, disconnectedCgs, deletedCgs := make([]models.Cg, 0), make([]models.Cg, 0), make([]models.Cg, 0)
	// Only native stations have CGs
	if station.IsNative {
		connectedCgs, disconnectedCgs, deletedCgs, err = h.Consumers.GetCgsByStation(sn, station)
		if err != nil {
			return map[string]any{}, err
		}
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

	poisonMessages, schemaFailMessages, totalDlsAmount, err := h.PoisonMsgs.GetDlsMsgsByStationLight(station)
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
			"connected_producers":      connectedProducers,
			"disconnected_producers":   disconnectedProducers,
			"deleted_producers":        deletedProducers,
			"connected_cgs":            connectedCgs,
			"disconnected_cgs":         disconnectedCgs,
			"deleted_cgs":              deletedCgs,
			"total_messages":           totalMessages,
			"average_message_size":     avgMsgSize,
			"audit_logs":               auditLogs,
			"messages":                 messages,
			"poison_messages":          poisonMessages,
			"schema_fail_messages":     schemaFailMessages,
			"tags":                     tags,
			"leader":                   leader,
			"followers":                followers,
			"schema":                   struct{}{},
			"idempotency_window_in_ms": station.IdempotencyWindow,
			"dls_configuration":        station.DlsConfiguration,
			"total_dls_messages":       totalDlsAmount,
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
		"connected_producers":      connectedProducers,
		"disconnected_producers":   disconnectedProducers,
		"deleted_producers":        deletedProducers,
		"connected_cgs":            connectedCgs,
		"disconnected_cgs":         disconnectedCgs,
		"deleted_cgs":              deletedCgs,
		"total_messages":           totalMessages,
		"average_message_size":     avgMsgSize,
		"audit_logs":               auditLogs,
		"messages":                 messages,
		"poison_messages":          poisonMessages,
		"schema_fail_messages":     schemaFailMessages,
		"tags":                     tags,
		"leader":                   leader,
		"followers":                followers,
		"schema":                   schemaDetails,
		"idempotency_window_in_ms": station.IdempotencyWindow,
		"dls_configuration":        station.DlsConfiguration,
		"total_dls_messages":       totalDlsAmount,
	}

	return response, nil
}

func memphisWSGetSchemasOverviewData(h *Handlers) ([]models.ExtendedSchema, error) {
	schemas, err := h.Schemas.GetAllSchemasDetails()
	if err != nil {
		return schemas, err
	}
	return schemas, nil
}

func memphisWSGetStationsOverviewData(h *Handlers) ([]models.ExtendedStationDetails, error) {
	stations, err := h.Stations.GetStationsDetails()
	if err != nil {
		return stations, err
	}
	return stations, nil
}

func memphisWSGetSystemLogs(h *Handlers, filterSubjectSuffix string) (models.SystemLogsResponse, error) {
	const amount = 100
	const timeout = 3 * time.Second
	filterSubject := ""
	if filterSubjectSuffix != "" {
		filterSubject = "$memphis_syslogs.*." + filterSubjectSuffix
	}
	return h.Monitoring.S.GetSystemLogs(amount, timeout, true, 0, filterSubject, false)
}
