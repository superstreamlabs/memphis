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
	"strings"
	"time"

	"github.com/memphisdev/memphis/analytics"
	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/memphis_cache"
	"github.com/memphisdev/memphis/models"

	"k8s.io/utils/strings/slices"
)

type ProducersHandler struct{ S *Server }

const (
	producerObjectName = "Producer"
)

func validateProducerName(name string) error {
	return validateName(name, producerObjectName)
}

func validateProducerType(producerType string) error {
	if producerType != "application" && producerType != "connector" {
		return errors.New("producer type has to be one of the following application/connector")
	}
	return nil
}

func (s *Server) createProducerDirectCommon(c *client, pName, pType, pConnectionId string, pStationName StationName, username string, tenantName string, version int, appId, sdkLang string) (bool, bool, error, models.Station) {
	name := strings.ToLower(pName)
	err := validateProducerName(name)
	if err != nil {
		serv.Warnf("createProducerDirectCommon at validateProducerName: Producer %v at station %v: %v", pName, pStationName.external, err.Error())
		return false, false, err, models.Station{}
	}
	producerType := strings.ToLower(pType)
	err = validateProducerType(producerType)
	if err != nil {
		serv.Warnf("createProducerDirectCommon at validateProducerType: Producer %v at station %v: %v", pName, pStationName.external, err.Error())
		return false, false, err, models.Station{}
	}

	exist, user, err := memphis_cache.GetUser(username, tenantName, false)
	if err != nil {
		serv.Errorf("createProducerDirectCommon at GetUser: Producer %v at station %v: %v", pName, pStationName.external, err.Error())
		return false, false, err, models.Station{}
	}
	if !exist {
		serv.Warnf("createProducerDirectCommon: User %v does not exist", username)
		return false, false, errors.New("User " + username + " does not exist"), models.Station{}
	}

	exist, station, err := db.GetStationByName(pStationName.Ext(), user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]createProducerDirectCommon at GetStationByName: Producer %v at station %v: %v", user.TenantName, user.Username, pName, pStationName.external, err.Error())
		return false, false, err, models.Station{}
	}
	if !exist {
		if version < 2 {
			err := errors.New("this station does not exist, a default station can not be created automatically, please upgrade your SDK version")
			serv.Warnf("[tenant: %v]createProducerDirectCommon : Producer %v at station %v : %v", user.TenantName, pName, pStationName, err.Error())
			return false, false, err, models.Station{}
		}
		var created bool
		station, created, err = CreateDefaultStation(user.TenantName, s, pStationName, user.ID, user.Username, _EMPTY_, 0)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "max amount") {
				serv.Warnf("[tenant: %v][user: %v]createProducerDirectCommon at CreateDefaultStation: creating default station error - producer %v at station %v: %v", user.TenantName, user.Username, pName, pStationName.external, err.Error())
			} else {
				serv.Errorf("[tenant: %v][user: %v]createProducerDirectCommon at CreateDefaultStation: creating default station error - producer %v at station %v: %v", user.TenantName, user.Username, pName, pStationName.external, err.Error())
			}
			return false, false, err, models.Station{}
		}
		if created {
			message := "Station " + pStationName.Ext() + " has been created by user " + user.Username
			serv.Noticef("[tenant: %v][user: %v]: %v", user.TenantName, user.Username, message)
			var auditLogs []interface{}
			newAuditLog := models.AuditLog{
				StationName:       pStationName.Ext(),
				Message:           message,
				CreatedBy:         user.ID,
				CreatedByUsername: user.Username,
				CreatedAt:         time.Now(),
				TenantName:        user.TenantName,
			}
			auditLogs = append(auditLogs, newAuditLog)
			err = CreateAuditLogs(auditLogs)
			if err != nil {
				serv.Errorf("[tenant: %v][user: %v]createProducerDirectCommon: Producer %v at station %v: %v", user.TenantName, user.Username, pName, pStationName.external, err.Error())
			}

			shouldSendAnalytics, _ := shouldSendAnalytics()
			if shouldSendAnalytics {
				analyticsParams := map[string]interface{}{"station-name": pStationName.Ext(), "storage-type": "disk"}
				analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-create-station-sdk")
			}
		}
	} else {
		if version < 2 && len(station.PartitionsList) > 0 {
			err := errors.New("to produce to this station please upgrade your SDK version")
			serv.Warnf("[tenant: %v]createProducerDirectCommon : Producer %v at station %v : %v", user.TenantName, pName, pStationName, err.Error())
			return false, false, err, models.Station{}
		}
	}

	err = validateProducersCount(station.ID, user.TenantName)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]createProducerDirectCommon at validateProducersCount at station %s: %v", user.TenantName, user.Username, pStationName.Ext(), err.Error())
		return false, false, err, models.Station{}
	}

	sdkName := sdkLang
	if sdkLang == "" {
		sdkName = c.opts.Lang
	}

	if strings.HasPrefix(user.Username, "$") && name != "gui" {
		_, err := db.InsertNewProducer(name, station.ID, "connector", pConnectionId, station.TenantName, station.PartitionsList, version, sdkName, appId)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]createProducerDirectCommon at InsertNewProducer: %v", user.TenantName, user.Username, err.Error())
			return false, false, err, models.Station{}
		}
	} else {
		newProducer, err := db.InsertNewProducer(name, station.ID, producerType, pConnectionId, station.TenantName, station.PartitionsList, version, sdkName, appId)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]createProducerDirectCommon at InsertNewProducer: %v", user.TenantName, user.Username, err.Error())
			return false, false, err, models.Station{}
		}
		message := "Producer " + name + " connected"
		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			StationName:       pStationName.Ext(),
			Message:           message,
			CreatedBy:         user.ID,
			CreatedByUsername: user.Username,
			CreatedAt:         time.Now(),
			TenantName:        user.TenantName,
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]createProducerDirectCommon at CreateAuditLogs: Producer %v at station %v: %v", user.TenantName, user.Username, pName, pStationName.external, err.Error())
			return false, false, err, models.Station{}
		}
		shouldSendAnalytics, _ := shouldSendAnalytics()
		if shouldSendAnalytics {
			ip := serv.getIp()
			analyticsParams := map[string]interface{}{"producer-name": newProducer.Name, "ip": ip}
			analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-create-producer-sdk")
			if strings.HasPrefix(newProducer.Name, "rest_gateway") {
				analyticsParams = map[string]interface{}{}
				analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-send-messages-via-rest-gw")
			}
		}
	}

	shouldSendNotifications := shouldSendNotification(user.TenantName, SchemaVAlert)
	return shouldSendNotifications, station.DlsConfigurationSchemaverse, nil, station
}

func (s *Server) createProducerDirectV0(c *client, reply string, cpr createProducerRequestV0, tenantName string) {
	sn, err := StationNameFromStr(cpr.StationName)
	if err != nil {
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	_, _, err, _ = s.createProducerDirectCommon(c, cpr.Name,
		cpr.ProducerType, cpr.ConnectionId, sn, cpr.Username, tenantName, 0, cpr.ConnectionId, "")
	respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
}

func (s *Server) createProducerDirect(c *client, reply string, msg []byte) {
	var cpr createProducerRequestV3
	var resp createProducerResponse

	tenantName, message, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("createProducerDirect at getTenantNameAndMessage: %v", err.Error())
		return
	}

	if err := json.Unmarshal([]byte(message), &cpr); err != nil || cpr.RequestVersion < 4 {
		var cprV2 createProducerRequestV2
		if err := json.Unmarshal([]byte(message), &cprV2); err != nil {
			var cprV1 createProducerRequestV1
			if err := json.Unmarshal([]byte(message), &cprV1); err != nil {
				var cprV0 createProducerRequestV0
				if err := json.Unmarshal([]byte(message), &cprV0); err != nil {
					s.Errorf("[tenant: %v]createProducerDirect: %v", tenantName, err.Error())
					respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
					return
				}
				s.createProducerDirectV0(c, reply, cprV0, tenantName)
				return
			}

			cpr = createProducerRequestV3{
				Name:           cprV1.Name,
				StationName:    cprV1.StationName,
				ConnectionId:   cprV1.ConnectionId,
				ProducerType:   cprV1.ProducerType,
				RequestVersion: cprV1.RequestVersion,
				Username:       cprV1.Username,
				AppId:          cprV1.ConnectionId,
			}
		}

		cpr = createProducerRequestV3{
			Name:           cprV2.Name,
			StationName:    cprV2.StationName,
			ConnectionId:   cprV2.ConnectionId,
			ProducerType:   cprV2.ProducerType,
			RequestVersion: cprV2.RequestVersion,
			Username:       cprV2.Username,
			AppId:          cprV2.ConnectionId,
		}
	}
	cpr.TenantName = tenantName
	sn, err := StationNameFromStr(cpr.StationName)
	if err != nil {
		s.Warnf("[tenant: %v][user: %v]createProducerDirect at StationNameFromStr: Producer %v at station %v: %v", cpr.TenantName, cpr.Username, cpr.Name, cpr.StationName, err.Error())
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
		return
	}

	clusterSendNotification, schemaVerseToDls, err, station := s.createProducerDirectCommon(c, cpr.Name, cpr.ProducerType, cpr.ConnectionId, sn, cpr.Username, tenantName, cpr.RequestVersion, cpr.AppId, cpr.SdkLang)
	if err != nil {
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
		return
	}

	firstFunctions, err := GetAllFirstActiveFunctionsIDByStationID(station.ID, tenantName)
	if err != nil {
		s.Errorf("[tenant: %v][user: %v]createProducerDirect at GetAllFirstActiveFunctionsIDByStationID: Producer %v at station %v: %v", cpr.TenantName, cpr.Username, cpr.Name, cpr.StationName, err.Error())
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, fmt.Errorf("got an error while getting the functions data"), &resp)
	}
	resp.StationPartitionsFirstFunctions = firstFunctions
	resp.StationVersion = station.Version
	partitions := models.PartitionsUpdate{PartitionsList: station.PartitionsList}
	resp.PartitionsUpdate = partitions
	resp.SchemaVerseToDls = schemaVerseToDls
	resp.ClusterSendNotification = clusterSendNotification
	schemaUpdate, err := getSchemaUpdateInitFromStation(sn, cpr.TenantName)
	if err == ErrNoSchema {
		respondWithResp(s.MemphisGlobalAccountString(), s, reply, &resp)
		return
	}
	if err != nil {
		s.Errorf("[tenant: %v][user: %v]createProducerDirect at getSchemaUpdateInitFromStation: Producer %v at station %v: %v", cpr.TenantName, cpr.Username, cpr.Name, cpr.StationName, err.Error())
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
		return
	}

	resp.SchemaUpdate = *schemaUpdate
	respondWithResp(s.MemphisGlobalAccountString(), s, reply, &resp)
}

func (ph ProducersHandler) GetProducersByStation(station models.Station) ([]models.ExtendedProducerResponse, []models.ExtendedProducerResponse, []models.ExtendedProducerResponse, error) { // for socket io endpoint
	producers, err := db.GetAllProducersByStationID(station.ID)
	if err != nil {
		return []models.ExtendedProducerResponse{}, []models.ExtendedProducerResponse{}, []models.ExtendedProducerResponse{}, err
	}

	var connectedProducers []models.ExtendedProducerResponse
	var disconnectedProducers []models.ExtendedProducerResponse
	var deletedProducers []models.ExtendedProducerResponse
	producersNames := []string{}

	for _, producer := range producers {
		if slices.Contains(producersNames, producer.Name) {
			continue
		}
		needToUpdateVersion := false
		if producer.Version < lastProducerCreationReqVersion && producer.IsActive {
			needToUpdateVersion = true
		}

		producerExtendedRes := models.ExtendedProducerResponse{
			ID:                         producer.ID,
			Name:                       producer.Name,
			StationName:                producer.StationName,
			UpdatedAt:                  producer.UpdatedAt,
			IsActive:                   producer.IsActive,
			DisconnedtedProducersCount: producer.DisconnedtedProducersCount,
			ConnectedProducersCount:    producer.ConnectedProducersCount,
			SdkLanguage:                producer.Sdk,
			UpdateAvailable:            needToUpdateVersion,
		}

		producersNames = append(producersNames, producer.Name)
		if producer.IsActive {
			connectedProducers = append(connectedProducers, producerExtendedRes)
		} else {
			disconnectedProducers = append(disconnectedProducers, producerExtendedRes)
		}
	}

	if len(connectedProducers) == 0 {
		connectedProducers = []models.ExtendedProducerResponse{}
	}

	if len(disconnectedProducers) == 0 {
		disconnectedProducers = []models.ExtendedProducerResponse{}
	}

	if len(deletedProducers) == 0 {
		deletedProducers = []models.ExtendedProducerResponse{}
	}

	sort.Slice(connectedProducers, func(i, j int) bool {
		return connectedProducers[j].UpdatedAt.Before(connectedProducers[i].UpdatedAt)
	})
	sort.Slice(disconnectedProducers, func(i, j int) bool {
		return disconnectedProducers[j].UpdatedAt.Before(disconnectedProducers[i].UpdatedAt)
	})
	sort.Slice(deletedProducers, func(i, j int) bool {
		return deletedProducers[j].UpdatedAt.Before(deletedProducers[i].UpdatedAt)
	})
	return connectedProducers, disconnectedProducers, deletedProducers, nil
}

func (s *Server) destroyProducerDirect(c *client, reply string, msg []byte) {
	var dpr destroyProducerRequestV1
	tenantName, destoryMessage, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("destroyProducerDirect at getTenantNameAndMessage: %v", err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	if err := json.Unmarshal([]byte(destoryMessage), &dpr); err != nil || dpr.RequestVersion < 1 {
		var dprV0 destroyProducerRequestV0
		if err := json.Unmarshal([]byte(destoryMessage), &dprV0); err != nil {
			s.Errorf("destroyProducerDirect: %v", err.Error())
			respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
			return
		}
		dprV0.TenantName = tenantName
		if c.memphisInfo.connectionId == _EMPTY_ {
			s.destroyProducerDirectV0(c, reply, dprV0)
			return
		} else {
			dpr = destroyProducerRequestV1{
				StationName:    dprV0.StationName,
				ProducerName:   dprV0.ProducerName,
				Username:       dprV0.Username,
				ConnectionId:   c.memphisInfo.connectionId,
				RequestVersion: 1,
			}
		}
	}

	dpr.TenantName = tenantName
	stationName, err := StationNameFromStr(dpr.StationName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]destroyProducerDirect at StationNameFromStr: Producer %v at station %v: %v", dpr.TenantName, dpr.Username, dpr.ProducerName, dpr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	name := strings.ToLower(dpr.ProducerName)
	_, station, err := db.GetStationByName(stationName.Ext(), dpr.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]destroyProducerDirect at GetStationByName: Producer %v at station %v: %v", dpr.TenantName, dpr.Username, dpr.ProducerName, dpr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	exist, err := db.DeleteProducerByNameStationIDAndConnID(name, station.ID, dpr.ConnectionId)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]destroyProducerDirect at DeleteProducerByNameStationIDAndConnID: Producer %v at station %v: %v", dpr.TenantName, dpr.Username, name, dpr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Producer %v at station %v does not exist", name, dpr.StationName)
		serv.Warnf("[tenant: %v][user: %v]destroyProducerDirect: %v", dpr.TenantName, dpr.Username, errMsg)
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, errors.New(errMsg))
		return
	}

	username := c.memphisInfo.username
	if username == _EMPTY_ {
		username = dpr.Username
	}
	_, user, err := memphis_cache.GetUser(username, dpr.TenantName, false)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]destroyProducerDirect at GetUser: Producer %v at station %v: %v", dpr.TenantName, dpr.Username, name, dpr.StationName, err.Error())
	}
	message := "Producer " + name + " has been destroyed"
	serv.Noticef("[tenant: %v][user: %v]: %v", tenantName, username, message)
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
		serv.Errorf("[tenant: %v][user: %v]destroyProducerDirect at CreateAuditLogs: Producer %v at station %v: %v", dpr.TenantName, dpr.Username, name, dpr.StationName, err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, username, analyticsParams, "user-remove-producer-sdk")
	}

	respondWithErr(s.MemphisGlobalAccountString(), s, reply, nil)
}

func (s *Server) destroyProducerDirectV0(c *client, reply string, dpr destroyProducerRequestV0) {
	stationName, err := StationNameFromStr(dpr.StationName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]destroyProducerDirectV0 at StationNameFromStr: Producer %v at station %v: %v", dpr.TenantName, dpr.Username, dpr.ProducerName, dpr.StationName, err.Error())
		respondWithErr(MEMPHIS_GLOBAL_ACCOUNT, s, reply, err)
		return
	}
	name := strings.ToLower(dpr.ProducerName)
	_, station, err := db.GetStationByName(stationName.Ext(), dpr.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]destroyProducerDirectV0 at GetStationByName: Producer %v at station %v: %v", dpr.TenantName, dpr.Username, dpr.ProducerName, dpr.StationName, err.Error())
		respondWithErr(MEMPHIS_GLOBAL_ACCOUNT, s, reply, err)
		return
	}

	exist, err := db.DeleteProducerByNameAndStationID(name, station.ID)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]destroyProducerDirectV0 at DeleteProducerByNameStationIDAndConnID: Producer %v at station %v: %v", dpr.TenantName, dpr.Username, name, dpr.StationName, err.Error())
		respondWithErr(MEMPHIS_GLOBAL_ACCOUNT, s, reply, err)
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Producer %v at station %v does not exist", name, dpr.StationName)
		serv.Warnf("[tenant: %v][user: %v]destroyProducerDirectV0: %v", dpr.TenantName, dpr.Username, errMsg)
		respondWithErr(MEMPHIS_GLOBAL_ACCOUNT, s, reply, errors.New(errMsg))
		return
	}

	username := c.memphisInfo.username
	if username == _EMPTY_ {
		username = dpr.Username
	}
	_, user, err := memphis_cache.GetUser(username, dpr.TenantName, false)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]destroyProducerDirectV0 at GetUser: Producer %v at station %v: %v", dpr.TenantName, dpr.Username, name, dpr.StationName, err.Error())
	}
	message := "Producer " + name + " has been destroyed"
	serv.Noticef("[tenant: %v][user: %v]: %v", dpr.TenantName, username, message)
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
		serv.Errorf("[tenant: %v][user: %v]destroyProducerDirectV0 at CreateAuditLogs: Producer %v at station %v: %v", dpr.TenantName, dpr.Username, name, dpr.StationName, err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, username, analyticsParams, "user-remove-producer-sdk")
	}

	respondWithErr(MEMPHIS_GLOBAL_ACCOUNT, s, reply, nil)
}
