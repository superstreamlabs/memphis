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
	"strconv"
	"strings"
	"time"

	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"

	"github.com/gofrs/uuid"
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
	memphisWS_Subj_GetSystemMessages    = "get_system_messages"
	memphisWS_subj_GetAsyncTasks        = "get_async_tasks"
	ws_updates_interval_sec             = 20
	memphisWS_subj_GetAllFunctions      = "get_all_functions"
	memphisWS_subj_GetGraphOverview     = "get_graph_overview"
	memphisWS_subj_GetFunctionsOverview = "get_functions_overview"
)

type memphisWSReqFiller func(tenantName string) (any, error)
type memphisWSReqTenantsToFiller struct {
	tenants map[string]memphisWSReqFiller
}

func (s *Server) initWS() {
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

	s.queueSubscribe(s.MemphisGlobalAccountString(), memphisWS_Subj_Subs,
		memphisWs_Cgroup_Subs,
		s.createWSRegistrationHandler(&handlers))

	go memphisWSLoop(s, ws.subscriptions, ws.quitCh)
}

func deleteTenantFromSub(tenantName string, subs *concurrentMap[memphisWSReqTenantsToFiller], key string) {
	subs.Lock()
	defer subs.Unlock()
	if f, ok := subs.m[key]; ok {
		delete(f.tenants, tenantName)
		subs.m[key] = f
	}
}

func addTenantToSub(tenantName string, subs *concurrentMap[memphisWSReqTenantsToFiller], key string, filler memphisWSReqFiller) error {
	subs.Lock()
	defer subs.Unlock()
	if f, ok := subs.m[key]; ok {
		if _, ok := f.tenants[tenantName]; !ok {
			f.tenants[tenantName] = filler
			subs.m[key] = f
		}
	} else {
		return errors.New("addTenantToSub: sub not found")
	}
	return nil
}

func memphisWSLoop(s *Server, subs *concurrentMap[memphisWSReqTenantsToFiller], quitCh chan struct{}) {
	ticker := time.NewTicker(ws_updates_interval_sec * time.Second)
	for {
		select {
		case <-ticker.C:
			keys, values := subs.Array()
			for i, updateFiller := range values {
				k := keys[i]
				replySubj := fmt.Sprintf(memphisWS_TemplSubj_Publish, k+"."+s.opts.ServerName)
				for tenant, filler := range updateFiller.tenants {
					acc, err := s.lookupAccount(tenant)
					if err != nil {
						s.Warnf("[tenant: %v]memphisWSLoop at lookupAccount: %v ", tenant, err.Error())
						deleteTenantFromSub(tenant, subs, k)
						continue
					}
					if !acc.SubscriptionInterest(replySubj) {
						s.Debugf("removing tenant %v ws subscription %s", tenant, replySubj)
						deleteTenantFromSub(tenant, subs, k)
						continue
					}
					update, err := filler(tenant)
					if err != nil {
						if !IsNatsErr(err, JSStreamNotFoundErr) && !strings.Contains(err.Error(), "not exist") && !strings.Contains(err.Error(), "alphanumeric") {
							s.Errorf("[tenant: %v]memphisWSLoop at filler: %v", tenant, err.Error())
						}
						deleteTenantFromSub(tenant, subs, k)
						continue
					}
					updateRaw, err := json.Marshal(update)
					if err != nil {
						s.Errorf("[tenant: %v]memphisWSLoop at json.Marshal: %v", tenant, err.Error())
						deleteTenantFromSub(tenant, subs, k)
						continue
					}

					s.sendInternalAccountMsgWithEcho(acc, replySubj, updateRaw)
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
		tenantName, message, err := s.getTenantNameAndMessage(msg)
		if err != nil {
			s.Errorf("memphis websocket at getTenantNameAndMessage: %v", err.Error())
			return
		}
		s.Debugf("[tenant: %v]memphisWS registration - %s,%s", tenantName, subj, message)
		subscriptions := s.memphis.ws.subscriptions
		filteredSubj := tokensFromToEnd(subj, 2)
		trimmedMsg := strings.TrimSuffix(message, "\r\n")
		switch trimmedMsg {
		case memphisWS_SubscribeMsg:
			reqFiller, err := memphisWSGetReqFillerFromSubj(s, h, filteredSubj, tenantName)
			if err != nil {
				s.Errorf("[tenant: %v]memphis websocket at memphisWSGetReqFillerFromSubj: %v", tenantName, err.Error())
				return
			}
			if _, ok := subscriptions.Load(filteredSubj); !ok {
				subscriptions.Add(filteredSubj, memphisWSReqTenantsToFiller{tenants: map[string]memphisWSReqFiller{tenantName: reqFiller}})
			} else {
				err := addTenantToSub(tenantName, subscriptions, filteredSubj, reqFiller)
				if err != nil {
					s.Errorf("[tenant: %v]memphis websocket: %v", tenantName, err.Error())
				}
			}

		default:
			s.Errorf("[tenant: %v]memphis websocket: invalid sub/unsub operation", tenantName)
		}

		type brokerName struct {
			Name string `json:"name"`
		}

		broName := brokerName{s.opts.ServerName}
		serverName, err := json.Marshal(broName)
		if err != nil {
			s.Errorf("[tenant: %v]memphis websocket at json.Marshal: %v", tenantName, err.Error())
			return
		}
		s.sendInternalAccountMsgWithEcho(s.MemphisGlobalAccount(), reply, serverName)
	}
}

func memphisWSGetReqFillerFromSubj(s *Server, h *Handlers, subj string, tenantName string) (memphisWSReqFiller, error) {
	subjectHead := tokenAt(subj, 1)
	switch subjectHead {
	case memphisWS_Subj_MainOverviewData:
		return func(string) (any, error) {
			return h.Monitoring.getMainOverviewDataDetails(tenantName)
		}, nil

	case memphisWS_Subj_StationOverviewData:
		splitedResp := strings.Split(subj, ".")
		partitionNumberStr := splitedResp[len(splitedResp)-1]

		partitionNumber, err := strconv.Atoi(partitionNumberStr)
		if err != nil {
			return nil, fmt.Errorf("invalid partition number - %v", partitionNumberStr)
		}

		stationName := strings.Join(splitedResp[1:len(splitedResp)-1], ".")
		if stationName == _EMPTY_ {
			return nil, errors.New("invalid station name")
		}
		return func(string) (any, error) {
			return memphisWSGetStationOverviewData(s, h, stationName, tenantName, partitionNumber)
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
		return func(string) (any, error) {
			return h.Stations.GetStationsDetails(tenantName)
		}, nil

	case memphisWS_Subj_SysLogsData:
		logLevel := tokenAt(subj, 2)
		logSource := tokenAt(subj, 3)
		return func(string) (any, error) {
			return memphisWSGetSystemLogs(h, logLevel, logSource)
		}, nil
	case memphisWS_Subj_AllSchemasData:
		return func(string) (any, error) {
			return h.Schemas.GetAllSchemasDetails(tenantName)
		}, nil
	case memphisWS_Subj_GetSystemMessages:
		return func(string) (any, error) {
			return h.userMgmt.GetRelevantSystemMessages()
		}, nil
	case memphisWS_subj_GetAsyncTasks:
		return func(string) (any, error) {
			return h.AsyncTasks.GetAllAsyncTasks(tenantName)
		}, nil
	case memphisWS_subj_GetAllFunctions:
		return func(string) (any, error) {
			return h.Functions.GetFunctions(tenantName)
		}, nil
	case memphisWS_subj_GetGraphOverview:
		return func(string) (any, error) {
			return h.Monitoring.getGraphOverview(tenantName)
		}, nil
	case memphisWS_subj_GetFunctionsOverview:
		return func(string) (any, error) {
			stationName := tokenAt(subj, 2)
			partition := tokenAt(subj, 3)
			partitionInt, err := strconv.Atoi(partition)
			if err != nil {
				return "", err
			}
			return h.Monitoring.GetFunctionsOverview(stationName, tenantName, partitionInt)
		}, nil
	default:
		return nil, errors.New("invalid subject")
	}
}

func memphisWSGetStationOverviewData(s *Server, h *Handlers, stationName string, tenantName string, partitionNumber int) (map[string]any, error) {
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

	var functionsEnabled bool
	if station.Version >= 2 {
		functionsEnabled = true
	} else {
		functionsEnabled = false
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

	var totalMessages int
	var avgMsgSize int64
	messages := make([]models.MessageDetails, 0)
	messagesToFetch := 1000
	if partitionNumber == -1 {
		totalMessages, err = h.Stations.GetTotalMessages(station.TenantName, station.Name, station.PartitionsList)
		if err != nil {
			return map[string]any{}, err
		}
		avgMsgSize, err = h.Stations.GetAvgMsgSize(station)
		if err != nil {
			return map[string]any{}, err
		}
		messages, err = h.Stations.GetMessages(station, messagesToFetch)
		if err != nil {
			return map[string]any{}, err
		}
	} else {
		totalMessages, err = h.Stations.GetTotalPartitionMessages(station.TenantName, station.Name, partitionNumber)
		if err != nil {
			return map[string]any{}, err
		}
		avgMsgSize, err = h.Stations.GetPartitionAvgMsgSize(station.TenantName, fmt.Sprintf("%v$%v", sn.Intern(), partitionNumber))
		if err != nil {
			return map[string]any{}, err
		}
		messages, err = h.Stations.GetMessagesFromPartition(station, fmt.Sprintf("%v$%v", sn.Intern(), partitionNumber), messagesToFetch, partitionNumber)
		if err != nil {
			return map[string]any{}, err
		}
	}

	poisonMessages, schemaFailMessages, functionsMessages, totalDlsAmount, err := h.PoisonMsgs.GetDlsMsgsByStationLight(station, partitionNumber)
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
	leader, followers, err := h.Stations.GetLeaderAndFollowers(station, partitionNumber)
	if err != nil {
		return map[string]any{}, err
	}

	schema, err := h.Schemas.GetSchemaByStationName(sn, station.TenantName)

	if err != nil && err != ErrNoSchema {
		return map[string]any{}, err
	}

	var response map[string]any

	usageLimit, err := getUsageLimitProduersLimitPerStation(station.TenantName, station.Name)
	if err != nil {
		return map[string]any{}, err
	}

	if err == ErrNoSchema { // non native stations will always reach this point
		if !station.IsNative {
			cp, dp, cc, dc := getFakeProdsAndConsForPreview()
			response = map[string]any{
				"connected_producers":             cp,
				"disconnected_producers":          dp,
				"deleted_producers":               deletedProducers,
				"connected_cgs":                   cc,
				"disconnected_cgs":                disconnectedCgs,
				"deleted_cgs":                     dc,
				"total_messages":                  totalMessages,
				"average_message_size":            avgMsgSize,
				"audit_logs":                      auditLogs,
				"messages":                        messages,
				"poison_messages":                 poisonMessages,
				"schema_failed_messages":          schemaFailMessages,
				"functions_failed_messages":       functionsMessages,
				"tags":                            tags,
				"leader":                          leader,
				"followers":                       followers,
				"schema":                          struct{}{},
				"idempotency_window_in_ms":        station.IdempotencyWindow,
				"dls_configuration_poison":        station.DlsConfigurationPoison,
				"dls_configuration_schemaverse":   station.DlsConfigurationSchemaverse,
				"total_dls_messages":              totalDlsAmount,
				"tiered_storage_enabled":          station.TieredStorageEnabled,
				"created_by_username":             station.CreatedByUsername,
				"resend_disabled":                 station.ResendDisabled,
				"functions_enabled":               functionsEnabled,
				"max_amount_of_allowed_producers": usageLimit,
			}
		} else {
			response = map[string]any{
				"connected_producers":             connectedProducers,
				"disconnected_producers":          disconnectedProducers,
				"deleted_producers":               deletedProducers,
				"connected_cgs":                   connectedCgs,
				"disconnected_cgs":                disconnectedCgs,
				"deleted_cgs":                     deletedCgs,
				"total_messages":                  totalMessages,
				"average_message_size":            avgMsgSize,
				"audit_logs":                      auditLogs,
				"messages":                        messages,
				"poison_messages":                 poisonMessages,
				"schema_failed_messages":          schemaFailMessages,
				"functions_failed_messages":       functionsMessages,
				"tags":                            tags,
				"leader":                          leader,
				"followers":                       followers,
				"schema":                          struct{}{},
				"idempotency_window_in_ms":        station.IdempotencyWindow,
				"dls_configuration_poison":        station.DlsConfigurationPoison,
				"dls_configuration_schemaverse":   station.DlsConfigurationSchemaverse,
				"total_dls_messages":              totalDlsAmount,
				"tiered_storage_enabled":          station.TieredStorageEnabled,
				"created_by_username":             station.CreatedByUsername,
				"resend_disabled":                 station.ResendDisabled,
				"functions_enabled":               functionsEnabled,
				"max_amount_of_allowed_producers": usageLimit,
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
		"connected_producers":             connectedProducers,
		"disconnected_producers":          disconnectedProducers,
		"deleted_producers":               deletedProducers,
		"connected_cgs":                   connectedCgs,
		"disconnected_cgs":                disconnectedCgs,
		"deleted_cgs":                     deletedCgs,
		"total_messages":                  totalMessages,
		"average_message_size":            avgMsgSize,
		"audit_logs":                      auditLogs,
		"messages":                        messages,
		"poison_messages":                 poisonMessages,
		"schema_failed_messages":          schemaFailMessages,
		"functions_failed_messages":       functionsMessages,
		"tags":                            tags,
		"leader":                          leader,
		"followers":                       followers,
		"schema":                          schemaDetails,
		"idempotency_window_in_ms":        station.IdempotencyWindow,
		"dls_configuration_poison":        station.DlsConfigurationPoison,
		"dls_configuration_schemaverse":   station.DlsConfigurationSchemaverse,
		"total_dls_messages":              totalDlsAmount,
		"tiered_storage_enabled":          station.TieredStorageEnabled,
		"created_by_username":             station.CreatedByUsername,
		"resend_disabled":                 station.ResendDisabled,
		"functions_enabled":               functionsEnabled,
		"max_amount_of_allowed_producers": usageLimit,
	}

	return response, nil
}

func (s *Server) sendSystemMessageOnWS(user models.User, systemMessage SystemMessage) error {
	v, err := serv.Varz(nil)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]sendSystemMessageOnWS: %v", user.TenantName, user.Username, err.Error())
		return err
	}
	var serverNames []string
	if len(v.Cluster.URLs) == 0 {
		serverNames = append(serverNames, "memphis-0")
	}
	for i := range v.Cluster.URLs {
		serverNames = append(serverNames, "memphis-"+strconv.Itoa(i))
	}

	acc, err := serv.lookupAccount(user.TenantName)
	if err != nil {
		err = fmt.Errorf("sendSystemMessageOnWS at lookupAccount: %v", err.Error())
		return err
	}

	if systemMessage.Id == "" {
		uid, err := uuid.NewV4()
		if err != nil {
			err = fmt.Errorf("sendSystemMessageOnWS at uuid.NewV4: %v", err.Error())
			return err
		}
		systemMessage.Id = uid.String()
	}
	systemMessages := []SystemMessage{}
	systemMessages = append(systemMessages, systemMessage)
	updateRaw, err := json.Marshal(systemMessages)
	if err != nil {
		err = fmt.Errorf("sendSystemMessageOnWS at json.Marshal: %v", err.Error())
		return err
	}

	for _, serverName := range serverNames {
		replySubj := fmt.Sprintf(memphisWS_TemplSubj_Publish, memphisWS_Subj_GetSystemMessages+"."+serverName)
		serv.sendInternalAccountMsgWithEcho(acc, replySubj, updateRaw)
	}

	return nil
}
