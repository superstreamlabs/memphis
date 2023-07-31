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

	"memphis/analytics"
	"memphis/db"
	"memphis/memphis_cache"
	"memphis/models"
	"strings"
	"time"

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
	err := s.createConsumerDirectCommon(c, ccr.Name, ccr.StationName, ccr.ConsumerGroup, ccr.ConsumerType, ccr.ConnectionId, tenantName, ccr.Username, ccr.MaxAckTimeMillis, ccr.MaxMsgDeliveries, requestVersion, 1, -1)
	respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
}

func (s *Server) createConsumerDirectCommon(c *client, consumerName, cStationName, cGroup, cType, connectionId, tenantName, userName string, maxAckTime, maxMsgDeliveries, requestVersion int, startConsumeFromSequence uint64, lastMessages int64) error {
	name := strings.ToLower(consumerName)
	err := validateConsumerName(name)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]createConsumerDirectCommon at validateConsumerName: Failed creating consumer %v at station %v : %v", tenantName, userName, consumerName, cStationName, err.Error())
		return err
	}

	consumerGroup := strings.ToLower(cGroup)
	if consumerGroup != "" {
		err = validateConsumerName(consumerGroup)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]createConsumerDirectCommon at validateConsumerName: Failed creating consumer %v at station %v : %v", tenantName, userName, consumerName, cStationName, err.Error())
			return err
		}
	} else {
		consumerGroup = name
	}

	consumerType := strings.ToLower(cType)
	err = validateConsumerType(consumerType)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]createConsumerDirectCommon at validateConsumerType: Failed creating consumer %v at station %v : %v", tenantName, userName, consumerName, cStationName, err.Error())
		return err
	}

	stationName, err := StationNameFromStr(cStationName)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]createConsumerDirectCommon at StationNameFromStr: Consumer %v at station %v : %v", tenantName, userName, consumerName, cStationName, err.Error())
		return err
	}

	exist, user, err := memphis_cache.GetUser(userName, tenantName, false)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]createConsumerDirectCommon at GetUser from cache: Consumer %v at station %v : %v", tenantName, userName, consumerName, cStationName, err.Error())
		return err
	} else if !exist {
		serv.Warnf("[tenant: %v][user: %v] createConsumerDirectCommon at GetUser from cache: user does not exist", tenantName, userName)
		return fmt.Errorf("user does not exist in db")
	}

	exist, station, err := db.GetStationByName(stationName.Ext(), user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v]createConsumerDirectCommon at GetStationByName: Consumer %v at station %v : %v", tenantName, consumerName, cStationName, err.Error())
		return err
	}

	if !exist {
		var created bool
		station, created, err = CreateDefaultStation(user.TenantName, s, stationName, user.ID, user.Username)
		if err != nil {
			serv.Warnf("[tenant: %v]createConsumerDirectCommon at CreateDefaultStation: Consumer %v at station %v : %v", tenantName, consumerName, cStationName, err.Error())
			return err
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
	}

	consumerGroupExist, consumerFromGroup, err := isConsumerGroupExist(consumerGroup, station.ID)
	if err != nil {
		serv.Errorf("[tenant: %v]createConsumerDirectCommon at isConsumerGroupExist: Consumer %v at station %v :%v", user.TenantName, consumerName, cStationName, err.Error())
		return err
	}

	exist, newConsumer, rowsUpdated, err := db.InsertNewConsumer(name, station.ID, consumerType, connectionId, consumerGroup, maxAckTime, maxMsgDeliveries, startConsumeFromSequence, lastMessages, tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v]createConsumerDirectCommon at InsertNewConsumer: Consumer %v at station %v :%v", user.TenantName, consumerName, cStationName, err.Error())
		return err
	}
	if exist {
		errMsg := fmt.Sprintf("Consumer %v at station %v: Consumer name has to be unique per station", consumerName, cStationName)
		serv.Errorf("[tenant: %v]createConsumerDirectCommon: %v", user.TenantName, errMsg)
		return fmt.Errorf("memphis: %v", errMsg)
	}

	if rowsUpdated == 1 {
		message := "Consumer " + name + " connected"
		serv.Noticef("[tenant: %v][user: %v]: %v", user.TenantName, user.Username, message)
		if consumerGroupExist {
			if requestVersion == 1 {
				if newConsumer.StartConsumeFromSeq != consumerFromGroup.StartConsumeFromSeq || newConsumer.LastMessages != consumerFromGroup.LastMessages {
					errMsg := errors.New("consumer already exists with different uneditable configuration parameters (StartConsumeFromSequence/LastMessages)")
					serv.Warnf("createConsumerDirectCommon: %v", errMsg.Error())
					return errMsg
				}
			}

			if newConsumer.MaxAckTimeMs != consumerFromGroup.MaxAckTimeMs || newConsumer.MaxMsgDeliveries != consumerFromGroup.MaxMsgDeliveries {
				err := s.CreateConsumer(station.TenantName, newConsumer, station)
				if err != nil {
					if IsNatsErr(err, JSStreamNotFoundErr) {
						serv.Warnf("[tenant: %v][user: %v]createConsumerDirectCommon: Consumer %v at station %v: station does not exist", user.TenantName, user.Username, consumerName, cStationName)
					} else {
						serv.Errorf("[tenant: %v][user: %v]createConsumerDirectCommon at CreateConsumer: Consumer %v at station %v: %v", user.TenantName, user.Username, consumerName, cStationName, err.Error())
					}
					return err
				}
			}
		} else {
			err := s.CreateConsumer(station.TenantName, newConsumer, station)
			if err != nil {
				if IsNatsErr(err, JSStreamNotFoundErr) {
					serv.Warnf("[tenant: %v][user: %v]createConsumerDirectCommon: Consumer %v at station %v: station does not exist", user.TenantName, user.Username, consumerName, cStationName)
				} else {
					serv.Errorf("[tenant: %v][user: %v]createConsumerDirectCommon at CreateConsumer: Consumer %v at station %v: %v", user.TenantName, user.Username, consumerName, cStationName, err.Error())
				}
				return err
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
	}
	return nil
}

func (s *Server) createConsumerDirect(c *client, reply string, msg []byte) {
	var ccr createConsumerRequestV1
	var resp createConsumerResponse

	tenantName, message, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("createConsumerDirect at getTenantNameAndMessage: %v", err.Error())
		return
	}

	if err := json.Unmarshal([]byte(message), &ccr); err != nil || ccr.RequestVersion < 1 {
		var ccrV0 createConsumerRequestV0
		if err := json.Unmarshal(msg, &ccrV0); err != nil {
			s.Errorf("[tenant: %v]createConsumerDirect at json.Unmarshal: Failed creating consumer: %v: %v", tenantName, err.Error(), string(msg))
			respondWithRespErr(serv.MemphisGlobalAccountString(), s, reply, err, &resp)
			return
		}
		s.createConsumerDirectV0(c, reply, tenantName, ccrV0, ccr.RequestVersion)
		return
	}

	ccr.TenantName = tenantName
	if ccr.StartConsumeFromSequence <= 0 {
		errMsg := errors.New("startConsumeFromSequence has to be a positive number")
		serv.Warnf("[tenant: %v]createConsumerDirect: %v", tenantName, errMsg.Error())
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, errMsg)
		return
	}

	if ccr.LastMessages < -1 {
		errMsg := errors.New("min value for LastMessages is -1")
		serv.Warnf("[tenant: %v]createConsumerDirect: %v", tenantName, errMsg.Error())
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, errMsg)
		return
	}

	if ccr.StartConsumeFromSequence > 1 && ccr.LastMessages > -1 {
		errMsg := errors.New("consumer creation options can't contain both startConsumeFromSequence and lastMessages")
		serv.Warnf("[tenant: %v]createConsumerDirect: %v", tenantName, errMsg.Error())
		respondWithErr(serv.MemphisGlobalAccountString(), s, reply, errMsg)
		return
	}

	err = s.createConsumerDirectCommon(c, ccr.Name, ccr.StationName, ccr.ConsumerGroup, ccr.ConsumerType, ccr.ConnectionId, tenantName, ccr.Username, ccr.MaxAckTimeMillis, ccr.MaxMsgDeliveries, 1, ccr.StartConsumeFromSequence, ccr.LastMessages)
	respondWithErr(serv.MemphisGlobalAccountString(), s, reply, err)
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
				ConnectedConsumers:    []models.ExtendedConsumer{},
				DisconnectedConsumers: []models.ExtendedConsumer{},
				DeletedConsumers:      []models.ExtendedConsumer{},
				IsActive:              consumer.IsActive,
				LastStatusChangeDate:  consumer.UpdatedAt,
			}
			m[consumer.ConsumersGroup] = cg
		} else {
			m[consumer.ConsumersGroup].Name = consumer.ConsumersGroup
			m[consumer.ConsumersGroup].MaxAckTimeMs = consumer.MaxAckTimeMs
			m[consumer.ConsumersGroup].MaxMsgDeliveries = consumer.MaxMsgDeliveries
			m[consumer.ConsumersGroup].IsActive = consumer.IsActive
			m[consumer.ConsumersGroup].LastStatusChangeDate = consumer.UpdatedAt
			cg = m[consumer.ConsumersGroup]
		}

		consumerRes := models.ExtendedConsumer{
			ID:               consumer.ID,
			Name:             consumer.Name,
			IsActive:         consumer.IsActive,
			ConsumersGroup:   consumer.ConsumersGroup,
			MaxAckTimeMs:     consumer.MaxAckTimeMs,
			MaxMsgDeliveries: consumer.MaxMsgDeliveries,
			StationName:      consumer.StationName,
			Count:            consumer.Count,
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

		cgInfo, err := ch.S.GetCgInfo(station.TenantName, stationName, cg.Name)
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

func (ch ConsumersHandler) GetDelayedCgsByTenant(tenantName string) ([]models.DelayedCgResp, error) {
	streams, err := ch.S.memphisAllStreamsInfo(tenantName)
	if err != nil {
		return []models.DelayedCgResp{}, err
	}
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
			if consumer.NumPending > 0 {
				stationName := StationNameFromStreamName(consumer.Stream)
				consumerName := revertDelimiters(consumer.Name)
				externalStationName := stationName.Ext()
				if _, ok := consumers[externalStationName]; !ok {
					consumers[externalStationName] = map[string]models.DelayedCg{consumerName: {CGName: consumerName, NumOfDelayedMsgs: uint64(consumer.NumPending)}}
				} else {
					consumers[externalStationName][consumerName] = models.DelayedCg{CGName: consumerName, NumOfDelayedMsgs: uint64(consumer.NumPending)}
				}
				consumerNames = append(consumerNames, consumerName)
			}
		}
	}
	consumersFromDb, err := db.GetActiveConsumersByName(consumerNames, tenantName)
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
		err = s.RemoveConsumer(station.TenantName, stationName, consumer.ConsumersGroup)
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
