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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"memphis/analytics"
	"memphis/db"
	"memphis/models"
	"memphis/utils"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type StationsHandler struct{ S *Server }

const (
	stationObjectName       = "Station"
	schemaToDlsUpdateType   = "schemaverse_to_dls"
	removeStationUpdateType = "remove_station"
)

type StationName struct {
	internal string
	external string
}

func (sn StationName) Ext() string {
	return sn.external
}

func (sn StationName) Intern() string {
	return sn.internal
}

func StationNameFromStr(name string) (StationName, error) {
	var intern, extern string
	if strings.Contains(name, delimiterReplacement) {
		intern = strings.ToLower(name)
		err := validateName(intern, stationObjectName)
		if err != nil {
			return StationName{}, err
		}
		extern = revertDelimiters(name)
		extern = strings.ToLower(extern)

	} else {
		extern = strings.ToLower(name)
		err := validateName(extern, stationObjectName)
		if err != nil {
			return StationName{}, err
		}

		intern = replaceDelimiters(name)
		intern = strings.ToLower(intern)
	}

	return StationName{internal: intern, external: extern}, nil
}

func StationNameFromStreamName(streamName string) StationName {
	intern := streamName
	extern := revertDelimiters(intern)

	return StationName{internal: intern, external: extern}
}

func validateRetentionType(retentionType string) error {
	if retentionType != "message_age_sec" && retentionType != "messages" && retentionType != "bytes" {
		return errors.New("retention type can be one of the following message_age_sec/messages/bytes")
	}

	return nil
}

func validateStorageType(storageType string) error {
	if storageType != "file" && storageType != "memory" {
		return errors.New("storage type can be one of the following file/memory")
	}

	return nil
}

func validateIdempotencyWindow(retentionType string, retentionValue int, idempotencyWindow int64) error {
	if idempotencyWindow > 86400000 { // 24 hours
		return errors.New("idempotency window can not exceed 24 hours")
	}
	if retentionType == "message_age_sec" && (int64(retentionValue)*1000 < idempotencyWindow) {
		return errors.New("idempotency window cannot be greater than the station retention")
	}

	return nil
}

// TODO remove the station resources - functions, connectors
func removeStationResources(s *Server, station models.Station, shouldDeleteStream bool) error {
	stationName, err := StationNameFromStr(station.Name)
	if err != nil {
		return err
	}

	if shouldDeleteStream {
		err = s.RemoveStream(station.TenantName, stationName.Intern())
		if err != nil && !IsNatsErr(err, JSStreamNotFoundErr) {
			return err
		}
	}

	DeleteTagsFromStation(station.ID)

	err = db.DeleteDLSMessagesByStationID(station.ID)
	if err != nil {
		return err
	}

	err = db.DeleteProducersByStationID(station.ID)
	if err != nil {
		return err
	}

	err = db.DeleteAllConsumersByStationID(station.ID)
	if err != nil {
		return err
	}

	err = RemoveAllAuditLogsByStation(station.Name, station.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v]removeStationResources: Station %v: %v", station.TenantName, station.Name, err.Error())
	}

	return nil
}

func (s *Server) createStationDirect(c *client, reply string, msg []byte) {
	var csr createStationRequest
	var tenantName string
	tenantName, message, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("[tenant: %v]createStationDirect at getTenantNameAndMessage: %v", tenantName, err.Error())
		return
	}
	if err := json.Unmarshal([]byte(message), &csr); err != nil {
		s.Errorf("[tenant: %v]createStationDirect: failed creating station: %v", tenantName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	csr.TenantName = tenantName
	s.createStationDirectIntern(c, reply, &csr, true)
}

func (s *Server) createStationDirectIntern(c *client,
	reply string,
	csr *createStationRequest,
	shouldCreateStream bool) {
	isNative := shouldCreateStream
	jsApiResp := JSApiStreamCreateResponse{ApiResponse: ApiResponse{Type: JSApiStreamCreateResponseType}}
	memphisGlobalAcc := s.MemphisGlobalAccount()
	stationName, err := StationNameFromStr(csr.StationName)
	if err != nil {
		serv.Warnf("[tenant: %v][user:%v]createStationDirect at StationNameFromStr: Station %v: %v", csr.TenantName, csr.Username, csr.StationName, err.Error())
		jsApiResp.Error = NewJSStreamCreateError(err)
		respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}

	// for NATS compatibility
	username, tenantId, err := getUserAndTenantIdFromString(csr.Username)
	if err != nil {
		serv.Warnf("[tenant: %v][user:%v]createStationDirect at getUserAndTenantIdFromString: Station %v: %v", csr.TenantName, csr.Username, csr.StationName, err.Error())
		jsApiResp.Error = NewJSStreamCreateError(err)
		respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}
	if tenantId != -1 {
		exist, t, err := db.GetTenantById(tenantId)
		if err != nil {
			serv.Warnf("[tenant: %v][user:%v]createStationDirect at db.GetTenantById: Station %v: %v", csr.TenantName, csr.Username, csr.StationName, err.Error())
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}
		if !exist {
			msg := fmt.Sprintf("createStationDirect: Station %v: Tenant with id %v does not exist", csr.StationName, strconv.Itoa(tenantId))
			serv.Warnf("[tenant: %v][user:%v] %v", csr.TenantName, csr.Username, msg)
			err = errors.New(msg)
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}
		csr.TenantName = t.Name
	}

	exist, _, err := db.GetStationByName(stationName.Ext(), csr.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user:%v]createStationDirect at db.GetStationByName: Station %v: %v", csr.TenantName, csr.Username, csr.StationName, err.Error())
		jsApiResp.Error = NewJSStreamCreateError(err)
		respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}

	if exist {
		errMsg := fmt.Sprintf("Station %v already exists", stationName.Ext())
		serv.Warnf("[tenant: %v][user:%v]createStationDirect: %v", csr.TenantName, csr.Username, errMsg)
		jsApiResp.Error = NewJSStreamNameExistError()
		respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}

	schemaName := csr.SchemaName
	var schemaDetails models.SchemaDetails
	if schemaName != "" {
		schemaName = strings.ToLower(csr.SchemaName)
		exist, schema, err := db.GetSchemaByName(schemaName, csr.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user:%v]createStationDirect db.GetSchemaByName: Station %v: %v", csr.TenantName, csr.Username, csr.StationName, err.Error())
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}
		if !exist {
			errMsg := fmt.Sprintf("Schema %v does not exist", csr.SchemaName)
			serv.Warnf("[tenant: %v][user:%v]createStationDirect: %v", csr.TenantName, csr.Username, errMsg)
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}

		schemaVersion, err := getActiveVersionBySchemaId(schema.ID)
		if err != nil {
			serv.Errorf("[tenant: %v][user:%v]createStationDirect at getActiveVersionBySchemaId: Station %v: %v", csr.TenantName, csr.Username, csr.StationName, err.Error())
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}
		schemaDetails = models.SchemaDetails{SchemaName: schemaName, VersionNumber: schemaVersion.VersionNumber}
	}

	var retentionType string
	var retentionValue int
	if csr.RetentionType != "" {
		retentionType = strings.ToLower(csr.RetentionType)
		err = validateRetentionType(retentionType)
		if err != nil {
			serv.Warnf("[tenant: %v][user:%v]createStationDirect at validateRetentionType: %v", csr.TenantName, csr.Username, err.Error())
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}
		retentionValue = csr.RetentionValue
	} else {
		retentionType = "message_age_sec"
		retentionValue = 604800 // 1 week
	}

	var storageType string
	if csr.StorageType != "" {
		storageType = getStationStorageType(csr.StorageType)
		err = validateStorageType(storageType)
		if err != nil {
			serv.Warnf("[tenant: %v][user:%v]createStationDirect at validateStorageType: %v", csr.TenantName, csr.Username, err.Error())
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}
	} else {
		storageType = "file"
	}

	replicas := getStationReplicas(csr.Replicas)
	err = validateReplicas(replicas)
	if err != nil {
		serv.Warnf("[tenant: %v][user:%v]createStationDirect at validateReplicas: %v", csr.TenantName, csr.Username, err.Error())
		jsApiResp.Error = NewJSStreamCreateError(err)
		respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}

	if csr.IdempotencyWindow <= 0 {
		csr.IdempotencyWindow = 120000 // default
	} else if csr.IdempotencyWindow < 100 {
		csr.IdempotencyWindow = 100 // minimum is 100 millis
	}

	err = validateIdempotencyWindow(retentionType, retentionValue, csr.IdempotencyWindow)
	if err != nil {
		serv.Warnf("[tenant: %v][user:%v]createStationDirect at validateIdempotencyWindow: %v", csr.TenantName, csr.Username, err.Error())
		jsApiResp.Error = NewJSStreamCreateError(err)
		respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}

	if shouldCreateStream {
		err = s.CreateStream(csr.TenantName, stationName, retentionType, retentionValue, storageType, csr.IdempotencyWindow, replicas, csr.TieredStorageEnabled)
		if err != nil {
			if IsNatsErr(err, JSStreamReplicasNotSupportedErr) {
				serv.Warnf("[tenant: %v][user:%v]CreateStationDirect: Station %v: Station can not be created, probably since replicas count is larger than the cluster size", csr.TenantName, csr.Username, stationName.Ext())
				respondWithErr(s.MemphisGlobalAccountString(), s, reply, errors.New("station can not be created, probably since replicas count is larger than the cluster size"))
				return
			}

			serv.Errorf("[tenant: %v][user:%v]createStationDirect: Station %v: %v", csr.TenantName, csr.Username, csr.StationName, err.Error())
			respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
			return
		}
	}

	exist, user, err := db.GetUserByUsername(username, csr.TenantName)
	if err != nil {
		serv.Warnf("[tenant: %v][user:%v]createStationDirect at GetUserByUsername: Station %v: %v", csr.TenantName, csr.Username, csr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	if !exist {
		serv.Warnf("[tenant: %v][user:%v]createStationDirect at GetUserByUsername: user %v is not exists", csr.TenantName, csr.Username, csr.Username)
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	_, rowsUpdated, err := db.InsertNewStation(stationName.Ext(), user.ID, user.Username, retentionType, retentionValue, storageType, replicas, schemaDetails.SchemaName, schemaDetails.VersionNumber, csr.IdempotencyWindow, isNative, csr.DlsConfiguration, csr.TieredStorageEnabled, user.TenantName)
	if err != nil {
		if !strings.Contains(err.Error(), "already exist") {
			serv.Errorf("[tenant: %v][user:%v]createStationDirect at InsertNewStation: Station %v: %v", csr.TenantName, csr.Username, csr.StationName, err.Error())
		}
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	if rowsUpdated > 0 {
		message := "Station " + stationName.Ext() + " has been created by user " + username
		serv.Noticef("[tenant:%v][user: %v] %v", user.TenantName, user.Username, message)
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
			serv.Errorf("[tenant: %v][user:%v]createStationDirect: Station %v - create audit logs error: %v", csr.TenantName, csr.Username, csr.StationName, err.Error())
		}

		shouldSendAnalytics, _ := shouldSendAnalytics()
		if shouldSendAnalytics {
			storageType = "memory"
			if storageType == "file" {
				storageType = "disk"
			}
			analyticsParams := map[string]interface{}{"station-name": stationName.Ext(), "tiered-storage": strconv.FormatBool(csr.TieredStorageEnabled), "nats-comp": strconv.FormatBool(!isNative), "storage-type": storageType}
			analytics.SendEvent(user.TenantName, username, analyticsParams, "user-create-station-sdk")
		}
	}

	respondWithErr(s.MemphisGlobalAccountString(), s, reply, nil)
}

func (sh StationsHandler) GetStation(c *gin.Context) {
	var body models.GetStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	tagsHandler := TagsHandler{S: sh.S}
	stationName := strings.ToLower(body.StationName)
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetStation at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	exist, station, err := db.GetStationByName(stationName, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetStation at GetStationByName: Station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	} else if !exist {
		errMsg := fmt.Sprintf("Station %v does not exist", body.StationName)
		serv.Warnf("[tenant: %v][user: %v]GetStation: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}
	tags, err := tagsHandler.GetTagsByEntityWithID("station", station.ID)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetStation: Station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if station.StorageType == "file" {
		station.StorageType = "disk"
	}

	if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(user.TenantName); !ok {
		station.TieredStorageEnabled = false
	} else {
		_, ok = tenantInetgrations["s3"].(models.Integration)
		if !ok {
			station.TieredStorageEnabled = false
		} else if station.TieredStorageEnabled {
			station.TieredStorageEnabled = true
		} else {
			station.TieredStorageEnabled = false
		}
	}

	stationResponse := models.GetStationResponseSchema{
		ID:                   station.ID,
		Name:                 station.Name,
		RetentionType:        station.RetentionType,
		RetentionValue:       station.RetentionValue,
		StorageType:          station.StorageType,
		Replicas:             station.Replicas,
		CreatedBy:            station.CreatedBy,
		CreatedByUsername:    station.CreatedByUsername,
		CreatedAt:            station.CreatedAt,
		LastUpdate:           station.UpdatedAt,
		IsDeleted:            station.IsDeleted,
		IdempotencyWindow:    station.IdempotencyWindow,
		IsNative:             station.IsNative,
		DlsConfiguration:     models.DlsConfiguration{Poison: station.DlsConfigurationPoison, Schemaverse: station.DlsConfigurationSchemaverse},
		TieredStorageEnabled: station.TieredStorageEnabled,
		Tags:                 tags,
	}

	c.IndentedJSON(200, stationResponse)
}

func (sh StationsHandler) GetStationsDetails(tenantName string) ([]models.ExtendedStationDetails, error) {
	var exStations []models.ExtendedStationDetails
	stations, err := db.GetActiveStationsPerTenant(tenantName)
	if err != nil {
		return []models.ExtendedStationDetails{}, err
	}
	stationTotalMsgs := make(map[string]int)
	if len(stations) == 0 {
		return []models.ExtendedStationDetails{}, nil
	} else {
		allStreamInfo, err := serv.memphisAllStreamsInfo(tenantName)
		if err != nil {
			return []models.ExtendedStationDetails{}, err
		}
		for _, info := range allStreamInfo {
			streamName := info.Config.Name
			if !strings.Contains(streamName, "$memphis") {
				stationTotalMsgs[streamName] = int(info.State.Msgs)
			}
		}
		stationIdsDlsMsgs, err := db.GetStationIdsFromDlsMsgs(tenantName)
		if err != nil {
			return []models.ExtendedStationDetails{}, err
		}
		tagsHandler := TagsHandler{S: sh.S}
		for _, station := range stations {
			_, hasDlsMsgs := stationIdsDlsMsgs[station.ID]
			tags, err := tagsHandler.GetTagsByEntityWithID("station", station.ID)
			if err != nil {
				return []models.ExtendedStationDetails{}, err
			}
			if station.StorageType == "file" {
				station.StorageType = "disk"
			}
			fullStationName, err := StationNameFromStr(station.Name)
			if err != nil {
				return []models.ExtendedStationDetails{}, err
			}
			totalMsgInfo := stationTotalMsgs[fullStationName.Intern()]

			activity := false
			activeCount, err := db.CountActiveProudcersByStationID(station.ID)
			if err != nil {
				return []models.ExtendedStationDetails{}, err
			}
			if activeCount > 0 {
				activity = true
			} else {
				activeCount, err = db.CountActiveConsumersByStationID(station.ID)
				if err != nil {
					return []models.ExtendedStationDetails{}, err
				}
				if activeCount > 0 {
					activity = true
				}
			}
			if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(tenantName); !ok {
				station.TieredStorageEnabled = false
			} else {
				_, ok = tenantInetgrations["s3"].(models.Integration)
				if !ok {
					station.TieredStorageEnabled = false
				} else if station.TieredStorageEnabled {
					station.TieredStorageEnabled = true
				} else {
					station.TieredStorageEnabled = false
				}
			}

			stationRes := models.Station{
				ID:                   station.ID,
				Name:                 station.Name,
				RetentionType:        station.RetentionType,
				RetentionValue:       station.RetentionValue,
				StorageType:          station.StorageType,
				Replicas:             station.Replicas,
				CreatedByUsername:    station.CreatedByUsername,
				CreatedAt:            station.CreatedAt,
				SchemaName:           station.SchemaName,
				IsNative:             station.IsNative,
				TieredStorageEnabled: station.TieredStorageEnabled,
			}

			exStations = append(exStations, models.ExtendedStationDetails{Station: stationRes, HasDlsMsgs: hasDlsMsgs, TotalMessages: totalMsgInfo, Tags: tags, Activity: activity})
		}
		if exStations == nil {
			return []models.ExtendedStationDetails{}, nil
		}
		return exStations, nil
	}
}

// TODO: check if need to remove
func (sh StationsHandler) GetAllStationsDetails(shouldGetTags bool, tenantName string) ([]models.ExtendedStation, uint64, uint64, error) {
	var stations []models.ExtendedStation
	totalMessages := uint64(0)
	if tenantName == "" {
		tenantName = serv.MemphisGlobalAccountString()
	}
	totalDlsMessages, err := db.GetTotalDlsMessages(tenantName)
	if err != nil {
		return []models.ExtendedStation{}, totalMessages, totalDlsMessages, err
	}

	stations, err = db.GetAllStationsDetailsPerTenant(tenantName)
	if err != nil {
		return stations, totalMessages, totalDlsMessages, err
	}
	if len(stations) == 0 {
		return []models.ExtendedStation{}, totalMessages, totalDlsMessages, nil
	} else {
		stationTotalMsgs := make(map[string]int)
		tagsHandler := TagsHandler{S: sh.S}
		acc, err := sh.S.lookupAccount(tenantName)
		if err != nil {
			return []models.ExtendedStation{}, totalMessages, totalDlsMessages, err
		}
		accName := acc.Name
		allStreamInfo, err := serv.memphisAllStreamsInfo(accName)
		if err != nil {
			return []models.ExtendedStation{}, totalMessages, totalDlsMessages, err
		}
		for _, info := range allStreamInfo {
			streamName := info.Config.Name
			if !strings.Contains(streamName, "$memphis") {
				totalMessages += info.State.Msgs
				stationTotalMsgs[streamName] = int(info.State.Msgs)
			}
		}

		stationIdsDlsMsgs, err := db.GetStationIdsFromDlsMsgs(tenantName)
		if err != nil {
			return []models.ExtendedStation{}, totalMessages, totalDlsMessages, err
		}

		var extStations []models.ExtendedStation
		for i := 0; i < len(stations); i++ {
			fullStationName, err := StationNameFromStr(stations[i].Name)
			if err != nil {
				return []models.ExtendedStation{}, totalMessages, totalDlsMessages, err
			}
			_, hasDlsMsgs := stationIdsDlsMsgs[stations[i].ID]
			if shouldGetTags {
				tags, err := tagsHandler.GetTagsByEntityWithID("station", stations[i].ID)
				if err != nil {
					return []models.ExtendedStation{}, totalMessages, totalDlsMessages, err
				}
				stations[i].Tags = tags
			}

			stations[i].TotalMessages = stationTotalMsgs[fullStationName.Intern()]
			stations[i].HasDlsMsgs = hasDlsMsgs

			found := false
			for _, p := range stations[i].Producers {
				if p.IsActive {
					stations[i].Activity = true
					found = true
					break
				}
			}

			if !found {
				for _, c := range stations[i].Consumers {
					if c.IsActive {
						stations[i].Activity = true
						break
					}
				}
			}
			if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(tenantName); !ok {
				stations[i].TieredStorageEnabled = false
			} else {
				_, ok = tenantInetgrations["s3"].(models.Integration)
				if !ok {
					stations[i].TieredStorageEnabled = false
				} else if stations[i].TieredStorageEnabled {
					stations[i].TieredStorageEnabled = true
				} else {
					stations[i].TieredStorageEnabled = false
				}
			}

			stationRes := models.ExtendedStation{
				ID:            stations[i].ID,
				Name:          stations[i].Name,
				CreatedAt:     stations[i].CreatedAt,
				TotalMessages: stations[i].TotalMessages,
				HasDlsMsgs:    stations[i].HasDlsMsgs,
				Activity:      stations[i].Activity,
				IsNative:      stations[i].IsNative,
			}

			extStations = append(extStations, stationRes)
		}
		return extStations, totalMessages, totalDlsMessages, nil
	}
}

func (sh StationsHandler) GetAllStationsDetailsLight(shouldExtend bool, tenantName string) ([]models.ExtendedStationLight, uint64, uint64, error) {
	var stations []models.ExtendedStationLight
	totalMessages := uint64(0)
	if tenantName == "" {
		tenantName = serv.MemphisGlobalAccountString()
	}
	totalDlsMessages, err := db.GetTotalDlsMessages(tenantName)
	if err != nil {
		return []models.ExtendedStationLight{}, totalMessages, totalDlsMessages, err
	}

	stations, err = db.GetAllStationsDetailsLight(tenantName)
	if err != nil {
		return stations, totalMessages, totalDlsMessages, err
	}
	if len(stations) == 0 {
		return []models.ExtendedStationLight{}, totalMessages, totalDlsMessages, nil
	} else {
		stationTotalMsgs := make(map[string]int)
		tagsHandler := TagsHandler{S: sh.S}
		acc, err := sh.S.lookupAccount(tenantName)
		if err != nil {
			return []models.ExtendedStationLight{}, totalMessages, totalDlsMessages, err
		}
		accName := acc.Name
		allStreamInfo, err := serv.memphisAllStreamsInfo(accName)
		if err != nil {
			return []models.ExtendedStationLight{}, totalMessages, totalDlsMessages, err
		}
		for _, info := range allStreamInfo {
			streamName := info.Config.Name
			if !strings.Contains(streamName, "$memphis") {
				totalMessages += info.State.Msgs
				stationTotalMsgs[streamName] = int(info.State.Msgs)
			}
		}
		stationIdsDlsMsgs, err := db.GetStationIdsFromDlsMsgs(tenantName)
		if err != nil {
			return []models.ExtendedStationLight{}, totalMessages, totalDlsMessages, err
		}

		var extStations []models.ExtendedStationLight
		for i := 0; i < len(stations); i++ {
			fullStationName, err := StationNameFromStr(stations[i].Name)
			if err != nil {
				return []models.ExtendedStationLight{}, totalMessages, totalDlsMessages, err
			}
			_, hasDlsMsgs := stationIdsDlsMsgs[stations[i].ID]
			if shouldExtend {
				tags, err := tagsHandler.GetTagsByEntityWithID("station", stations[i].ID)
				if err != nil {
					return []models.ExtendedStationLight{}, totalMessages, totalDlsMessages, err
				}
				stations[i].Tags = tags

				if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(tenantName); !ok {
					stations[i].TieredStorageEnabled = false
				} else {
					_, ok = tenantInetgrations["s3"].(models.Integration)
					if !ok {
						stations[i].TieredStorageEnabled = false
					} else if stations[i].TieredStorageEnabled {
						stations[i].TieredStorageEnabled = true
					} else {
						stations[i].TieredStorageEnabled = false
					}
				}
			}

			stations[i].TotalMessages = stationTotalMsgs[fullStationName.Intern()]
			stations[i].HasDlsMsgs = hasDlsMsgs

			extStations = append(extStations, stations[i])
		}
		return extStations, totalMessages, totalDlsMessages, nil
	}
}

func (sh StationsHandler) GetStations(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Warnf("GetStations at getUserDetailsFromMiddleware: Station %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	stations, err := sh.GetStationsDetails(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetStations at GetStationsDetails: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-enter-stations-page")
	}

	c.IndentedJSON(200, gin.H{
		"stations": stations,
	})
}

func (sh StationsHandler) GetAllStations(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetAllStations at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	stations, _, _, err := sh.GetAllStationsDetailsLight(true, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetAllStations at GetAllStationsDetails: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, stations)
}

func (sh StationsHandler) CreateStation(c *gin.Context) {
	var body models.CreateStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	tenantName := user.TenantName
	if err != nil {
		serv.Errorf("CreateStation at getUserDetailsFromMiddleware: At station %v: %v", body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	stationName, err := StationNameFromStr(body.Name)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]CreateStation at StationNameFromStr: Station %v: %v", user.TenantName, user.Username, body.Name, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, _, err := db.GetStationByName(stationName.Ext(), tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateStation GetStationByName: Station %v: %v", user.TenantName, user.Username, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		errMsg := fmt.Sprintf("Station %v already exists", stationName.external)
		serv.Warnf("[tenant: %v][user: %v]CreateStation: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	var schemaVersionNumber int
	schemaName := body.SchemaName
	if schemaName != "" {
		schemaName = strings.ToLower(body.SchemaName)
		exist, schema, err := db.GetSchemaByName(schemaName, tenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]CreateStation at GetSchemaByName: Station %v: %v", user.TenantName, user.Username, body.Name, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
			return
		}
		if !exist {
			errMsg := fmt.Sprintf("Schema %v does not exist", schemaName)
			serv.Warnf("[tenant: %v][user: %v]CreateStation: Station %v: %v", user.TenantName, user.Username, body.Name, errMsg)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
			return
		}

		schemaVersion, err := getActiveVersionBySchemaId(schema.ID)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]CreateStation at getActiveVersionBySchemaId: Station %v: %v", user.TenantName, user.Username, body.Name, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
			return
		}

		schemaVersionNumber = schemaVersion.VersionNumber
	} else {
		schemaName = ""
		schemaVersionNumber = 0
	}

	var retentionType string
	if body.RetentionType != "" && body.RetentionValue > 0 {
		retentionType = strings.ToLower(body.RetentionType)
		err = validateRetentionType(retentionType)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]CreateStation at validateRetentionType: Station %v: %v", user.TenantName, user.Username, body.Name, err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	} else {
		retentionType = "message_age_sec"
		body.RetentionValue = 604800 // 1 week
	}

	if body.StorageType != "" {
		body.StorageType = getStationStorageType(body.StorageType)
		err = validateStorageType(body.StorageType)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]CreateStation at validateStorageType: Station %v: %v", user.TenantName, user.Username, body.Name, err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	} else {
		body.StorageType = "file"
	}

	storageTypeForResponse := "disk"
	if body.StorageType == "memory" {
		storageTypeForResponse = body.StorageType
	}

	body.Replicas = getStationReplicas(body.Replicas)
	err = validateReplicas(body.Replicas)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]CreateStation at validateReplicas: Station %v: %v", user.TenantName, user.Username, body.Name, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	err = validateIdempotencyWindow(body.RetentionType, body.RetentionValue, body.IdempotencyWindow)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]CreateStation at validateIdempotencyWindow: Station %v: %v", user.TenantName, user.Username, body.Name, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	if body.IdempotencyWindow <= 0 {
		body.IdempotencyWindow = 120000 // default
	} else if body.IdempotencyWindow < 100 {
		body.IdempotencyWindow = 100 // minimum is 100 millis
	}

	newStation, rowsUpdated, err := db.InsertNewStation(stationName.Ext(), user.ID, user.Username, retentionType, body.RetentionValue, body.StorageType, body.Replicas, schemaName, schemaVersionNumber, body.IdempotencyWindow, true, body.DlsConfiguration, body.TieredStorageEnabled, tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateStation at db.InsertNewStation: Station %v: %v", user.TenantName, user.Username, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	//rowsUpdated == 0 means that the row already exists
	if rowsUpdated == 0 {
		errMsg := fmt.Sprintf("Station %v already exists", newStation.Name)
		serv.Warnf("[tenant: %v][user: %v]CreateStation: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	err = sh.S.CreateStream(tenantName, stationName, retentionType, body.RetentionValue, body.StorageType, body.IdempotencyWindow, body.Replicas, body.TieredStorageEnabled)
	if err != nil {
		if IsNatsErr(err, JSInsufficientResourcesErr) {
			serv.Warnf("[tenant: %v][user: %v]CreateStation: Station %v: Station can not be created, probably since replicas count is larger than the cluster size", user.TenantName, user.Username, body.Name)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station can not be created, probably since replicas count is larger than the cluster size"})
			return
		}

		serv.Errorf("[tenant: %v][user: %v]CreateStation at CreateStream: Station %v: %v", user.TenantName, user.Username, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(body.Tags) > 0 {
		err = AddTagsToEntity(body.Tags, "station", newStation.ID, newStation.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]CreateStation: : Station %v Failed adding tags: %v", user.TenantName, user.Username, body.Name, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}
	message := "Station " + stationName.Ext() + " has been created by " + user.Username
	serv.Noticef("[tenant: %v][user: %v] %v ", user.TenantName, user.Username, message)
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
		serv.Errorf("[tenant: %v][user: %v]CreateStation at CreateAuditLogs: Station %v: %v", user.TenantName, user.Username, body.Name, err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := map[string]interface{}{"station-name": stationName.Ext(), "tiered-storage": strconv.FormatBool(newStation.TieredStorageEnabled), "storage-type": storageTypeForResponse}
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-create-station")
	}

	if schemaName != "" {
		c.IndentedJSON(200, gin.H{
			"id":                            newStation.ID,
			"name":                          stationName.Ext(),
			"retention_type":                retentionType,
			"retention_value":               body.RetentionValue,
			"storage_type":                  storageTypeForResponse,
			"replicas":                      body.Replicas,
			"created_by_username":           user.Username,
			"created_at":                    time.Now(),
			"last_update":                   time.Now(),
			"is_deleted":                    false,
			"schema_name":                   schemaName,
			"schema_version_number":         schemaVersionNumber,
			"idempotency_window_in_ms":      newStation.IdempotencyWindow,
			"dls_configuration_poison":      newStation.DlsConfigurationPoison,
			"dls_configuration_schemaverse": newStation.DlsConfigurationSchemaverse,
			"tiered_storage_enabled":        newStation.TieredStorageEnabled,
		})
	} else {
		c.IndentedJSON(200, gin.H{
			"id":                            newStation.ID,
			"name":                          stationName.Ext(),
			"retention_type":                retentionType,
			"retention_value":               body.RetentionValue,
			"storage_type":                  storageTypeForResponse,
			"replicas":                      body.Replicas,
			"created_by_username":           user.Username,
			"created_at":                    time.Now(),
			"last_update":                   time.Now(),
			"is_deleted":                    false,
			"schema_name":                   schemaName,
			"schema_version_number":         schemaVersionNumber,
			"idempotency_window_in_ms":      newStation.IdempotencyWindow,
			"dls_configuration_poison":      newStation.DlsConfigurationPoison,
			"dls_configuration_schemaverse": newStation.DlsConfigurationSchemaverse,
			"tiered_storage_enabled":        newStation.TieredStorageEnabled,
		})
	}
}

func (sh StationsHandler) RemoveStation(c *gin.Context) {
	var body models.RemoveStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	var stationNames []string
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveStation at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	for _, name := range body.StationNames {
		stationName, err := StationNameFromStr(name)
		if err != nil {
			serv.Warnf("RemoveStation: Station %v: %v", name, err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}

		stationNames = append(stationNames, stationName.Ext())

		exist, station, err := db.GetStationByName(stationName.Ext(), user.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]RemoveStation at GetStationByName: Station %v: %v", user.TenantName, user.Username, stationName.external, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			errMsg := fmt.Sprintf("Station %v does not exist", name)
			serv.Warnf("[tenant: %v][user: %v]RemoveStation: %v", user.TenantName, user.Username, errMsg)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
			return
		}

		err = removeStationResources(sh.S, station, true)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]RemoveStation at removeStationResources: Station %v: %v", user.TenantName, user.Username, stationName.external, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	err = db.DeleteStationsByNames(stationNames, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RemoveStation at DeleteStationsByNames: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-remove-station")
	}

	for _, name := range body.StationNames {
		stationName, err := StationNameFromStr(name)
		if err != nil {
			serv.Warnf("RemoveStation at StationNameFromStr: Station %v: %v", name, err.Error())
			continue
		}

		serv.Noticef("[tenant: %v][user: %v]Station %v has been deleted by user %v", user.TenantName, user.Username, stationName.Ext(), user.Username)

		removeStationUpdate := models.SdkClientsUpdates{
			StationName: stationName.Intern(),
			Type:        removeStationUpdateType,
		}
		serv.SendUpdateToClients(removeStationUpdate)
	}
	c.IndentedJSON(200, gin.H{})
}

func (s *Server) removeStationDirect(c *client, reply string, msg []byte) {
	var dsr destroyStationRequest
	tenantName, message, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("removeStationDirect at getTenantNameAndMessage: %v", err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	if err := json.Unmarshal([]byte(message), &dsr); err != nil {
		s.Errorf("[tenant: %v]removeStationDirect at json.Unmarshal: %v", tenantName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	dsr.TenantName = tenantName
	s.removeStationDirectIntern(c, reply, &dsr, true)
}

func (s *Server) removeStationDirectIntern(c *client,
	reply string,
	dsr *destroyStationRequest,
	shouldDeleteStream bool) {
	isNative := shouldDeleteStream
	jsApiResp := JSApiStreamDeleteResponse{ApiResponse: ApiResponse{Type: JSApiStreamDeleteResponseType}}
	memphisGlobalAcc := s.MemphisGlobalAccount()

	// for NATS compatibility
	username, tenantId, err := getUserAndTenantIdFromString(dsr.Username)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]removeStationDirectIntern at getUserAndTenantIdFromString: Station %v: %v", dsr.TenantName, dsr.Username, dsr.StationName, err.Error())
		jsApiResp.Error = NewJSStreamDeleteError(err)
		respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}
	if tenantId != -1 {
		exist, t, err := db.GetTenantById(tenantId)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]removeStationDirectIntern at GetTenantById: Station %v: %v", dsr.TenantName, dsr.Username, dsr.StationName, err.Error())
			jsApiResp.Error = NewJSStreamDeleteError(err)
			respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}
		if !exist {
			msg := fmt.Sprintf("removeStationDirectIntern: Station %v: Tenant with id %v does not exist", dsr.StationName, strconv.Itoa(tenantId))
			serv.Warnf("[tenant: %v][user: %v]: %v", dsr.TenantName, dsr.Username, msg)
			err = errors.New(msg)
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}
		dsr.TenantName = t.Name
		dsr.Username = username
	}

	stationName, err := StationNameFromStr(dsr.StationName)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]removeStationDirectIntern at StationNameFromStr: Station %v: %v", dsr.TenantName, dsr.Username, dsr.StationName, err.Error())
		jsApiResp.Error = NewJSStreamDeleteError(err)
		respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}

	exist, station, err := db.GetStationByName(stationName.Ext(), dsr.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]removeStationDirectIntern at GetStationByName: Station %v: %v", dsr.TenantName, dsr.Username, dsr.StationName, err.Error())
		jsApiResp.Error = NewJSStreamDeleteError(err)
		respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Station %v does not exist", station.Name)
		serv.Warnf("[tenant: %v][user: %v]removeStationDirectIntern: %v", dsr.TenantName, dsr.Username, errMsg)
		err := errors.New(errMsg)
		jsApiResp.Error = NewJSStreamDeleteError(err)
		respondWithErrOrJsApiRespWithEcho(!isNative, c, memphisGlobalAcc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}

	err = removeStationResources(s, station, shouldDeleteStream)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]removeStationDirectIntern at removeStationResources: Station %v: %v", dsr.TenantName, dsr.Username, dsr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	err = db.DeleteStation(station.Name, station.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]removeStationDirectIntern at DeleteStation: Station %v: %v", dsr.TenantName, dsr.Username, dsr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	_, user, err := db.GetUserByUsername(dsr.Username, dsr.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]removeStationDirectIntern at GetUserByUsername: Station %v: %v", dsr.TenantName, dsr.Username, dsr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	message := "Station " + stationName.Ext() + " has been deleted by user " + dsr.Username
	serv.Noticef("[tenant: %v][user: %v] %v ", user.TenantName, user.Username, message)
	if isNative {
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
			serv.Warnf("[tenant: %v][user: %v]removeStationDirectIntern: Station %v - create audit logs error: %v", dsr.TenantName, dsr.Username, stationName.Ext(), err.Error())
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, dsr.Username, analyticsParams, "user-delete-station-sdk")
	}

	respondWithErr(s.MemphisGlobalAccountString(), s, reply, nil)
}

func (sh StationsHandler) GetTotalMessages(tenantName, stationNameExt string) (int, error) {
	stationName, err := StationNameFromStr(stationNameExt)
	if err != nil {
		return 0, err
	}
	totalMessages, err := sh.S.GetTotalMessagesInStation(tenantName, stationName)
	return totalMessages, err
}

func (sh StationsHandler) GetAvgMsgSize(station models.Station) (int64, error) {
	avgMsgSize, err := sh.S.GetAvgMsgSizeInStation(station)
	return avgMsgSize, err
}

func (sh StationsHandler) GetMessages(station models.Station, messagesToFetch int) ([]models.MessageDetails, error) {
	messages, err := sh.S.GetMessages(station, messagesToFetch)
	if err != nil {
		return messages, err
	}

	return messages, nil
}

func (sh StationsHandler) GetLeaderAndFollowers(station models.Station) (string, []string, error) {
	if sh.S.JetStreamIsClustered() {
		leader, followers, err := sh.S.GetLeaderAndFollowers(station)
		if err != nil {
			return "", []string{}, err
		}

		return leader, followers, nil
	} else {
		return "memphis-0", []string{}, nil
	}
}

func getCgStatus(members []models.CgMember) (bool, bool) {
	for _, member := range members {
		if member.IsActive {
			return true, false
		}
	}
	return false, false
}

func (sh StationsHandler) GetPoisonMessageJourney(c *gin.Context) {
	var body models.GetPoisonMessageJourneySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetPoisonMessageJourney at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	poisonMsgsHandler := PoisonMessagesHandler{S: sh.S}
	poisonMessage, err := poisonMsgsHandler.GetDlsMessageDetailsById(body.MessageId, "poison", user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetPoisonMessageJourney at GetDlsMessageDetailsById: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-enter-message-journey")
	}

	c.IndentedJSON(200, poisonMessage)
}

func (sh StationsHandler) DropDlsMessages(c *gin.Context) {
	var body models.DropDlsMessagesSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	err := db.DropDlsMessages(body.DlsMessageIds)
	if err != nil {
		serv.Errorf("DropDlsMessages at db.DropDlsMessages: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-ack-poison-message")
	}

	c.IndentedJSON(200, gin.H{})
}

func (s *Server) ResendUnackedMsgs(dlsMsg models.DlsMessage, user models.User, stationName string) (string, error) {
	size := int64(0)
	for _, cgName := range dlsMsg.PoisonedCgs {
		headersJson := map[string]string{}
		for key, value := range dlsMsg.MessageDetails.Headers {
			headersJson[key] = value
		}
		headersJson["$memphis_pm_id"] = strconv.Itoa(dlsMsg.ID)
		headersJson["$memphis_pm_cg_name"] = cgName

		headers, err := json.Marshal(headersJson)
		if err != nil {
			err = fmt.Errorf("Failed ResendUnackedMsgs at json.Marshal: Poisoned consumer group: %v: %v", cgName, err.Error())
			return cgName, err
		}

		data, err := hex.DecodeString(dlsMsg.MessageDetails.Data)
		if err != nil {
			err = fmt.Errorf("Failed ResendUnackedMsgs at DecodeString: Poisoned consumer group: %v: %v", cgName, err.Error())
			return cgName, err
		}
		err = s.ResendPoisonMessage(user.TenantName, "$memphis_dls_"+replaceDelimiters(stationName)+"_"+replaceDelimiters(cgName), []byte(data), headers)
		if err != nil {
			err = fmt.Errorf("Failed ResendUnackedMsgs at ResendPoisonMessage: Poisoned consumer group: %v: %v", cgName, err.Error())
			return cgName, err
		}
		size += int64(dlsMsg.MessageDetails.Size)
	}
	IncrementEventCounter(user.TenantName, "dls-resend", size, int64(len(dlsMsg.PoisonedCgs)), "", []byte{}, []byte{})
	return "", nil
}

func (sh StationsHandler) ResendPoisonMessages(c *gin.Context) {
	var body models.ResendPoisonMessagesSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("ResendPoisonMessages at getUserDetailsFromMiddleware: At station %v: %v", body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	stationName := strings.ToLower(body.StationName)
	exist, station, err := db.GetStationByName(stationName, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]ResendPoisonMessages at GetStationByName: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !exist {
		errMsg := fmt.Sprintf("Station %v does not exist", stationName)
		serv.Warnf("[tenant: %v][user: %v]ResendPoisonMessages at GetStationByName: %s", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	if len(body.PoisonMessageIds) == 0 {
		sh.S.ResendAll(stationName, station.ID, user.TenantName, user)
	} else {
		stationName := strings.ToLower(body.StationName)
		exist, _, err := db.GetStationByName(stationName, user.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]ResendPoisonMessages at GetStationByName: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		if !exist {
			errMsg := fmt.Sprintf("Station %v does not exist", stationName)
			serv.Warnf("[tenant: %v][user: %v]ResendPoisonMessages at GetStationByName: %s", user.TenantName, user.Username, errMsg)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
			return
		}

		for _, id := range body.PoisonMessageIds {
			_, dlsMsg, err := db.GetDlsMessageById(id)
			if err != nil {
				serv.Errorf("[tenant: %v][user: %v]ResendPoisonMessages at db.GetDlsMessageById: %v", user.TenantName, user.Username, err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			cgName, err := sh.S.ResendUnackedMsgs(dlsMsg, user, stationName)
			if err != nil {
				serv.Errorf("[tenant: %v][user: %v]ResendUnackedMsgs at ResendUnackedMsgs: Poisoned consumer group: %v: %v", user.TenantName, user.Username, cgName, err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}

		}
	}
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-resend-poison-message")
	}

	c.IndentedJSON(200, gin.H{})
}

func (sh StationsHandler) GetMessageDetails(c *gin.Context) {
	var body models.GetMessageDetailsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	msgId := body.MessageId

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetMessageDetails at getUserDetailsFromMiddlewares: At station %v: %v", body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	poisonMsgsHandler := PoisonMessagesHandler{S: sh.S}
	if body.IsDls {
		dlsMessage, err := poisonMsgsHandler.GetDlsMessageDetailsById(body.MessageId, body.DlsType, user.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]GetMessageDetails at GetDlsMessageDetailsById: Message ID: %v :%v", user.TenantName, user.Username, strconv.Itoa(msgId), err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		c.IndentedJSON(200, dlsMessage)
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]GetMessageDetails at StationNameFromStr: Message ID: %v: %v", user.TenantName, user.Username, strconv.Itoa(msgId), err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, station, err := db.GetStationByName(stationName.Ext(), user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetMessageDetails at GetStationByName: Message ID: %v: %v", user.TenantName, user.Username, strconv.Itoa(msgId), err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !exist {
		errMsg := fmt.Sprintf("Station %v does not exist", stationName.external)
		serv.Warnf("[tenant: %v][user: %v]GetMessageDetails: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	sm, err := sh.S.GetMessage(station.TenantName, stationName, uint64(body.MessageSeq))
	if err != nil {
		if IsNatsErr(err, JSNoMessageFoundErr) {
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "The message was not found since it had probably already been deleted"})
			return
		}
		serv.Errorf("[tenant: %v][user: %v]GetMessageDetails at GetMessage: Message ID: %v: %v", user.TenantName, user.Username, strconv.Itoa(msgId), err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	var headersJson map[string]string
	if sm.Header != nil {
		headersJson, err = DecodeHeader(sm.Header)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]GetMessageDetails at DecodeHeader: Message ID: %v: %v", user.TenantName, user.Username, strconv.Itoa(msgId), err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	// For non-native stations - default values
	if !station.IsNative {
		msg := models.MessageResponse{
			MessageSeq: body.MessageSeq,
			Message: models.MessagePayload{
				TimeSent: sm.Time,
				Size:     len(sm.Subject) + len(sm.Data) + len(sm.Header),
				Data:     string(sm.Data),
				Headers:  headersJson,
			},
			Producer: models.ProducerDetailsResp{
				Name:     "",
				IsActive: false,
			},
			PoisonedCgs: []models.PoisonedCg{},
		}
		c.IndentedJSON(200, msg)
		return
	}

	connectionIdHeader := headersJson["$memphis_connectionId"]
	producedByHeader := strings.ToLower(headersJson["$memphis_producedBy"])

	for header := range headersJson {
		if strings.HasPrefix(header, "$memphis") {
			delete(headersJson, header)
		}
	}

	// This check for backward compatability
	if connectionIdHeader == "" || producedByHeader == "" {
		connectionIdHeader = headersJson["connectionId"]
		producedByHeader = strings.ToLower(headersJson["producedBy"])
		if connectionIdHeader == "" || producedByHeader == "" {
			serv.Warnf("[tenant: %v][user: %v]GetMessageDetails: missing mandatory message headers, please upgrade the SDK version you are using", user.TenantName, user.Username)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "missing mandatory message headers, please upgrade the SDK version you are using"})
			return
		}
	}

	connectionId := connectionIdHeader
	poisonedCgs := make([]models.PoisonedCg, 0)
	// Only native stations have CGs
	if station.IsNative {
		poisonedCgs, err = GetPoisonedCgsByMessage(station, int(sm.Sequence))
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]GetMessageDetails at GetPoisonedCgsByMessage: Message ID: %v: %v", user.TenantName, user.Username, strconv.Itoa(msgId), err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		for i, cg := range poisonedCgs {
			cgInfo, err := serv.GetCgInfo(station.TenantName, stationName, cg.CgName)
			if err != nil {
				serv.Errorf("[tenant: %v][user: %v]GetMessageDetails at GetCgInfo: Message ID: %v: %v", user.TenantName, user.Username, strconv.Itoa(msgId), err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			cgMembers, err := GetConsumerGroupMembers(cg.CgName, station)
			if err != nil {
				serv.Errorf("[tenant: %v][user: %v]GetMessageDetails at GetConsumerGroupMembers: Message ID: %v: %v", user.TenantName, user.Username, strconv.Itoa(msgId), err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}

			isActive, isDeleted := getCgStatus(cgMembers)

			poisonedCgs[i].MaxAckTimeMs = cgMembers[0].MaxAckTimeMs
			poisonedCgs[i].MaxMsgDeliveries = cgMembers[0].MaxMsgDeliveries
			poisonedCgs[i].UnprocessedMessages = int(cgInfo.NumPending)
			poisonedCgs[i].InProcessMessages = cgInfo.NumAckPending
			poisonedCgs[i].TotalPoisonMessages = -1
			poisonedCgs[i].IsActive = isActive
			poisonedCgs[i].IsDeleted = isDeleted
		}
		sort.Slice(poisonedCgs, func(i, j int) bool {
			return poisonedCgs[i].CgName < poisonedCgs[j].CgName
		})
	}
	isActive := false
	exist, producer, err := db.GetProducerByStationIDAndConnectionId(producedByHeader, station.ID, connectionId)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetMessageDetails at GetProducerByStationIDAndUsername: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		isActive = producer.IsActive
	}

	msg := models.MessageResponse{
		MessageSeq: body.MessageSeq,
		Message: models.MessagePayload{
			TimeSent: sm.Time,
			Size:     len(sm.Subject) + len(sm.Data) + len(sm.Header),
			Data:     hex.EncodeToString(sm.Data),
			Headers:  headersJson,
		},
		Producer: models.ProducerDetailsResp{
			Name:     producedByHeader,
			IsActive: isActive,
		},
		PoisonedCgs: poisonedCgs,
	}
	c.IndentedJSON(200, msg)
}

func (sh StationsHandler) UseSchema(c *gin.Context) {
	var body models.UseSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("UseSchema at getUserDetailsFromMiddleware: Schema %v: %v", body.SchemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	tenantName := user.TenantName
	schemaName := strings.ToLower(body.SchemaName)
	exist, schema, err := db.GetSchemaByName(schemaName, tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]UseSchema at GetSchemaByName: Schema %v :%v", user.TenantName, user.Username, body.SchemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Schema %v does not exist", schemaName)
		serv.Warnf("[tenant: %v][user: %v]UseSchema: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	schemaVersion, err := getActiveVersionBySchemaId(schema.ID)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]UseSchema at getActiveVersionBySchemaId: Schema %v: %v", user.TenantName, user.Username, body.SchemaName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}
	schemaDetailsResponse := models.StationOverviewSchemaDetails{
		SchemaName:       schemaName,
		VersionNumber:    schemaVersion.VersionNumber,
		UpdatesAvailable: false,
		SchemaType:       schema.Type,
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	for _, stationName := range body.StationNames {
		stationName, err := StationNameFromStr(stationName)
		if err != nil {
			serv.Warnf("[tenant: %v][user: %v]UseSchema at StationNameFromStr: Schema %v at station %v : %v", user.TenantName, user.Username, body.SchemaName, stationName.Ext(), err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}

		exist, station, err := db.GetStationByName(stationName.Ext(), tenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]UseSchema at GetStationByName: Schema %v at station %v : %v", user.TenantName, user.Username, body.SchemaName, stationName.Ext(), err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			errMsg := fmt.Sprintf("Station %v does not exist", station.Name)
			serv.Warnf("[tenant: %v][user: %v]UseSchema at GetStationByName: Schema %v: %v", user.TenantName, user.Username, body.SchemaName, errMsg)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
			return
		}

		err = db.AttachSchemaToStation(stationName.Ext(), schemaName, schemaVersion.VersionNumber, station.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]UseSchema at AttachSchemaToStation: Schema %v at station %v : %v", user.TenantName, user.Username, body.SchemaName, stationName.Ext(), err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
			return
		}

		message := "Schema " + schemaName + " has been attached to station " + stationName.Ext() + " by user " + user.Username
		serv.Noticef("[tenant: %v][user: %v] %v ", user.TenantName, user.Username, message)

		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			StationName:       stationName.Intern(),
			Message:           message,
			CreatedBy:         user.ID,
			CreatedByUsername: user.Username,
			CreatedAt:         time.Now(),
			TenantName:        user.TenantName,
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]UseSchema at CreateAuditLogs: Schema %v at station %v - create audit logs: %v", user.TenantName, user.Username, body.SchemaName, stationName.Ext(), err.Error())
		}

		updateContent, err := generateSchemaUpdateInit(schema)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]UseSchema at generateSchemaUpdateInit: Schema %v at station %v : %v", user.TenantName, user.Username, body.SchemaName, stationName.Ext(), err.Error())
			return
		}
		update := models.ProducerSchemaUpdate{
			UpdateType: models.SchemaUpdateTypeInit,
			Init:       *updateContent,
		}
		sh.S.updateStationProducersOfSchemaChange(station.TenantName, stationName, update)

		if shouldSendAnalytics {
			analyticsParams := map[string]interface{}{"station-name": stationName.Ext(), "schema-name": schemaName}
			analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-attach-schema-to-station")
		}
	}

	c.IndentedJSON(200, schemaDetailsResponse)
}

func (s *Server) useSchemaDirect(c *client, reply string, msg []byte) {
	var asr attachSchemaRequest
	tenantName, attachSchemaMessage, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("useSchemaDirect at getTenantNameAndMessage: %v", err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	if err := json.Unmarshal([]byte(attachSchemaMessage), &asr); err != nil {
		errMsg := fmt.Sprintf("failed attaching schema %v: %v", asr.Name, err.Error())
		s.Errorf("[tenant: %v]useSchemaDirect: At station %v %v", tenantName, asr.StationName, errMsg)
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, errors.New(errMsg))
		return
	}

	asr.TenantName = tenantName
	stationName, err := StationNameFromStr(asr.StationName)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]useSchemaDirect at StationNameFromStr: Schema %v at station %v: %v", asr.TenantName, asr.Username, asr.Name, asr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	exist, station, err := db.GetStationByName(stationName.Ext(), asr.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]useSchemaDirect at GetStationByName: Schema %v at station %v: %v", asr.TenantName, asr.Username, asr.Name, asr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	if !exist {
		errMsg := fmt.Sprintf("Station %v does not exist", stationName.external)
		serv.Warnf("[tenant: %v][user: %v]useSchemaDirect: %v", asr.TenantName, asr.Username, errMsg)
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, errors.New("memphis: "+errMsg))
		return
	}
	schemaName := strings.ToLower(asr.Name)
	exist, schema, err := db.GetSchemaByName(schemaName, station.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]useSchemaDirect at GetSchemaByName: Schema %v at station %v: %v", asr.TenantName, asr.Username, asr.Name, asr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Schema %v does not exist", schemaName)
		serv.Warnf("[tenant: %v][user: %v]useSchemaDirect: %v", asr.TenantName, asr.Username, errMsg)
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, errors.New(errMsg))
		return
	}

	schemaVersion, err := getActiveVersionBySchemaId(schema.ID)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]useSchemaDirect at getActiveVersionBySchemaId: Schema %v at station %v: %v", asr.TenantName, asr.Username, asr.Name, asr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	err = db.AttachSchemaToStation(stationName.Ext(), schemaName, schemaVersion.VersionNumber, station.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]useSchemaDirect at db.AttachSchemaToStation: Schema %v at station %v: %v", asr.TenantName, asr.Username, asr.Name, asr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	message := fmt.Sprintf("Schema %v has been attached to station %v by user %v", schemaName, stationName.Ext(), asr.Username)
	serv.Noticef("[tenant: %v][user: %v]: %v", asr.TenantName, asr.Username, message)
	_, user, err := db.GetUserByUsername(asr.Username, asr.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]useSchemaDirect at GetUserByUsername: Schema %v at station %v: %v", asr.TenantName, asr.Username, asr.Name, asr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		StationName:       stationName.Intern(),
		Message:           message,
		CreatedBy:         user.ID,
		CreatedByUsername: user.Username,
		CreatedAt:         time.Now(),
		TenantName:        user.TenantName,
	}
	auditLogs = append(auditLogs, newAuditLog)
	err = CreateAuditLogs(auditLogs)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]useSchemaDirect : Schema %v at station %v - create audit logs %v", asr.TenantName, asr.Username, asr.Name, asr.StationName, err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := map[string]interface{}{"station-name": stationName.Ext(), "schema-name": schemaName}
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-attach-schema-to-station")
	}

	updateContent, err := generateSchemaUpdateInit(schema)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]useSchemaDirect at generateSchemaUpdateInit: Schema %v at station %v: %v", asr.TenantName, asr.Username, asr.Name, asr.StationName, err.Error())
		return
	}

	update := models.ProducerSchemaUpdate{
		UpdateType: models.SchemaUpdateTypeInit,
		Init:       *updateContent,
	}

	serv.updateStationProducersOfSchemaChange(station.TenantName, stationName, update)
	respondWithErr(s.MemphisGlobalAccountString(), s, reply, nil)
}

func removeSchemaFromStation(s *Server, sn StationName, updateDB bool, tenantName string) error {
	exist, station, err := db.GetStationByName(sn.Ext(), tenantName)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("station %v does not exist", sn.external)
	}

	if updateDB {
		err = db.DetachSchemaFromStation(sn.Ext(), station.TenantName)
		if err != nil {
			return err
		}
	}

	update := models.ProducerSchemaUpdate{
		UpdateType: models.SchemaUpdateTypeDrop,
	}

	s.updateStationProducersOfSchemaChange(station.TenantName, sn, update)
	return nil
}

func (s *Server) removeSchemaFromStationDirect(c *client, reply string, msg []byte) {
	var dsr detachSchemaRequest
	tenantName, message, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("removeSchemaFromStationDirect at getTenantNameAndMessage: %v", err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	if err := json.Unmarshal([]byte(message), &dsr); err != nil {
		s.Errorf("[tenant: %v]removeSchemaFromStationDirect at json.Unmarshal: failed removing schema at station %v: %v", tenantName, dsr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	dsr.TenantName = tenantName
	stationName, err := StationNameFromStr(dsr.StationName)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]removeSchemaFromStationDirec at StationNameFromStrt: At station %v: %v", dsr.TenantName, dsr.Username, dsr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	err = removeSchemaFromStation(serv, stationName, true, dsr.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]removeSchemaFromStationDirect at removeSchemaFromStation: At station %v: %v", dsr.TenantName, dsr.Username, dsr.StationName, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(tenantName, dsr.Username, analyticsParams, "user-detach-schema-from-station-sdk")
	}

	respondWithErr(s.MemphisGlobalAccountString(), s, reply, nil)
}

func (sh StationsHandler) RemoveSchemaFromStation(c *gin.Context) {
	var body models.RemoveSchemaFromStation
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("RemoveSchemaFromStation at StationNameFromStr: At station %v: %v", body.StationName, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveSchemaFromStation at getUserDetailsFromMiddleware: At station %v: %v", body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	exist, station, err := db.GetStationByName(stationName.Ext(), user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RemoveSchemaFromStation at GetStationByName: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Station %v does not exist", body.StationName)
		serv.Warnf("[tenant: %v][user: %v]RemoveSchemaFromStation at GetStationByName: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	tenantName := strings.ToLower(station.TenantName)

	err = removeSchemaFromStation(sh.S, stationName, true, tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RemoveSchemaFromStation: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	message := fmt.Sprintf("Schema %v has been deleted from station %v by user %v", station.SchemaName, stationName.Ext(), user.Username)
	serv.Noticef("[tenant: %v][user: %v]: %v", user.TenantName, user.Username, message)
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		StationName:       stationName.Intern(),
		Message:           message,
		CreatedBy:         user.ID,
		CreatedByUsername: user.Username,
		CreatedAt:         time.Now(),
		TenantName:        user.TenantName,
	}
	auditLogs = append(auditLogs, newAuditLog)
	err = CreateAuditLogs(auditLogs)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RemoveSchemaFromStation: At station %v - create audit logs error: %v", user.TenantName, user.Username, body.StationName, err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-remove-schema-from-station")
	}

	c.IndentedJSON(200, gin.H{})
}

func (sh StationsHandler) GetUpdatesForSchemaByStation(c *gin.Context) {
	var body models.GetUpdatesForSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetUpdatesForSchemaByStation at getUserDetailsFromMiddleware: At station %v: %v", body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]GetUpdatesForSchemaByStation at StationNameFromStr: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, station, err := db.GetStationByName(stationName.Ext(), user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetUpdatesForSchemaByStation at GetStationByName: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Station %v does not exist", body.StationName)
		serv.Warnf("[tenant: %v][user: %v]GetUpdatesForSchemaByStation: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	exist, schema, err := db.GetSchemaByName(station.SchemaName, station.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetUpdatesForSchemaByStation at GetSchemaByName: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !exist {
		errMsg := fmt.Sprintf("Schema %v does not exist", station.SchemaName)
		serv.Warnf("[tenant: %v][user: %v]GetUpdatesForSchemaByStation at GetSchemaByName: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	schemasHandler := SchemasHandler{S: sh.S}
	extedndedSchemaDetails, err := schemasHandler.getExtendedSchemaDetailsUpdateAvailable(station.SchemaVersionNumber, schema, user.TenantName)

	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetUpdatesForSchemaByStation at getExtendedSchemaDetailsUpdateAvailable: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-apply-schema-updates-on-station")
	}

	c.IndentedJSON(200, extedndedSchemaDetails)
}

func (sh StationsHandler) TierdStorageClicked(c *gin.Context) {
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-pushed-tierd-storage-button")
	}

	c.IndentedJSON(200, gin.H{})
}

func (sh StationsHandler) UpdateDlsConfig(c *gin.Context) {
	var body models.UpdateDlsConfigSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("UpdateDlsConfig at getUserDetailsFromMiddleware: At station %v: %v", body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]UpdateDlsConfig at StationNameFromStr: At station, %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, station, err := db.GetStationByName(stationName.Ext(), user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]UpdateDlsConfig at GetStationByName: At station, %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Station %v does not exist", body.StationName)
		serv.Warnf("[tenant: %v][user: %v]UpdateDlsConfig: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	poisonConfigChanged := station.DlsConfigurationPoison != body.Poison
	schemaverseConfigChanged := station.DlsConfigurationSchemaverse != body.Schemaverse
	if poisonConfigChanged || schemaverseConfigChanged {
		err = db.UpdateStationDlsConfig(station.Name, body.Poison, body.Schemaverse, station.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]UpdateDlsConfig at db.UpdateStationDlsConfig: At station, %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}
	configUpdate := models.SdkClientsUpdates{
		StationName: stationName.Intern(),
		Type:        schemaToDlsUpdateType,
		Update:      station.DlsConfigurationSchemaverse,
	}
	serv.SendUpdateToClients(configUpdate)

	c.IndentedJSON(200, gin.H{"poison": body.Poison, "schemaverse": body.Schemaverse})
}

func (sh StationsHandler) PurgeStation(c *gin.Context) {
	var body models.PurgeStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("PurgeStation: station name at StationNameFromStr: %v: %v", body.StationName, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("PurgeStation at getUserDetailsFromMiddleware: At station %v: %v", body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	exist, station, err := db.GetStationByName(stationName.Ext(), user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]PurgeStation at GetStationByName: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Station %v does not exist", stationName.external)
		serv.Warnf("[tenant: %v][user: %v]PurgeStation: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	if body.PurgeStation {
		err = sh.S.PurgeStream(station.TenantName, stationName.Intern())
		if err != nil && !IsNatsErr(err, JSStreamNotFoundErr) {
			serv.Errorf("[tenant: %v][user: %v]PurgeStation: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	if body.PurgeDls {
		err := db.PurgeDlsMsgsFromStation(station.ID)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]PurgeStation dls at PurgeDlsMsgsFromStation: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-purge-station")
	}
	c.IndentedJSON(200, gin.H{})
}

func (sh StationsHandler) RemoveMessages(c *gin.Context) {
	var body models.RemoveMessagesSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveMessages at getUserDetailsFromMiddleware: At station %v: %v", body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]RemoveMessages at StationNameFromStr: station name: %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, station, err := db.GetStationByName(stationName.Ext(), user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RemoveMessages at GetStationByName: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Station %v does not exist", stationName.external)
		serv.Warnf("[tenant: %v][user: %v]RemoveMessages at GetStationByName: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	for _, msg := range body.MessageSeqs {
		err = sh.S.RemoveMsg(station.TenantName, stationName, msg)
		if err != nil {
			if IsNatsErr(err, JSStreamNotFoundErr) || IsNatsErr(err, JSStreamMsgDeleteFailedF) {
				continue
			}
			serv.Errorf("[tenant: %v][user: %v]RemoveMessages at RemoveMsg: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-remove-messages")
	}

	c.IndentedJSON(200, gin.H{})
}

func getUserAndTenantIdFromString(username string) (string, int, error) {
	re := regexp.MustCompile(`^(.*)(\$\d+)$`)
	matches := re.FindStringSubmatch(username)
	if len(matches) == 3 {
		beforeSuffix := matches[1]
		numberAfterSuffix := strings.TrimLeft(matches[2], userNameItemSep)
		tenantId, err := strconv.Atoi(numberAfterSuffix)
		if err != nil {
			return "", 0, err
		}
		return beforeSuffix, tenantId, nil
	}
	return username, -1, nil

}

func (s *Server) RemoveOldStations() {
	stations, err := db.GetDeletedStations()
	if err != nil {
		s.Errorf("RemoveOldStations: at GetDeletedStations: %v", err.Error())
		return
	}
	for _, station := range stations {
		err = removeStationResources(s, station, true)
		if err != nil {
			s.Errorf("[tenant: %v]RemoveOldStations: at removeStationResources: %v", station.TenantName, err.Error())
			return
		}
	}
	err = db.RemoveDeletedStations()
	if err != nil {
		s.Warnf("RemoveOldStations: at RemoveDeletedStations: %v", err.Error())
		return
	}
}
