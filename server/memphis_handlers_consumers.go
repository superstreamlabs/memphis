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
	"sort"
	"strconv"

	"strings"
	"time"

	"github.com/memphisdev/memphis/analytics"
	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/memphis_cache"
	"github.com/memphisdev/memphis/models"

	"k8s.io/utils/strings/slices"
)

type ConsumersHandler struct{ S *Server }

const (
	consumerObjectName = "Consumer"
)

func validateConsumerName(consumerName string) error {
	return validateName(consumerName, consumerObjectName)
}

func validateConsumerType(consumerType string) error {
	if consumerType != "application" && consumerType != "connector" {
		return fmt.Errorf("consumer type has to be one of the following application/connector and not %v", consumerType)
	}
	return nil
}

func isConsumerGroupExist(consumerGroup string, stationId int) (bool, models.Consumer, error) {
	exist, consumer, err := db.GetActiveConsumerByCG(consumerGroup, stationId)
	if err != nil {
		return false, models.Consumer{}, err
	} else if !exist {
		return false, models.Consumer{}, nil
	}
	return true, consumer, nil
}

func GetConsumerGroupMembers(cgName string, station models.Station) ([]models.CgMember, error) {
	consumers, err := db.GetConsumerGroupMembers(cgName, station.ID)
	if err != nil {
		return consumers, err
	}

	var dedupedConsumers []models.CgMember
	consumersNames := []string{}

	for _, consumer := range consumers {
		if slices.Contains(consumersNames, consumer.Name) {
			continue
		}
		consumersNames = append(consumersNames, consumer.Name)
		dedupedConsumers = append(dedupedConsumers, consumer)
	}

	return dedupedConsumers, nil
}

func (s *Server) createConsumerDirectV0(c *client, reply, tenantName string, ccr createConsumerRequestV0, requestVersion int) {
	_, err := s.createConsumerDirectCommon(c, ccr.Name, ccr.StationName, ccr.ConsumerGroup, ccr.ConsumerType, ccr.ConnectionId, tenantName, ccr.Username, ccr.MaxAckTimeMillis, ccr.MaxMsgDeliveries, requestVersion, 1, -1, ccr.ConnectionId)
	respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
}

func (s *Server) createConsumerDirectCommon(c *client, consumerName, cStationName, cGroup, cType, connectionId, tenantName, userName string, maxAckTime, maxMsgDeliveries, requestVersion int, startConsumeFromSequence uint64, lastMessages int64, appId string) ([]int, error) {
	name := strings.ToLower(consumerName)
	err := validateConsumerName(name)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]createConsumerDirectCommon at validateConsumerName: Failed creating consumer %v at station %v : %v", tenantName, userName, consumerName, cStationName, err.Error())
		return []int{}, err
	}

	consumerGroup := strings.ToLower(cGroup)
	if consumerGroup != "" {
		err = validateConsumerName(consumerGroup)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]createConsumerDirectCommon at validateConsumerName: Failed creating consumer %v at station %v : %v", tenantName, userName, consumerName, cStationName, err.Error())
			return []int{}, err
		}
	} else {
		consumerGroup = name
	}

	consumerType := strings.ToLower(cType)
	err = validateConsumerType(consumerType)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]createConsumerDirectCommon at validateConsumerType: Failed creating consumer %v at station %v : %v", tenantName, userName, consumerName, cStationName, err.Error())
		return []int{}, err
	}

	stationName, err := StationNameFromStr(cStationName)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]createConsumerDirectCommon at StationNameFromStr: Consumer %v at station %v : %v", tenantName, userName, consumerName, cStationName, err.Error())
		return []int{}, err
	}

	exist, user, err := memphis_cache.GetUser(userName, tenantName, false)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]createConsumerDirectCommon at GetUser from cache: Consumer %v at station %v : %v", tenantName, userName, consumerName, cStationName, err.Error())
		return []int{}, err
	} else if !exist {
		err := errors.New("user does not exist")
		serv.Warnf("[tenant: %v][user: %v] createConsumerDirectCommon at GetUser from cache: %s", tenantName, userName, err.Error())
		return []int{}, err
	}

	exist, station, err := db.GetStationByName(stationName.Ext(), user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v]createConsumerDirectCommon at GetStationByName: Consumer %v at station %v : %v", tenantName, consumerName, cStationName, err.Error())
		return []int{}, err
	}

	if !exist {
		if requestVersion < 2 {
			err := errors.New("This station does not exist, a default station can not be created automatically, please upgrade your SDK version")
			serv.Warnf("[tenant: %v]createConsumerDirectCommon at CreateDefaultStation: Consumer %v at station %v : %v", tenantName, consumerName, cStationName, err.Error())
			return []int{}, err
		}
		var created bool
		station, created, err = CreateDefaultStation(user.TenantName, s, stationName, user.ID, user.Username, "", 0)
		if err != nil {
			serv.Warnf("[tenant: %v]createConsumerDirectCommon at CreateDefaultStation: Consumer %v at station %v : %v", tenantName, consumerName, cStationName, err.Error())
			return []int{}, err
		}

		if created {
			message := fmt.Sprintf("Station %v has been created by user %v", stationName.Ext(), user.Username)
			serv.Noticef("[tenant: %v][user: %v]: %v", user.TenantName, user.Username, message)
			var auditLogs []interface{}
			newAuditLog := models.AuditLog{
				StationName:       stationName.Ext(),
				Message:           message,
				CreatedBy:         user.ID,
				CreatedByUsername: user.Username,
				CreatedAt:         time.Now(),
				TenantName:        user.TenantName,
			}
			auditLogs = append(auditLogs, newAuditLog)
			err = CreateAuditLogs(auditLogs)
			if err != nil {
				serv.Errorf("[tenant: %v]createConsumerDirect at CreateAuditLogs: Consumer %v at station %v :%v", user.TenantName, consumerName, cStationName, err.Error())
			}

			shouldSendAnalytics, _ := shouldSendAnalytics()
			if shouldSendAnalytics {
				analyticsParams := map[string]interface{}{"station-name": stationName.Ext(), "storage-type": "disk"}
				analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-create-station-sdk")
			}
		}
	} else {
		if requestVersion < 2 && len(station.PartitionsList) > 0 {
			err := errors.New("to consume from this station please upgrade your SDK version")
			serv.Warnf("[tenant: %v]createConsumerDirectCommon at CreateDefaultStation: Consumer %v at station %v : %v", tenantName, consumerName, cStationName, err.Error())
			return []int{}, err
		}
	}

	consumerGroupExist, consumerFromGroup, err := isConsumerGroupExist(consumerGroup, station.ID)
	if err != nil {
		serv.Errorf("[tenant: %v]createConsumerDirectCommon at isConsumerGroupExist: Consumer %v at station %v :%v", user.TenantName, consumerName, cStationName, err.Error())
		return []int{}, err
	}

	splitted := strings.Split(c.opts.Lang, ".")
	sdkName := splitted[len(splitted)-1]
	var newConsumer models.Consumer
	if strings.HasPrefix(user.Username, "$") {
		newConsumer, err = db.InsertNewConsumer(name, station.ID, "connector", connectionId, consumerGroup, maxAckTime, maxMsgDeliveries, startConsumeFromSequence, lastMessages, tenantName, station.PartitionsList, requestVersion, sdkName, appId)
		if err != nil {
			serv.Errorf("[tenant: %v]createConsumerDirectCommon at InsertNewConsumer: Consumer %v at station %v :%v", user.TenantName, consumerName, cStationName, err.Error())
			return []int{}, err
		}
	} else {
		newConsumer, err = db.InsertNewConsumer(name, station.ID, consumerType, connectionId, consumerGroup, maxAckTime, maxMsgDeliveries, startConsumeFromSequence, lastMessages, tenantName, station.PartitionsList, requestVersion, sdkName, appId)
		if err != nil {
			serv.Errorf("[tenant: %v]createConsumerDirectCommon at InsertNewConsumer: Consumer %v at station %v :%v", user.TenantName, consumerName, cStationName, err.Error())
			return []int{}, err
		}
	}

	message := "Consumer " + name + " connected"
	if consumerGroupExist {
		if requestVersion == 1 {
			if newConsumer.StartConsumeFromSeq != consumerFromGroup.StartConsumeFromSeq || newConsumer.LastMessages != consumerFromGroup.LastMessages {
				err := errors.New("consumer already exists with different uneditable configuration parameters (StartConsumeFromSequence/LastMessages)")
				serv.Warnf("createConsumerDirectCommon: %v", err.Error())
				return []int{}, err
			}
			if !comparePartitionsList(consumerFromGroup.PartitionsList, newConsumer.PartitionsList) {
				existingPartitions := ""
				for i, pl := range consumerFromGroup.PartitionsList {
					existingPartitions += strconv.Itoa(pl)
					if i < len(consumerFromGroup.PartitionsList)-1 {
						existingPartitions += ", "
					}
				}
				err := errors.New("consumer already exists with different uneditable partition list: partition numbers: ")
				serv.Warnf("createConsumerDirectCommon: %v", err.Error())
				return []int{}, err
			}
		}

		if newConsumer.MaxAckTimeMs != consumerFromGroup.MaxAckTimeMs || newConsumer.MaxMsgDeliveries != consumerFromGroup.MaxMsgDeliveries {
			err := s.CreateConsumer(station.TenantName, newConsumer, station, station.PartitionsList)
			if err != nil {
				if IsNatsErr(err, JSStreamNotFoundErr) {
					serv.Warnf("[tenant: %v][user: %v]createConsumerDirectCommon: Consumer %v at station %v: station does not exist", user.TenantName, user.Username, consumerName, cStationName)
				} else {
					serv.Errorf("[tenant: %v][user: %v]createConsumerDirectCommon at CreateConsumer: Consumer %v at station %v: %v", user.TenantName, user.Username, consumerName, cStationName, err.Error())
				}
				return []int{}, err
			}
		}
	} else {
		err := s.CreateConsumer(station.TenantName, newConsumer, station, station.PartitionsList)
		if err != nil {
			if IsNatsErr(err, JSStreamNotFoundErr) {
				serv.Warnf("[tenant: %v][user: %v]createConsumerDirectCommon: Consumer %v at station %v: station does not exist", user.TenantName, user.Username, consumerName, cStationName)
			} else {
				serv.Errorf("[tenant: %v][user: %v]createConsumerDirectCommon at CreateConsumer: Consumer %v at station %v: %v", user.TenantName, user.Username, consumerName, cStationName, err.Error())
			}
			return []int{}, err
		}
	}
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		StationName:       stationName.Ext(),
		Message:           message,
		CreatedBy:         user.ID,
		CreatedByUsername: user.Username,
		CreatedAt:         time.Now(),
		TenantName:        user.TenantName,
	}
	auditLogs = append(auditLogs, newAuditLog)
	err = CreateAuditLogs(auditLogs)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]createConsumerDirectCommon at CreateAuditLogs: Consumer %v at station %v: %v", user.TenantName, user.Username, consumerName, cStationName, err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		ip := serv.getIp()
		analyticsParams := map[string]interface{}{"consumer-name": newConsumer.Name, "ip": ip}
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-create-consumer-sdk")
	}
	return station.PartitionsList, nil
}

func (s *Server) createConsumerDirect(c *client, reply string, msg []byte) {
	var ccr createConsumerRequestV2
	var resp createConsumerResponse

	tenantName, message, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("createConsumerDirect at getTenantNameAndMessage: %v", err.Error())
		return
	}

	if err := json.Unmarshal([]byte(message), &ccr); err != nil || ccr.RequestVersion < 3 {
		var ccrV1 createConsumerRequestV1
		if err := json.Unmarshal([]byte(message), &ccrV1); err != nil {
			var ccrV0 createConsumerRequestV0
			if err := json.Unmarshal([]byte(message), &ccrV0); err != nil {
				s.Errorf("[tenant: %v]createConsumerDirect at json.Unmarshal: Failed creating consumer: %v: %v", tenantName, err.Error(), string(msg))
				respondWithRespErr(serv.MemphisGlobalAccountString(), s, reply, err, &resp)
				return
			}
			s.createConsumerDirectV0(c, reply, tenantName, ccrV0, ccr.RequestVersion)
			return
		}
		ccr = createConsumerRequestV2{
			Name:                     ccrV1.Name,
			StationName:              ccrV1.StationName,
			ConnectionId:             ccrV1.ConnectionId,
			ConsumerType:             ccrV1.ConsumerType,
			ConsumerGroup:            ccrV1.ConsumerGroup,
			MaxAckTimeMillis:         ccrV1.MaxAckTimeMillis,
			MaxMsgDeliveries:         ccrV1.MaxMsgDeliveries,
			Username:                 ccrV1.Username,
			StartConsumeFromSequence: ccrV1.StartConsumeFromSequence,
			LastMessages:             ccrV1.LastMessages,
			RequestVersion:           ccrV1.RequestVersion,
			AppId:                    ccrV1.ConnectionId,
		}
	}

	ccr.TenantName = tenantName
	if ccr.StartConsumeFromSequence <= 0 {
		err := errors.New("startConsumeFromSequence has to be a positive number")
		serv.Warnf("[tenant: %v]createConsumerDirect: %v", tenantName, err.Error())
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	if ccr.LastMessages < -1 {
		err := errors.New("min value for LastMessages is -1")
		serv.Warnf("[tenant: %v]createConsumerDirect: %v", tenantName, err.Error())
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	if ccr.StartConsumeFromSequence > 1 && ccr.LastMessages > -1 {
		err := errors.New("consumer creation options can't contain both startConsumeFromSequence and lastMessages")
		serv.Warnf("[tenant: %v]createConsumerDirect: %v", tenantName, err.Error())
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	partitions, err := s.createConsumerDirectCommon(c, ccr.Name, ccr.StationName, ccr.ConsumerGroup, ccr.ConsumerType, ccr.ConnectionId, tenantName, ccr.Username, ccr.MaxAckTimeMillis, ccr.MaxMsgDeliveries, ccr.RequestVersion, ccr.StartConsumeFromSequence, ccr.LastMessages, ccr.AppId)
	if err != nil {
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
	}

	sn, err := StationNameFromStr(ccr.StationName)
	if err != nil {
		s.Warnf("[tenant: %v][user: %v]createConsumerDirect at StationNameFromStr: Consumer %v at station %v: %v", ccr.TenantName, ccr.Username, ccr.Name, ccr.StationName, err.Error())
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
		return
	}

	schemaUpdate, err := getSchemaUpdateInitFromStation(sn, ccr.TenantName)
	if err == ErrNoSchema {
		v1Resp := createConsumerResponseV1{PartitionsUpdate: models.PartitionsUpdate{PartitionsList: partitions}, Err: ""}
		respondWithResp(s.MemphisGlobalAccountString(), s, reply, &v1Resp)
		return
	}
	if err != nil {
		if strings.Contains(err.Error(), "not exist") {
			s.Warnf("[tenant: %v][user: %v]createConsumerDirect at getSchemaUpdateInitFromStation: Consumer %v at station %v: %v", ccr.TenantName, ccr.Username, ccr.Name, ccr.StationName, err.Error())
		} else {
			s.Errorf("[tenant: %v][user: %v]createConsumerDirect at getSchemaUpdateInitFromStation: Consumer %v at station %v: %v", ccr.TenantName, ccr.Username, ccr.Name, ccr.StationName, err.Error())
		}
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
		return
	}
	if len(partitions) == 0 && ccr.RequestVersion < 2 {
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
	} else {
		v1Resp := createConsumerResponseV1{SchemaUpdate: *schemaUpdate, PartitionsUpdate: models.PartitionsUpdate{PartitionsList: partitions}, Err: ""}
		respondWithResp(s.MemphisGlobalAccountString(), s, reply, &v1Resp)
	}
}

func (ch ConsumersHandler) GetCgsByStation(stationName StationName, station models.Station) ([]models.Cg, []models.Cg, []models.Cg, error) { // for socket io endpoint
	var cgs []models.Cg
	consumers, err := db.GetAllConsumersByStation(station.ID)
	if err != nil {
		return cgs, cgs, cgs, err
	}

	if len(consumers) == 0 {
		return []models.Cg{}, []models.Cg{}, []models.Cg{}, nil
	}

	m := make(map[string]*models.Cg)
	consumersGroupNames := []string{}

	for _, consumer := range consumers {
		if slices.Contains(consumersGroupNames, consumer.ConsumersGroup+consumer.Name) {
			continue
		}
		consumersGroupNames = append(consumersGroupNames, consumer.ConsumersGroup+consumer.Name)

		var cg *models.Cg
		if m[consumer.ConsumersGroup] == nil {
			cg = &models.Cg{
				Name:                  consumer.ConsumersGroup,
				MaxAckTimeMs:          consumer.MaxAckTimeMs,
				MaxMsgDeliveries:      consumer.MaxMsgDeliveries,
				ConnectedConsumers:    []models.ExtendedConsumerResponse{},
				DisconnectedConsumers: []models.ExtendedConsumerResponse{},
				DeletedConsumers:      []models.ExtendedConsumerResponse{},
				IsActive:              consumer.IsActive,
				LastStatusChangeDate:  consumer.UpdatedAt,
				PartitionsList:        consumer.PartitionsList,
				SdkLanguage:           consumers[0].Sdk,
			}
			m[consumer.ConsumersGroup] = cg
		} else {
			m[consumer.ConsumersGroup].Name = consumer.ConsumersGroup
			m[consumer.ConsumersGroup].MaxAckTimeMs = consumer.MaxAckTimeMs
			m[consumer.ConsumersGroup].MaxMsgDeliveries = consumer.MaxMsgDeliveries
			m[consumer.ConsumersGroup].IsActive = consumer.IsActive
			m[consumer.ConsumersGroup].LastStatusChangeDate = consumer.UpdatedAt
			cg = m[consumer.ConsumersGroup]
			cg.SdkLanguage = consumers[0].Sdk
		}

		needToUpdateVersion := false
		if consumer.Version < lastConsumerCreationReqVersion && consumer.IsActive {
			needToUpdateVersion = true
			cg.UpdateAvailable = true
		}

		consumerRes := models.ExtendedConsumerResponse{
			ID:               consumer.ID,
			Name:             consumer.Name,
			IsActive:         consumer.IsActive,
			ConsumersGroup:   consumer.ConsumersGroup,
			MaxAckTimeMs:     consumer.MaxAckTimeMs,
			MaxMsgDeliveries: consumer.MaxMsgDeliveries,
			StationName:      consumer.StationName,
			Count:            consumer.Count,
			PartitionsList:   consumer.PartitionsList,
			SdkLanguage:      consumer.Sdk,
			UpdateAvailable:  needToUpdateVersion,
		}

		if consumer.IsActive {
			cg.ConnectedConsumers = append(cg.ConnectedConsumers, consumerRes)
		} else {
			cg.DisconnectedConsumers = append(cg.DisconnectedConsumers, consumerRes)
		}
	}

	var connectedCgs []models.Cg
	var disconnectedCgs []models.Cg
	var deletedCgs []models.Cg

	for _, cg := range m {
		cgInfo, err := ch.S.GetCgInfo(station.TenantName, stationName, cg.Name, cg.PartitionsList)
		if err != nil {
			continue // ignoring cases where the consumer exist in memphis but not in nats
		}

		totalPoisonMsgs, err := db.GetTotalPoisonMsgsPerCg(cg.Name, station.ID)
		if err != nil {
			return []models.Cg{}, []models.Cg{}, []models.Cg{}, err
		}

		cg.InProcessMessages = cgInfo.NumAckPending
		cg.UnprocessedMessages = int(cgInfo.NumPending)
		cg.PoisonMessages = totalPoisonMsgs

		if len(cg.ConnectedConsumers) > 0 {
			cg.IsActive = true
			connectedCgs = append(connectedCgs, *cg)
		} else if len(cg.DisconnectedConsumers) > 0 {
			disconnectedCgs = append(disconnectedCgs, *cg)
		} else {
			deletedCgs = append(deletedCgs, *cg)
		}
	}

	if len(connectedCgs) == 0 {
		connectedCgs = []models.Cg{}
	}

	if len(disconnectedCgs) == 0 {
		disconnectedCgs = []models.Cg{}
	}

	if len(deletedCgs) == 0 {
		deletedCgs = []models.Cg{}
	}

	sort.Slice(connectedCgs, func(i, j int) bool {
		return connectedCgs[j].LastStatusChangeDate.Before(connectedCgs[i].LastStatusChangeDate)
	})
	sort.Slice(disconnectedCgs, func(i, j int) bool {
		return disconnectedCgs[j].LastStatusChangeDate.Before(disconnectedCgs[i].LastStatusChangeDate)
	})
	sort.Slice(deletedCgs, func(i, j int) bool {
		return deletedCgs[j].LastStatusChangeDate.Before(deletedCgs[i].LastStatusChangeDate)
	})
	return connectedCgs, disconnectedCgs, deletedCgs, nil
}

func (ch ConsumersHandler) GetDelayedCgsByTenant(tenantName string, streams []*StreamInfo) ([]models.DelayedCgResp, error) {
	consumers := make(map[string]map[string]models.DelayedCg, 0)
	consumerNames := []string{}
	for _, stream := range streams {
		if strings.HasPrefix(stream.Config.Name, "$memphis") {
			continue
		}
		offset := 0
		requestSubject := fmt.Sprintf(JSApiConsumerListT, stream.Config.Name)
		offsetReq := ApiPagedRequest{Offset: offset}
		request := JSApiConsumersRequest{ApiPagedRequest: offsetReq}
		rawRequest, err := json.Marshal(request)
		if err != nil {
			return []models.DelayedCgResp{}, err
		}
		var resp JSApiConsumerListResponse
		err = jsApiRequest(tenantName, ch.S, requestSubject, kindConsumerInfo, []byte(rawRequest), &resp)
		if err != nil {
			return []models.DelayedCgResp{}, err
		}
		err = resp.ToError()
		if err != nil {
			return []models.DelayedCgResp{}, err
		}
		for _, consumer := range resp.Consumers {
			if strings.HasPrefix(consumer.Config.FilterSubject, "$memphis") || !strings.HasSuffix(consumer.Config.FilterSubject, ".final") { // skip consumers that are not user consumers
				continue
			}
			if strings.Contains(consumer.Stream, "$") {
				consumer.Stream = strings.Split(consumer.Stream, "$")[0]
			}
			if consumer.NumPending > 0 {
				stationName := StationNameFromStreamName(consumer.Stream)
				consumerName := revertDelimiters(consumer.Name)
				externalStationName := stationName.Ext()
				if _, ok := consumers[externalStationName]; !ok {
					consumers[externalStationName] = map[string]models.DelayedCg{consumerName: {CGName: consumerName, NumOfDelayedMsgs: uint64(consumer.NumPending)}}
				} else {
					if _, ok := consumers[externalStationName][consumerName]; !ok {
						consumers[externalStationName][consumerName] = models.DelayedCg{CGName: consumerName, NumOfDelayedMsgs: uint64(consumer.NumPending)}
					} else {
						numOfDelayMsgs := consumers[externalStationName][consumerName].NumOfDelayedMsgs + uint64(consumer.NumPending)
						consumers[externalStationName][consumerName] = models.DelayedCg{CGName: consumerName, NumOfDelayedMsgs: numOfDelayMsgs}
					}
				}
				consumerNames = append(consumerNames, consumerName)
			}
		}
	}
	consumersFromDb, err := db.GetActiveCgsByName(consumerNames, tenantName)
	if err != nil {
		return []models.DelayedCgResp{}, err
	}
	if len(consumersFromDb) == 0 {
		return []models.DelayedCgResp{}, nil
	}
	consumersResp := make(map[string][]models.DelayedCg, 0)
	for _, c := range consumersFromDb {
		if mdcg, ok := consumers[c.StationName]; ok {
			if dcg, ok := mdcg[c.Name]; ok {
				if _, ok := consumersResp[c.StationName]; !ok {
					consumersResp[c.StationName] = []models.DelayedCg{dcg}
				} else {
					consumersResp[c.StationName] = append(consumersResp[c.StationName], dcg)
				}
			}
		}
	}
	delayedCgsResp := []models.DelayedCgResp{}
	for s, c := range consumersResp {
		delayedCgsResp = append(delayedCgsResp, models.DelayedCgResp{StationName: s, CGS: c})
	}
	return delayedCgsResp, nil
}

func (s *Server) destroyConsumerDirect(c *client, reply string, msg []byte) {
	var dcr destroyConsumerRequestV1
	tenantName, message, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("destroyConsumerDirect at getTenantNameAndMessage: %v", err.Error())
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	if err := json.Unmarshal([]byte(message), &dcr); err != nil || dcr.RequestVersion < 1 {
		var dcrV0 destroyConsumerRequestV0
		if err := json.Unmarshal([]byte(message), &dcrV0); err != nil {
			s.Errorf("[tenant: %v]destroyConsumerDirect at json.Unmarshal: %v", tenantName, err.Error())
			respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
			return
		}
		dcrV0.TenantName = tenantName
		if c.memphisInfo.connectionId == "" {
			s.destroyConsumerDirectV0(c, reply, dcrV0)
			return
		} else {
			dcr = destroyConsumerRequestV1{
				StationName:    dcrV0.StationName,
				ConsumerName:   dcrV0.ConsumerName,
				Username:       dcrV0.Username,
				ConnectionId:   c.memphisInfo.connectionId,
				RequestVersion: 1,
			}
		}

		dcr.TenantName = tenantName
	}

	dcr.TenantName = tenantName
	stationName, err := StationNameFromStr(dcr.StationName)
	if err != nil {
		serv.Errorf("[tenant: %v]DestroyConsumer at StationNameFromStr: Station %v: %v", tenantName, dcr.StationName, err.Error())
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	name := strings.ToLower(dcr.ConsumerName)
	_, station, err := db.GetStationByName(stationName.Ext(), dcr.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v]DestroyConsumer at GetStationByName: Station %v: %v", tenantName, dcr.StationName, err.Error())
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	exist, consumer, err := db.DeleteConsumerByNameStationIDAndConnID(dcr.ConnectionId, name, station.ID)
	if !exist {
		errMsg := fmt.Sprintf("Consumer %v at station %v does not exist", dcr.ConsumerName, dcr.StationName)
		serv.Warnf("[tenant: %v]DestroyConsumer: %v", tenantName, errMsg)
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, errors.New(errMsg))
		return
	}
	if err != nil {
		errMsg := fmt.Sprintf("Consumer %v at station %v: %v", dcr.ConsumerName, dcr.StationName, err.Error())
		serv.Errorf("[tenant: %v]DestroyConsumer: %v", tenantName, errMsg)
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	s.destroyCGFromNats(c, reply, dcr.Username, tenantName, stationName, consumer, station)
}

func (s *Server) destroyConsumerDirectV0(c *client, reply string, dcr destroyConsumerRequestV0) {
	stationName, err := StationNameFromStr(dcr.StationName)
	if err != nil {
		serv.Errorf("[tenant: %v]destroyConsumerDirectV0 at StationNameFromStr: Station %v: %v", dcr.TenantName, dcr.StationName, err.Error())
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	name := strings.ToLower(dcr.ConsumerName)
	_, station, err := db.GetStationByName(stationName.Ext(), dcr.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v]destroyConsumerDirectV0 at GetStationByName: Station %v: %v", dcr.TenantName, dcr.StationName, err.Error())
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	exist, consumer, err := db.DeleteConsumerByNameAndStationId(name, station.ID)
	if !exist {
		errMsg := fmt.Sprintf("Consumer %v at station %v does not exist", dcr.ConsumerName, dcr.StationName)
		serv.Warnf("[tenant: %v]destroyConsumerDirectV0: %v", dcr.TenantName, errMsg)
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, errors.New(errMsg))
		return
	}
	if err != nil {
		errMsg := fmt.Sprintf("Consumer %v at station %v: %v", dcr.ConsumerName, dcr.StationName, err.Error())
		serv.Errorf("[tenant: %v]destroyConsumerDirectV0: %v", dcr.TenantName, errMsg)
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	s.destroyCGFromNats(c, reply, dcr.Username, dcr.TenantName, stationName, consumer, station)
}

func (s *Server) destroyCGFromNats(c *client, reply, userName, tenantName string, stationName StationName, consumer models.Consumer, station models.Station) {

	// ensure not part of an active consumer group
	count, err := db.CountActiveConsumersInCG(consumer.ConsumersGroup, station.ID)
	if err != nil {
		errMsg := fmt.Sprintf("[tenant: %v]Consumer %v at station %v: %v", tenantName, consumer.Name, station.Name, err.Error())
		serv.Errorf("destroyCGFromNats at CountActiveConsumersInCG: %v", errMsg)
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	deleted := false
	if count == 0 { // no other members in this group
		err = s.RemoveConsumer(station.TenantName, stationName, consumer.ConsumersGroup, consumer.PartitionsList)
		if err != nil && !IsNatsErr(err, JSConsumerNotFoundErr) && !IsNatsErr(err, JSStreamNotFoundErr) {
			errMsg := fmt.Sprintf("[tenant: %v]Consumer group %v at station %v: %v", tenantName, consumer.ConsumersGroup, station.Name, err.Error())
			serv.Errorf("destroyCGFromNats at RemoveConsumer: %v", errMsg)
			respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
			return
		}
		if err == nil {
			deleted = true
		}

		err = db.RemovePoisonedCg(station.ID, consumer.ConsumersGroup)
		if err != nil && !IsNatsErr(err, JSConsumerNotFoundErr) && !IsNatsErr(err, JSStreamNotFoundErr) {
			errMsg := fmt.Sprintf("[tenant: %v]Consumer group %v at station %v: %v", tenantName, consumer.ConsumersGroup, station.Name, err.Error())
			serv.Errorf("DestroyConsumer at RemovePoisonedCg: %v", errMsg)
			respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
			return
		}
	}

	name := strings.ToLower(consumer.Name)
	if deleted {
		username := c.memphisInfo.username
		if username == "" {
			username = userName
		}
		_, user, err := memphis_cache.GetUser(username, consumer.TenantName, false)
		if err != nil && !IsNatsErr(err, JSConsumerNotFoundErr) && !IsNatsErr(err, JSStreamNotFoundErr) {
			errMsg := fmt.Sprintf("[tenant: %v]Consumer group %v at station %v: %v", tenantName, consumer.ConsumersGroup, station.Name, err.Error())
			serv.Errorf("destroyCGFromNats at GetUserByUsername: " + errMsg)
			respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
			return
		}
		message := fmt.Sprintf("Consumer %v has been destroyed", name)
		serv.Noticef("[tenant: %v][user: %v]: %v", user.TenantName, user.Username, message)
		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			StationName:       stationName.Ext(),
			Message:           message,
			CreatedBy:         user.ID,
			CreatedByUsername: user.Username,
			CreatedAt:         time.Now(),
			TenantName:        user.TenantName,
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			serv.Errorf("[tenant: %v]destroyCGFromNats at CreateAuditLogs: Consumer %v at station %v: %v", user.TenantName, consumer.Name, station.Name, err.Error())
		}

		shouldSendAnalytics, _ := shouldSendAnalytics()
		if shouldSendAnalytics {
			analyticsParams := make(map[string]interface{})
			analytics.SendEvent(user.TenantName, username, analyticsParams, "user-remove-consumer-sdk")
		}
	}

	respondWithErr(serv.MemphisGlobalAccountString(), s, reply, nil)

}

func comparePartitionsList(pList1, pList2 []int) bool {
	if len(pList1) != len(pList2) {
		return false
	}
	for i := 0; i < len(pList1); i++ {
		if pList1[i] != pList2[i] {
			return false
		}
	}
	return true
}
