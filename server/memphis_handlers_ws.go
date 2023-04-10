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
	"errors"
	"fmt"
	"memphis/db"
	"memphis/models"
	"strconv"
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
	memphisWS_Subj_AllSchemasData       = "get_all_schema_data"
	ws_updates_interval_sec             = 5
)

type memphisWSReqFiller func() (any, error)

func (s *Server) initWS() {
	ws := &s.memphis.ws
	ws.subscriptions = NewConcurrentMap[memphisWSReqFiller]()
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

func memphisWSLoop(s *Server, subs *concurrentMap[memphisWSReqFiller], quitCh chan struct{}) {
	ticker := time.NewTicker(ws_updates_interval_sec * time.Second)
	for {
		select {
		case <-ticker.C:
			keys, values := subs.Array()
			for i, updateFiller := range values {
				k := keys[i]
				replySubj := fmt.Sprintf(memphisWS_TemplSubj_Publish, k+"."+s.opts.ServerName)
				if !s.GlobalAccount().SubscriptionInterest(replySubj) {
					s.Debugf("removing memphis ws subscription %s", replySubj)
					subs.Delete(k)
					continue
				}
				update, err := updateFiller()
				if err != nil {
					if !IsNatsErr(err, JSStreamNotFoundErr) {
						s.Errorf("memphisWSLoop: " + err.Error())
					}
					continue
				}
				updateRaw, err := json.Marshal(update)
				if err != nil {
					s.Errorf("memphisWSLoop: " + err.Error())
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
			if _, ok := subscriptions.Load(filteredSubj); !ok {
				reqFiller, err := memphisWSGetReqFillerFromSubj(s, h, filteredSubj)
				if err != nil {
					s.Errorf("memphis websocket: " + err.Error())
					return
				}
				subscriptions.Add(filteredSubj, reqFiller)
			}

		default:
			s.Errorf("memphis websocket: invalid sub/unsub operation")
		}

		type brokerName struct {
			Name string `json:"name"`
		}

		broName := brokerName{s.opts.ServerName}
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
			return memphisWSGetStationOverviewDataAsync(s, h, stationName)
		}, nil

	case memphisWS_Subj_PoisonMsgJourneyData:
		poisonMsgId := tokenAt(subj, 2)
		poisonMsgIdInt, err := strconv.Atoi(poisonMsgId)
		if err != nil {
			return nil, err
		}
		if poisonMsgId == _EMPTY_ {
			return nil, errors.New("invalid poison msg id")
		}
		return func() (any, error) {
			return h.PoisonMsgs.GetDlsMessageDetailsById(poisonMsgIdInt, "poison")
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
	stations, totalMessages, totalDlsMsgs, err := h.Stations.GetAllStationsDetails(false)
	if err != nil {
		return models.MainOverviewData{}, nil
	}
	systemComponents, metricsEnabled, err := h.Monitoring.GetSystemComponents()
	if err != nil {
		return models.MainOverviewData{}, err
	}
	k8sEnv := true
	if configuration.DOCKER_ENV == "true" || configuration.LOCAL_CLUSTER_ENV {
		k8sEnv = false
	}
	brokersThroughputs, err := h.Monitoring.GetBrokersThroughputs()
	if err != nil {
		return models.MainOverviewData{}, err
	}
	return models.MainOverviewData{
		TotalStations:     len(stations),
		TotalMessages:     totalMessages,
		TotalDlsMessages:  totalDlsMsgs,
		SystemComponents:  systemComponents,
		Stations:          stations,
		K8sEnv:            k8sEnv,
		BrokersThroughput: brokersThroughputs,
		MetricsEnabled:    metricsEnabled,
	}, nil
}

func memphisWSGetStationOverviewDataAsync(s *Server, h *Handlers, stationNameString string) (map[string]any, error) {
	stationOverviewData, stationName, err := GetStationOverviewData(h, stationNameString)
	if err != nil {
		return map[string]any{}, err
	}

	schema, err := h.Schemas.GetSchemaByStationName(stationName)

	if err != nil && err != ErrNoSchema {
		return map[string]any{}, err
	}

	var response map[string]any

	if err == ErrNoSchema { // non native stations will always reach this point
		if !stationOverviewData.Station.IsNative {
			cp, dp, cc, dc := getFakeProdsAndConsForPreview()
			response = map[string]any{
				"connected_producers":           cp,
				"disconnected_producers":        dp,
				"deleted_producers":             stationOverviewData.Producers.DeletedProducers,
				"connected_cgs":                 cc,
				"disconnected_cgs":              stationOverviewData.Cgs.DisconnectedCgs,
				"deleted_cgs":                   dc,
				"total_messages":                stationOverviewData.MessagesSample.TotalMessages,
				"average_message_size":          stationOverviewData.MessagesSample.AvgMsgSize,
				"audit_logs":                    stationOverviewData.AuditLogs,
				"messages":                      stationOverviewData.MessagesSample.Messages,
				"poison_messages":               stationOverviewData.DlsMessages.PoisonMessages,
				"schema_failed_messages":        stationOverviewData.DlsMessages.SchemaFailedMessages,
				"tags":                          stationOverviewData.Tags,
				"leader":                        stationOverviewData.LeaderAndFollowers.Leader,
				"followers":                     stationOverviewData.LeaderAndFollowers.Followers,
				"schema":                        struct{}{},
				"idempotency_window_in_ms":      stationOverviewData.Station.IdempotencyWindow,
				"dls_configuration_poison":      stationOverviewData.Station.DlsConfigurationPoison,
				"dls_configuration_schemaverse": stationOverviewData.Station.DlsConfigurationSchemaverse,
				"total_dls_messages":            stationOverviewData.DlsMessages.TotalDlsAmount,
			}
		} else {
			response = map[string]any{
				"connected_producers":           stationOverviewData.Producers.ConnectedProducers,
				"disconnected_producers":        stationOverviewData.Producers.DisconnectedProducers,
				"deleted_producers":             stationOverviewData.Producers.DeletedProducers,
				"connected_cgs":                 stationOverviewData.Cgs.ConnectedCgs,
				"disconnected_cgs":              stationOverviewData.Cgs.DisconnectedCgs,
				"deleted_cgs":                   stationOverviewData.Cgs.DeletedCgs,
				"total_messages":                stationOverviewData.MessagesSample.TotalMessages,
				"average_message_size":          stationOverviewData.MessagesSample.AvgMsgSize,
				"audit_logs":                    stationOverviewData.AuditLogs,
				"messages":                      stationOverviewData.MessagesSample.Messages,
				"poison_messages":               stationOverviewData.DlsMessages.PoisonMessages,
				"schema_failed_messages":        stationOverviewData.DlsMessages.SchemaFailedMessages,
				"tags":                          stationOverviewData.Tags,
				"leader":                        stationOverviewData.LeaderAndFollowers.Leader,
				"followers":                     stationOverviewData.LeaderAndFollowers.Followers,
				"schema":                        struct{}{},
				"idempotency_window_in_ms":      stationOverviewData.Station.IdempotencyWindow,
				"dls_configuration_poison":      stationOverviewData.Station.DlsConfigurationPoison,
				"dls_configuration_schemaverse": stationOverviewData.Station.DlsConfigurationSchemaverse,
				"total_dls_messages":            stationOverviewData.DlsMessages.TotalDlsAmount,
			}
		}

		return response, nil
	}

	_, schemaVersion, err := db.GetSchemaVersionByNumberAndID(stationOverviewData.Station.SchemaVersionNumber, schema.ID)
	if err != nil {
		return map[string]any{}, err
	}
	updatesAvailable := !schemaVersion.Active
	schemaDetails := models.StationOverviewSchemaDetails{
		SchemaName:       schema.Name,
		VersionNumber:    stationOverviewData.Station.SchemaVersionNumber,
		UpdatesAvailable: updatesAvailable,
		SchemaType:       schema.Type,
	}

	response = map[string]any{
		"connected_producers":           stationOverviewData.Producers.ConnectedProducers,
		"disconnected_producers":        stationOverviewData.Producers.DisconnectedProducers,
		"deleted_producers":             stationOverviewData.Producers.DeletedProducers,
		"connected_cgs":                 stationOverviewData.Cgs.ConnectedCgs,
		"disconnected_cgs":              stationOverviewData.Cgs.DisconnectedCgs,
		"deleted_cgs":                   stationOverviewData.Cgs.DeletedCgs,
		"total_messages":                stationOverviewData.MessagesSample.TotalMessages,
		"average_message_size":          stationOverviewData.MessagesSample.AvgMsgSize,
		"audit_logs":                    stationOverviewData.AuditLogs,
		"messages":                      stationOverviewData.MessagesSample.Messages,
		"poison_messages":               stationOverviewData.DlsMessages.PoisonMessages,
		"schema_failed_messages":        stationOverviewData.DlsMessages.SchemaFailedMessages,
		"tags":                          stationOverviewData.Tags,
		"leader":                        stationOverviewData.LeaderAndFollowers.Leader,
		"followers":                     stationOverviewData.LeaderAndFollowers.Followers,
		"schema":                        schemaDetails,
		"idempotency_window_in_ms":      stationOverviewData.Station.IdempotencyWindow,
		"dls_configuration_poison":      stationOverviewData.Station.DlsConfigurationPoison,
		"dls_configuration_schemaverse": stationOverviewData.Station.DlsConfigurationSchemaverse,
		"total_dls_messages":            stationOverviewData.DlsMessages.TotalDlsAmount,
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

func memphisWSGetSystemLogs(h *Handlers, logLevel string) (models.SystemLogsResponse, error) {
	const amount = 100
	const timeout = 3 * time.Second
	filterSubjectSuffix := ""
	switch logLevel {
	case "err":
		filterSubjectSuffix = syslogsErrSubject
	case "warn":
		filterSubjectSuffix = syslogsWarnSubject
	case "info":
		filterSubjectSuffix = syslogsInfoSubject
	default:
		filterSubjectSuffix = syslogsExternalSubject
	}

	filterSubject := "$memphis_syslogs.*." + filterSubjectSuffix

	return h.Monitoring.S.GetSystemLogs(amount, timeout, true, 0, filterSubject, false)
}
