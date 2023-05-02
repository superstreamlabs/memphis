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

type memphisWSReqFiller func(tenantName string) (any, error)
type memphisWSReqTenantsToFiller struct {
	tenants map[string]bool
	filler  memphisWSReqFiller
}

func (s *Server) initWS(tenantName string) {
	ws := &s.memphis.ws
	ws.subscriptions = NewConcurrentMap[memphisWSReqTenantsToFiller]()
	handlers := Handlers{
		Producers:  ProducersHandler{S: s},
		Consumers:  ConsumersHandler{S: s},
		AuditLogs:  AuditLogsHandler{},
		Stations:   StationsHandler{S: s},
		Monitoring: MonitoringHandler{S: s},
		PoisonMsgs: PoisonMessagesHandler{S: s},
		Schemas:    SchemasHandler{S: s},
	}

	s.queueSubscribe(tenantName, memphisWS_Subj_Subs,
		memphisWs_Cgroup_Subs,
		s.createWSRegistrationHandler(&handlers))

	go memphisWSLoop(tenantName, s, ws.subscriptions, ws.quitCh)
}

func deleteTenantFromSub(tenantName string, subs *concurrentMap[memphisWSReqTenantsToFiller], key string) {
	subs.Lock()
	defer subs.Unlock()
	if f, ok := subs.m[key]; ok {
		delete(f.tenants, tenantName)
		subs.m[key] = f
	}
}

func addTenantToSub(tenantName string, subs *concurrentMap[memphisWSReqTenantsToFiller], key string) error {
	subs.Lock()
	defer subs.Unlock()
	if f, ok := subs.m[key]; ok {
		f.tenants[tenantName] = true
		subs.m[key] = f
	} else {
		return errors.New("addTenantToSub: sub not found")
	}
	return nil
}

func memphisWSLoop(tenantName string, s *Server, subs *concurrentMap[memphisWSReqTenantsToFiller], quitCh chan struct{}) {
	ticker := time.NewTicker(ws_updates_interval_sec * time.Second)
	for {
		select {
		case <-ticker.C:
			keys, values := subs.Array()
			for i, updateFiller := range values {
				k := keys[i]
				replySubj := fmt.Sprintf(memphisWS_TemplSubj_Publish, k+"."+s.opts.ServerName)
				for tenant := range updateFiller.tenants {
					if tenant == tenantName {
						acc, err := s.lookupAccount(tenant)
						if err != nil {
							s.Errorf("memphisWSLoop: tenant " + tenant + ": " + err.Error())
							continue
						}
						if !acc.SubscriptionInterest(replySubj) {
							s.Debugf("removing tenant "+tenant+" ws subscription %s", replySubj)
							deleteTenantFromSub(tenant, subs, k)
							continue
						}
						update, err := updateFiller.filler(tenant)
						if err != nil {
							if !IsNatsErr(err, JSStreamNotFoundErr) && !strings.Contains(err.Error(), "not exist") {
								s.Errorf("memphisWSLoop: tenant " + tenant + ": " + err.Error())
							}
							continue
						}
						updateRaw, err := json.Marshal(update)
						if err != nil {
							s.Errorf("memphisWSLoop: " + err.Error())
							continue
						}

						s.sendInternalAccountMsg(acc, replySubj, updateRaw)
					}
				}
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
		tenantName := c.Account().GetName()
		s.Debugf("memphisWS registration - %s,%s", subj, string(msg))
		subscriptions := s.memphis.ws.subscriptions
		filteredSubj := tokensFromToEnd(subj, 2)
		trimmedMsg := strings.TrimSuffix(string(msg), "\r\n")
		switch trimmedMsg {
		case memphisWS_SubscribeMsg:
			if _, ok := subscriptions.Load(filteredSubj); !ok {
				reqFiller, err := memphisWSGetReqFillerFromSubj(s, h, filteredSubj, tenantName)
				if err != nil {
					s.Errorf("memphis websocket: " + err.Error())
					return
				}
				subscriptions.Add(filteredSubj, memphisWSReqTenantsToFiller{tenants: map[string]bool{c.Account().GetName(): true}, filler: reqFiller})
			} else {
				err := addTenantToSub(tenantName, subscriptions, filteredSubj)
				if err != nil {
					s.Errorf("memphis websocket: " + err.Error())
				}
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

		account, err := s.lookupAccount(tenantName)
		if err != nil {
			s.Errorf("memphis websocket: " + err.Error())
			return
		}
		s.sendInternalAccountMsg(account, reply, serverName)
	}
}

func unwrapHandlersFunc[T interface{}](f func(*Handlers, string) (T, error), h *Handlers, tenantName string) func(string) (any, error) {
	return func(string) (any, error) {
		return f(h, tenantName)
	}
}

func memphisWSGetReqFillerFromSubj(s *Server, h *Handlers, subj string, tenantName string) (memphisWSReqFiller, error) {
	subjectHead := tokenAt(subj, 1)
	switch subjectHead {
	case memphisWS_Subj_MainOverviewData:
		return unwrapHandlersFunc(memphisWSGetMainOverviewData, h, tenantName), nil

	case memphisWS_Subj_StationOverviewData:
		stationName := strings.Join(strings.Split(subj, ".")[1:], ".")
		if stationName == _EMPTY_ {
			return nil, errors.New("invalid station name")
		}
		return func(string) (any, error) {
			return memphisWSGetStationOverviewData(s, h, stationName, tenantName)
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
		return func(string) (any, error) {
			return h.PoisonMsgs.GetDlsMessageDetailsById(poisonMsgIdInt, "poison", tenantName)
		}, nil

	case memphisWS_Subj_AllStationsData:
		return unwrapHandlersFunc(memphisWSGetStationsOverviewData, h, tenantName), nil

	case memphisWS_Subj_SysLogsData:
		logLevel := tokenAt(subj, 2)
		logSource := tokenAt(subj, 3)
		return func(string) (any, error) {
			return memphisWSGetSystemLogs(h, logLevel, logSource)
		}, nil

	case memphisWS_Subj_AllSchemasData:
		return unwrapHandlersFunc(memphisWSGetSchemasOverviewData, h, tenantName), nil
	default:
		return nil, errors.New("invalid subject")
	}
}

func memphisWSGetMainOverviewData(h *Handlers, tenantName string) (models.MainOverviewData, error) {
	response, err := h.Monitoring.getMainOverviewDataDetails(tenantName)
	if err != nil {
		return models.MainOverviewData{}, err
	}
	return response, nil
}

func memphisWSGetStationOverviewData(s *Server, h *Handlers, stationName string, tenantName string) (map[string]any, error) {
	sn, err := StationNameFromStr(stationName)
	if err != nil {
		return map[string]any{}, err
	}
	exist, station, err := db.GetStationByName(sn.Ext(), tenantName)
	if err != nil {
		return map[string]any{}, err
	}
	if !exist {
		return map[string]any{}, errors.New("Station " + stationName + " does not exist")
	}

	connectedProducers, disconnectedProducers, deletedProducers := make([]models.ExtendedProducer, 0), make([]models.ExtendedProducer, 0), make([]models.ExtendedProducer, 0)
	if station.IsNative {
		connectedProducers, disconnectedProducers, deletedProducers, err = h.Producers.GetProducersByStation(station)
		if err != nil {
			return map[string]any{}, err
		}
	}

	auditLogs, err := h.AuditLogs.GetAuditLogsByStation(station.Name, tenantName)
	if err != nil {
		return map[string]any{}, err
	}
	totalMessages, err := h.Stations.GetTotalMessages(station.TenantName, station.Name)
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

	connectedCgs, disconnectedCgs, deletedCgs := make([]models.Cg, 0), make([]models.Cg, 0), make([]models.Cg, 0)
	// Only native stations have CGs
	if station.IsNative {
		connectedCgs, disconnectedCgs, deletedCgs, err = h.Consumers.GetCgsByStation(sn, station)
		if err != nil {
			return map[string]any{}, err
		}
	}

	tags, err := h.Tags.GetTagsByEntityWithID("station", station.ID)
	if err != nil {
		return map[string]any{}, err
	}
	leader, followers, err := h.Stations.GetLeaderAndFollowers(station)
	if err != nil {
		return map[string]any{}, err
	}

	schema, err := h.Schemas.GetSchemaByStationName(sn, station.TenantName)

	if err != nil && err != ErrNoSchema {
		return map[string]any{}, err
	}

	var response map[string]any

	if err == ErrNoSchema { // non native stations will always reach this point
		if !station.IsNative {
			cp, dp, cc, dc := getFakeProdsAndConsForPreview()
			response = map[string]any{
				"connected_producers":           cp,
				"disconnected_producers":        dp,
				"deleted_producers":             deletedProducers,
				"connected_cgs":                 cc,
				"disconnected_cgs":              disconnectedCgs,
				"deleted_cgs":                   dc,
				"total_messages":                totalMessages,
				"average_message_size":          avgMsgSize,
				"audit_logs":                    auditLogs,
				"messages":                      messages,
				"poison_messages":               poisonMessages,
				"schema_failed_messages":        schemaFailMessages,
				"tags":                          tags,
				"leader":                        leader,
				"followers":                     followers,
				"schema":                        struct{}{},
				"idempotency_window_in_ms":      station.IdempotencyWindow,
				"dls_configuration_poison":      station.DlsConfigurationPoison,
				"dls_configuration_schemaverse": station.DlsConfigurationSchemaverse,
				"total_dls_messages":            totalDlsAmount,
			}
		} else {
			response = map[string]any{
				"connected_producers":           connectedProducers,
				"disconnected_producers":        disconnectedProducers,
				"deleted_producers":             deletedProducers,
				"connected_cgs":                 connectedCgs,
				"disconnected_cgs":              disconnectedCgs,
				"deleted_cgs":                   deletedCgs,
				"total_messages":                totalMessages,
				"average_message_size":          avgMsgSize,
				"audit_logs":                    auditLogs,
				"messages":                      messages,
				"poison_messages":               poisonMessages,
				"schema_failed_messages":        schemaFailMessages,
				"tags":                          tags,
				"leader":                        leader,
				"followers":                     followers,
				"schema":                        struct{}{},
				"idempotency_window_in_ms":      station.IdempotencyWindow,
				"dls_configuration_poison":      station.DlsConfigurationPoison,
				"dls_configuration_schemaverse": station.DlsConfigurationSchemaverse,
				"total_dls_messages":            totalDlsAmount,
			}
		}

		return response, nil
	}

	_, schemaVersion, err := db.GetSchemaVersionByNumberAndID(station.SchemaVersionNumber, schema.ID)
	if err != nil {
		return map[string]any{}, err
	}
	updatesAvailable := !schemaVersion.Active
	schemaDetails := models.StationOverviewSchemaDetails{
		SchemaName:       schema.Name,
		VersionNumber:    station.SchemaVersionNumber,
		UpdatesAvailable: updatesAvailable,
		SchemaType:       schema.Type,
	}

	response = map[string]any{
		"connected_producers":           connectedProducers,
		"disconnected_producers":        disconnectedProducers,
		"deleted_producers":             deletedProducers,
		"connected_cgs":                 connectedCgs,
		"disconnected_cgs":              disconnectedCgs,
		"deleted_cgs":                   deletedCgs,
		"total_messages":                totalMessages,
		"average_message_size":          avgMsgSize,
		"audit_logs":                    auditLogs,
		"messages":                      messages,
		"poison_messages":               poisonMessages,
		"schema_failed_messages":        schemaFailMessages,
		"tags":                          tags,
		"leader":                        leader,
		"followers":                     followers,
		"schema":                        schemaDetails,
		"idempotency_window_in_ms":      station.IdempotencyWindow,
		"dls_configuration_poison":      station.DlsConfigurationPoison,
		"dls_configuration_schemaverse": station.DlsConfigurationSchemaverse,
		"total_dls_messages":            totalDlsAmount,
	}

	return response, nil
}

func memphisWSGetSchemasOverviewData(h *Handlers, tenantName string) ([]models.ExtendedSchema, error) {
	schemas, err := h.Schemas.GetAllSchemasDetails(tenantName)
	if err != nil {
		return schemas, err
	}
	return schemas, nil
}

func memphisWSGetStationsOverviewData(h *Handlers, tenantName string) ([]models.ExtendedStationDetails, error) {
	stations, err := h.Stations.GetStationsDetails(tenantName)
	if err != nil {
		return stations, err
	}
	return stations, nil
}

func memphisWSGetSystemLogs(h *Handlers, logLevel, logSource string) (models.SystemLogsResponse, error) {
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

	if filterSubjectSuffix != _EMPTY_ {
		if logSource != "empty" && logLevel != "external" {
			filterSubject = fmt.Sprintf("%s.%s.%s", syslogsStreamName, logSource, filterSubjectSuffix)
		} else if logSource != "empty" && logLevel == "external" {
			filterSubject = fmt.Sprintf("%s.%s.%s.%s", syslogsStreamName, logSource, "extern", ">")
		} else {
			filterSubject = fmt.Sprintf("%s.%s.%s", syslogsStreamName, "*", filterSubjectSuffix)
		}
	}
	return h.Monitoring.S.GetSystemLogs(amount, timeout, true, 0, filterSubject, false)
}
