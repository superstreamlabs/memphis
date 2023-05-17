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
	"memphis/analytics"
	"memphis/db"
	"memphis/models"
	"memphis/utils"
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

func validateReplicas(replicas int) error {
	if replicas > 5 {
		return errors.New("max replicas in a cluster is 5")
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
		err = s.RemoveStream(stationName.Intern())
		if err != nil && !IsNatsErr(err, JSStreamNotFoundErr) {
			return err
		}
	}

	DeleteTagsFromStation(station.ID)

	err = db.DeleteProducersByStationID(station.ID)
	if err != nil {
		return err
	}

	err = db.DeleteConsumersByStationID(station.ID)
	if err != nil {
		return err
	}

	err = RemoveAllAuditLogsByStation(station.Name)
	if err != nil {
		serv.Errorf("removeStationResources: Station " + station.Name + ": " + err.Error())
	}

	return nil
}

func (s *Server) createStationDirect(c *client, reply string, msg []byte) {
	var csr createStationRequest
	if err := json.Unmarshal(msg, &csr); err != nil {
		s.Errorf("createStationDirect: failed creating station: %v", err.Error())
		respondWithErr(s, reply, err)
		return
	}
	s.createStationDirectIntern(c, reply, &csr, true)
}

func (s *Server) createStationDirectIntern(c *client,
	reply string,
	csr *createStationRequest,
	shouldCreateStream bool) {
	isNative := shouldCreateStream
	jsApiResp := JSApiStreamCreateResponse{ApiResponse: ApiResponse{Type: JSApiStreamCreateResponseType}}
	stationName, err := StationNameFromStr(csr.StationName)
	if err != nil {
		serv.Warnf("createStationDirect: Station " + csr.StationName + ": " + err.Error())
		jsApiResp.Error = NewJSStreamCreateError(err)
		respondWithErrOrJsApiResp(!isNative, c, c.acc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}

	exist, _, err := db.GetStationByName(stationName.Ext())
	if err != nil {
		serv.Errorf("createStationDirect: Station " + csr.StationName + ": " + err.Error())
		jsApiResp.Error = NewJSStreamCreateError(err)
		respondWithErrOrJsApiResp(!isNative, c, c.acc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}

	if exist {
		errMsg := "Station " + stationName.Ext() + " already exists"
		serv.Warnf("createStationDirect: " + errMsg)
		jsApiResp.Error = NewJSStreamNameExistError()
		respondWithErrOrJsApiResp(!isNative, c, c.acc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}

	schemaName := csr.SchemaName
	var schemaDetails models.SchemaDetails
	if schemaName != "" {
		schemaName = strings.ToLower(csr.SchemaName)
		exist, schema, err := db.GetSchemaByName(schemaName)
		if err != nil {
			serv.Errorf("createStationDirect: Station " + csr.StationName + ": " + err.Error())
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiResp(!isNative, c, c.acc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}
		if !exist {
			errMsg := "Schema " + csr.SchemaName + " does not exist"
			serv.Warnf("createStationDirect: " + errMsg)
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiResp(!isNative, c, c.acc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}

		schemaVersion, err := getActiveVersionBySchemaId(schema.ID)
		if err != nil {
			serv.Errorf("createStationDirect: Station " + csr.StationName + ": " + err.Error())
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiResp(!isNative, c, c.acc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
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
			serv.Warnf("createStationDirect: " + err.Error())
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiResp(!isNative, c, c.acc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}
		retentionValue = csr.RetentionValue
	} else {
		retentionType = "message_age_sec"
		retentionValue = 604800 // 1 week
	}

	var storageType string
	if csr.StorageType != "" {
		storageType = strings.ToLower(csr.StorageType)
		err = validateStorageType(storageType)
		if err != nil {
			serv.Warnf("createStationDirect: " + err.Error())
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiResp(!isNative, c, c.acc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}
	} else {
		storageType = "file"
	}

	replicas := csr.Replicas
	if replicas > 0 {
		err = validateReplicas(replicas)
		if err != nil {
			serv.Warnf("createStationDirect: " + err.Error())
			jsApiResp.Error = NewJSStreamCreateError(err)
			respondWithErrOrJsApiResp(!isNative, c, c.acc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
			return
		}
	} else {
		replicas = 1
	}

	if csr.IdempotencyWindow <= 0 {
		csr.IdempotencyWindow = 120000 // default
	} else if csr.IdempotencyWindow < 100 {
		csr.IdempotencyWindow = 100 // minimum is 100 millis
	}

	err = validateIdempotencyWindow(retentionType, retentionValue, csr.IdempotencyWindow)
	if err != nil {
		serv.Warnf("createStationDirect: " + err.Error())
		jsApiResp.Error = NewJSStreamCreateError(err)
		respondWithErrOrJsApiResp(!isNative, c, c.acc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}

	username := c.memphisInfo.username
	if username == "" {
		username = csr.Username
	}

	if shouldCreateStream {
		err = s.CreateStream(stationName, retentionType, retentionValue, storageType, csr.IdempotencyWindow, replicas, csr.TieredStorageEnabled)
		if err != nil {
			if IsNatsErr(err, JSStreamReplicasNotSupportedErr) {
				serv.Warnf("CreateStationDirect: Station " + stationName.Ext() + ": Station can not be created, probably since replicas count is larger than the cluster size")
				respondWithErr(s, reply, errors.New("station can not be created, probably since replicas count is larger than the cluster size"))
				return
			}

			serv.Errorf("createStationDirect: Station " + csr.StationName + ": " + err.Error())
			respondWithErr(s, reply, err)
			return
		}
	}

	_, user, err := db.GetUserByUsername(username)
	if err != nil {
		serv.Warnf("createStationDirect: " + err.Error())
		respondWithErr(s, reply, err)
	}
	_, rowsUpdated, err := db.InsertNewStation(stationName.Ext(), user.ID, user.Username, retentionType, retentionValue, storageType, replicas, schemaDetails.SchemaName, schemaDetails.VersionNumber, csr.IdempotencyWindow, isNative, csr.DlsConfiguration, csr.TieredStorageEnabled)
	if err != nil {
		if !strings.Contains(err.Error(), "already exist") {
			serv.Errorf("createStationDirect: Station " + csr.StationName + ": " + err.Error())
		}
		respondWithErr(s, reply, err)
		return
	}
	if rowsUpdated > 0 {
		message := "Station " + stationName.Ext() + " has been created by user " + username
		serv.Noticef(message)
		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			StationName:       stationName.Ext(),
			Message:           message,
			CreatedBy:         user.ID,
			CreatedByUsername: user.Username,
			CreatedAt:         time.Now(),
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			serv.Errorf("createStationDirect: Station " + csr.StationName + " - create audit logs error: " + err.Error())
		}

		shouldSendAnalytics, _ := shouldSendAnalytics()
		if shouldSendAnalytics {
			param1 := analytics.EventParam{
				Name:  "station-name",
				Value: stationName.Ext(),
			}
			param2 := analytics.EventParam{
				Name:  "tiered-storage",
				Value: strconv.FormatBool(csr.TieredStorageEnabled),
			}
			param3 := analytics.EventParam{
				Name:  "nats-comp",
				Value: strconv.FormatBool(!isNative),
			}
			analyticsParams := []analytics.EventParam{param1, param2, param3}
			analytics.SendEventWithParams(username, analyticsParams, "user-create-station-sdk")
		}
	}

	respondWithErr(s, reply, nil)
}

func (sh StationsHandler) GetStation(c *gin.Context) {
	var body models.GetStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	tagsHandler := TagsHandler{S: sh.S}
	stationName := strings.ToLower(body.StationName)
	exist, station, err := db.GetStationByName(stationName)
	if err != nil {
		serv.Errorf("GetStation: Station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	} else if !exist {
		errMsg := "Station " + body.StationName + " does not exist"
		serv.Warnf("GetStation: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}
	tags, err := tagsHandler.GetTagsByEntityWithID("station", station.ID)
	if err != nil {
		serv.Errorf("GetStation: Station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if station.StorageType == "file" {
		station.StorageType = "disk"
	}

	_, ok = IntegrationsCache["s3"].(models.Integration)
	if !ok {
		station.TieredStorageEnabled = false
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

func (sh StationsHandler) GetStationsDetails() ([]models.ExtendedStationDetails, error) {
	var exStations []models.ExtendedStationDetails
	stations, err := db.GetActiveStations()
	if err != nil {
		return []models.ExtendedStationDetails{}, err
	}
	stationTotalMsgs := make(map[string]int)
	if len(stations) == 0 {
		return []models.ExtendedStationDetails{}, nil
	} else {
		allStreamInfo, err := serv.memphisAllStreamsInfo()
		if err != nil {
			return []models.ExtendedStationDetails{}, err
		}
		for _, info := range allStreamInfo {
			streamName := info.Config.Name
			if !strings.Contains(streamName, "$memphis") {
				stationTotalMsgs[streamName] = int(info.State.Msgs)
			}
		}
		stationIdsDlsMsgs, err := db.GetStationIdsFromDlsMsgs()
		if err != nil {
			return []models.ExtendedStationDetails{}, err
		}
		tagsHandler := TagsHandler{S: sh.S}
		for _, station := range stations {
			hasDlsMsgs := false
			for _, stationId := range stationIdsDlsMsgs {
				if stationId == station.ID {
					hasDlsMsgs = true
				}
			}

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
			_, ok := IntegrationsCache["s3"].(models.Integration)
			if !ok {
				station.TieredStorageEnabled = false
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

func (sh StationsHandler) GetAllStationsDetails(shouldGetTags bool) ([]models.ExtendedStation, uint64, uint64, error) {
	totalMessages := uint64(0)
	totalDlsMessages, err := db.GetTotalDlsMessages()
	if err != nil {
		return []models.ExtendedStation{}, totalMessages, totalDlsMessages, err
	}

	stations, err := db.GetAllStationsDetails()
	if err != nil {
		return stations, totalMessages, totalDlsMessages, err
	}
	if len(stations) == 0 {
		return []models.ExtendedStation{}, totalMessages, totalDlsMessages, nil
	} else {
		stationTotalMsgs := make(map[string]int)
		tagsHandler := TagsHandler{S: sh.S}
		allStreamInfo, err := serv.memphisAllStreamsInfo()
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

		stationIdsDlsMsgs, err := db.GetStationIdsFromDlsMsgs()
		if err != nil {
			return []models.ExtendedStation{}, totalMessages, totalDlsMessages, err
		}

		var extStations []models.ExtendedStation
		for i := 0; i < len(stations); i++ {
			fullStationName, err := StationNameFromStr(stations[i].Name)
			if err != nil {
				return []models.ExtendedStation{}, totalMessages, totalDlsMessages, err
			}
			hasDlsMsgs := false
			for _, stationId := range stationIdsDlsMsgs {
				if stationId == stations[i].ID {
					hasDlsMsgs = true
				}
			}

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

			_, ok := IntegrationsCache["s3"].(models.Integration)
			if !ok {
				stations[i].TieredStorageEnabled = false
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

func (sh StationsHandler) GetStations(c *gin.Context) {
	stations, err := sh.GetStationsDetails()
	if err != nil {
		serv.Errorf("GetStations: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-enter-stations-page")
	}

	c.IndentedJSON(200, gin.H{
		"stations": stations,
	})
}

func (sh StationsHandler) GetAllStations(c *gin.Context) {
	stations, _, _, err := sh.GetAllStationsDetails(true)
	if err != nil {
		serv.Errorf("GetAllStations: " + err.Error())
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

	stationName, err := StationNameFromStr(body.Name)
	if err != nil {
		serv.Warnf("CreateStation: Station " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, _, err := db.GetStationByName(stationName.Ext())
	if err != nil {
		serv.Errorf("CreateStation: Station " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		errMsg := "Station " + stationName.external + " already exists"
		serv.Warnf("CreateStation: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("CreateStation: Station " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}

	var schemaVersionNumber int
	schemaName := body.SchemaName
	if schemaName != "" {
		schemaName = strings.ToLower(body.SchemaName)
		exist, schema, err := db.GetSchemaByName(schemaName)
		if err != nil {
			serv.Errorf("CreateStation: Station " + body.Name + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
			return
		}
		if !exist {
			errMsg := "Schema " + schemaName + " does not exist"
			serv.Warnf("CreateStation: Station " + body.Name + ": " + errMsg)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
			return
		}

		schemaVersion, err := getActiveVersionBySchemaId(schema.ID)
		if err != nil {
			serv.Errorf("CreateStation: Station " + body.Name + ": " + err.Error())
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
			serv.Warnf("CreateStation: Station " + body.Name + ": " + err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	} else {
		retentionType = "message_age_sec"
		body.RetentionValue = 604800 // 1 week
	}

	if body.StorageType != "" {
		body.StorageType = strings.ToLower(body.StorageType)
		err = validateStorageType(body.StorageType)
		if err != nil {
			serv.Warnf("CreateStation: Station " + body.Name + ": " + err.Error())
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

	if body.Replicas > 0 {
		err = validateReplicas(body.Replicas)
		if err != nil {
			serv.Warnf("CreateStation: Station " + body.Name + ": " + err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
	} else {
		body.Replicas = 1
	}

	err = validateIdempotencyWindow(body.RetentionType, body.RetentionValue, body.IdempotencyWindow)
	if err != nil {
		serv.Warnf("CreateStation: Station " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	if body.IdempotencyWindow <= 0 {
		body.IdempotencyWindow = 120000 // default
	} else if body.IdempotencyWindow < 100 {
		body.IdempotencyWindow = 100 // minimum is 100 millis
	}

	newStation, rowsUpdated, err := db.InsertNewStation(stationName.Ext(), user.ID, user.Username, retentionType, body.RetentionValue, body.StorageType, body.Replicas, schemaName, schemaVersionNumber, body.IdempotencyWindow, true, body.DlsConfiguration, body.TieredStorageEnabled)

	if err != nil {
		serv.Errorf("CreateStation: Station " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	//rowsUpdated == 0 means that the row already exists
	if rowsUpdated == 0 {
		errMsg := "Station " + newStation.Name + " already exists"
		serv.Warnf("CreateStation: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	err = sh.S.CreateStream(stationName, retentionType, body.RetentionValue, body.StorageType, body.IdempotencyWindow, body.Replicas, body.TieredStorageEnabled)
	if err != nil {
		if IsNatsErr(err, JSInsufficientResourcesErr) {
			serv.Warnf("CreateStation: Station " + body.Name + ": Station can not be created, probably since replicas count is larger than the cluster size")
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station can not be created, probably since replicas count is larger than the cluster size"})
			return
		}

		serv.Errorf("CreateStation: Station " + body.Name + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(body.Tags) > 0 {
		err = AddTagsToEntity(body.Tags, "station", newStation.ID)
		if err != nil {
			serv.Errorf("CreateStation: : Station " + body.Name + " Failed adding tags: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}
	message := "Station " + stationName.Ext() + " has been created by " + user.Username
	serv.Noticef(message)
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		StationName:       stationName.Ext(),
		Message:           message,
		CreatedBy:         user.ID,
		CreatedByUsername: user.Username,
		CreatedAt:         time.Now(),
	}
	auditLogs = append(auditLogs, newAuditLog)
	err = CreateAuditLogs(auditLogs)
	if err != nil {
		serv.Errorf("CreateStation: Station " + body.Name + ": " + err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		param1 := analytics.EventParam{
			Name:  "station-name",
			Value: stationName.Ext(),
		}
		param2 := analytics.EventParam{
			Name:  "tiered-storage",
			Value: strconv.FormatBool(newStation.TieredStorageEnabled),
		}
		analyticsParams := []analytics.EventParam{param1, param2}
		analytics.SendEventWithParams(user.Username, analyticsParams, "user-create-station")
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
	// if err := DenyForSandboxEnv(c); err != nil {
	//  return
	// }
	var body models.RemoveStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	var stationNames []string
	for _, name := range body.StationNames {
		stationName, err := StationNameFromStr(name)
		if err != nil {
			serv.Warnf("RemoveStation: Station " + name + ": " + err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}

		stationNames = append(stationNames, stationName.Ext())

		exist, station, err := db.GetStationByName(stationName.Ext())
		if err != nil {
			serv.Errorf("RemoveStation: Station " + stationName.external + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			errMsg := "Station " + name + " does not exist"
			serv.Warnf("RemoveStation: " + errMsg)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
			return
		}

		err = removeStationResources(sh.S, station, true)
		if err != nil {
			serv.Errorf("RemoveStation: Station " + stationName.external + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	err := db.DeleteStationsByNames(stationNames)
	if err != nil {
		serv.Errorf("RemoveStation: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-remove-station")
	}

	for _, name := range body.StationNames {
		stationName, err := StationNameFromStr(name)
		if err != nil {
			serv.Warnf("RemoveStation: Station " + name + ": " + err.Error())
			continue
		}

		user, err := getUserDetailsFromMiddleware(c)
		if err != nil {
			serv.Errorf("RemoveStation: Station " + name + ": " + err.Error())
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		}

		serv.Noticef("Station " + stationName.Ext() + " has been deleted by user " + user.Username)

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
	if err := json.Unmarshal(msg, &dsr); err != nil {
		s.Errorf("removeStationDirect: " + err.Error())
		respondWithErr(s, reply, err)
		return
	}
	s.removeStationDirectIntern(c, reply, &dsr, true)
}

func (s *Server) removeStationDirectIntern(c *client,
	reply string,
	dsr *destroyStationRequest,
	shouldDeleteStream bool) {
	isNative := shouldDeleteStream
	jsApiResp := JSApiStreamDeleteResponse{ApiResponse: ApiResponse{Type: JSApiStreamDeleteResponseType}}

	stationName, err := StationNameFromStr(dsr.StationName)
	if err != nil {
		serv.Warnf("removeStationDirect: Station " + dsr.StationName + ": " + err.Error())
		jsApiResp.Error = NewJSStreamDeleteError(err)
		respondWithErrOrJsApiResp(!isNative, c, c.acc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}

	exist, station, err := db.GetStationByName(stationName.Ext())
	if err != nil {
		serv.Errorf("removeStationDirect: Station " + dsr.StationName + ": " + err.Error())
		jsApiResp.Error = NewJSStreamDeleteError(err)
		respondWithErrOrJsApiResp(!isNative, c, c.acc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}
	if !exist {
		errMsg := "Station " + station.Name + " does not exist"
		serv.Warnf("removeStationDirect: " + errMsg)
		err := errors.New(errMsg)
		jsApiResp.Error = NewJSStreamDeleteError(err)
		respondWithErrOrJsApiResp(!isNative, c, c.acc, _EMPTY_, reply, _EMPTY_, jsApiResp, err)
		return
	}

	err = removeStationResources(s, station, shouldDeleteStream)
	if err != nil {
		serv.Errorf("RemoveStation: Station " + dsr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	err = db.DeleteStation(stationName.Ext())
	if err != nil {
		serv.Errorf("RemoveStation error: Station " + dsr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}
	_, user, err := db.GetUserByUsername(dsr.Username)
	if err != nil {
		serv.Errorf("RemoveStation error: Station " + dsr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}
	message := "Station " + stationName.Ext() + " has been deleted by user " + dsr.Username
	serv.Noticef(message)
	if isNative {
		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			StationName:       stationName.Ext(),
			Message:           message,
			CreatedBy:         user.ID,
			CreatedByUsername: user.Username,
			CreatedAt:         time.Now(),
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			serv.Warnf("removeStationDirect: Station " + stationName.Ext() + " - create audit logs error: " + err.Error())
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(dsr.Username, "user-delete-station-sdk")
	}

	respondWithErr(s, reply, nil)
}

func (sh StationsHandler) GetTotalMessages(stationNameExt string) (int, error) {
	stationName, err := StationNameFromStr(stationNameExt)
	if err != nil {
		return 0, err
	}
	totalMessages, err := sh.S.GetTotalMessagesInStation(stationName)
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
	deletedCount := 0
	for _, member := range members {
		if member.IsActive {
			return true, false
		}

		if member.IsDeleted {
			deletedCount++
		}
	}

	if len(members) == deletedCount {
		return false, true
	}

	return false, false
}

func (sh StationsHandler) GetPoisonMessageJourney(c *gin.Context) {
	var body models.GetPoisonMessageJourneySchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	poisonMsgsHandler := PoisonMessagesHandler{S: sh.S}
	poisonMessage, err := poisonMsgsHandler.GetDlsMessageDetailsById(body.MessageId, "poison")
	if err != nil {
		serv.Errorf("GetPoisonMessageJourney: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-enter-message-journey")
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
		serv.Errorf("DropDlsMessages: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-ack-poison-message")
	}

	c.IndentedJSON(200, gin.H{})
}

func (sh StationsHandler) ResendPoisonMessages(c *gin.Context) {
	var body models.ResendPoisonMessagesSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName := strings.ToLower(body.StationName)
	exist, _, err := db.GetStationByName(stationName)
	if err != nil {
		serv.Errorf("ResendPoisonMessages: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !exist {
		errMsg := "Station " + stationName + " does not exist"
		serv.Warnf("ResendPoisonMessages: %s", errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	for _, id := range body.PoisonMessageIds {
		_, dlsMsg, err := db.GetDlsMessageById(id)
		if err != nil {
			serv.Errorf("ResendPoisonMessages: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		for _, cgName := range dlsMsg.PoisonedCgs {
			headersJson := map[string]string{}
			for key, value := range dlsMsg.MessageDetails.Headers {
				headersJson[key] = value
			}
			headersJson["$memphis_pm_id"] = strconv.Itoa(dlsMsg.ID)
			headersJson["$memphis_pm_cg_name"] = cgName

			headers, err := json.Marshal(headersJson)
			if err != nil {
				serv.Errorf("ResendPoisonMessages: Poisoned consumer group: " + cgName + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}

			data, err := hex.DecodeString(dlsMsg.MessageDetails.Data)
			if err != nil {
				serv.Errorf("ResendPoisonMessages: Poisoned consumer group: " + cgName + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			err = sh.S.ResendPoisonMessage("$memphis_dls_"+replaceDelimiters(stationName)+"_"+replaceDelimiters(cgName), []byte(data), headers)
			if err != nil {
				serv.Errorf("ResendPoisonMessages: Poisoned consumer group: " + cgName + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
		}

	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-resend-poison-message")
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

	poisonMsgsHandler := PoisonMessagesHandler{S: sh.S}
	if body.IsDls {
		dlsMessage, err := poisonMsgsHandler.GetDlsMessageDetailsById(body.MessageId, body.DlsType)
		if err != nil {
			serv.Errorf("GetMessageDetails: Message ID: " + string(rune(msgId)) + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		c.IndentedJSON(200, dlsMessage)
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("GetMessageDetails: Message ID: " + string(rune(msgId)) + ": " + err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, station, err := db.GetStationByName(stationName.Ext())
	if err != nil {
		serv.Errorf("GetMessageDetails: Message ID: " + string(rune(msgId)) + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !exist {
		errMsg := "Station " + stationName.external + " does not exist"
		serv.Warnf("GetMessageDetails: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	sm, err := sh.S.GetMessage(stationName, uint64(body.MessageSeq))
	if err != nil {
		if IsNatsErr(err, JSNoMessageFoundErr) {
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "The message was not found since it had probably already been deleted"})
			return
		}
		serv.Errorf("GetMessageDetails: Message ID: Message ID: " + string(rune(msgId)) + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	var headersJson map[string]string
	if sm.Header != nil {
		headersJson, err = DecodeHeader(sm.Header)
		if err != nil {
			serv.Errorf("GetMessageDetails: Message ID: " + string(rune(msgId)) + ": " + err.Error())
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
			Producer: models.ProducerDetails{
				Name:          "",
				ClientAddress: "",
				IsActive:      false,
				IsDeleted:     false,
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
			serv.Warnf("GetMessageDetails: missing mandatory message headers, please upgrade the SDK version you are using")
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
			serv.Errorf("GetMessageDetails: Message ID: " + string(rune(msgId)) + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		for i, cg := range poisonedCgs {
			cgInfo, err := sh.S.GetCgInfo(stationName, cg.CgName)
			if err != nil {
				serv.Errorf("GetMessageDetails: Message ID: " + string(rune(msgId)) + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}

			cgMembers, err := GetConsumerGroupMembers(cg.CgName, station)
			if err != nil {
				serv.Errorf("GetMessageDetails: Message ID: " + string(rune(msgId)) + ": " + err.Error())
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

	exist, producer, err := db.GetProducerByStationIDAndUsername(producedByHeader, station.ID, connectionId)
	if err != nil {
		serv.Errorf("GetMessageDetails: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := "Some parts of the message data are missing, probably the message/the station have been deleted"
		serv.Warnf("GetMessageDetails: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	_, conn, err := db.GetConnectionByID(connectionId)
	if err != nil {
		serv.Errorf("GetMessageDetails: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	msg := models.MessageResponse{
		MessageSeq: body.MessageSeq,
		Message: models.MessagePayload{
			TimeSent: sm.Time,
			Size:     len(sm.Subject) + len(sm.Data) + len(sm.Header),
			Data:     hex.EncodeToString(sm.Data),
			Headers:  headersJson,
		},
		Producer: models.ProducerDetails{
			Name:              producedByHeader,
			ConnectionId:      connectionId,
			ClientAddress:     conn.ClientAddress,
			CreatedBy:         producer.CreatedBy,
			CreatedByUsername: producer.CreatedByUsername,
			IsActive:          producer.IsActive,
			IsDeleted:         producer.IsDeleted,
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

	schemaName := strings.ToLower(body.SchemaName)
	exist, schema, err := db.GetSchemaByName(schemaName)
	if err != nil {
		serv.Errorf("UseSchema: Schema " + body.SchemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if !exist {
		errMsg := "Schema " + schemaName + " does not exist"
		serv.Warnf("UseSchema: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	schemaVersion, err := getActiveVersionBySchemaId(schema.ID)
	if err != nil {
		serv.Errorf("UseSchema: Schema " + body.SchemaName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}
	schemaDetailsResponse := models.StationOverviewSchemaDetails{
		SchemaName:       schemaName,
		VersionNumber:    schemaVersion.VersionNumber,
		UpdatesAvailable: false,
		SchemaType:       schema.Type,
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("UseSchema: Schema " + body.SchemaName + ": " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	for _, stationName := range body.StationNames {
		stationName, err := StationNameFromStr(stationName)
		if err != nil {
			serv.Warnf("UseSchema: Schema " + body.SchemaName + " at station " + stationName.Ext() + ": " + err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}

		exist, station, err := db.GetStationByName(stationName.Ext())
		if err != nil {
			serv.Errorf("UseSchema: Schema " + body.SchemaName + " at station " + stationName.Ext() + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			errMsg := "Station " + station.Name + " does not exist"
			serv.Warnf("UseSchema: Schema " + body.SchemaName + ": " + errMsg)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
			return
		}

		err = db.AttachSchemaToStation(stationName.Ext(), schemaName, schemaVersion.VersionNumber)
		if err != nil {
			serv.Errorf("UseSchema: Schema " + body.SchemaName + " at station " + stationName.Ext() + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
			return
		}

		message := "Schema " + schemaName + " has been attached to station " + stationName.Ext() + " by user " + user.Username
		serv.Noticef(message)

		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			StationName:       stationName.Intern(),
			Message:           message,
			CreatedBy:         user.ID,
			CreatedByUsername: user.Username,
			CreatedAt:         time.Now(),
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			serv.Errorf("UseSchema: Schema " + body.SchemaName + " at station " + stationName.Ext() + " - create audit logs: " + err.Error())
		}

		updateContent, err := generateSchemaUpdateInit(schema)
		if err != nil {
			serv.Errorf("UseSchema: Schema " + body.SchemaName + " at station " + stationName.Ext() + ": " + err.Error())
			return
		}
		update := models.ProducerSchemaUpdate{
			UpdateType: models.SchemaUpdateTypeInit,
			Init:       *updateContent,
		}
		sh.S.updateStationProducersOfSchemaChange(stationName, update)

		if shouldSendAnalytics {
			user, _ := getUserDetailsFromMiddleware(c)
			param1 := analytics.EventParam{
				Name:  "station-name",
				Value: stationName.Ext(),
			}
			param2 := analytics.EventParam{
				Name:  "schema-name",
				Value: schemaName,
			}
			analyticsParams := []analytics.EventParam{param1, param2}
			analytics.SendEventWithParams(user.Username, analyticsParams, "user-attach-schema-to-station")
		}
	}

	c.IndentedJSON(200, schemaDetailsResponse)
}

func (s *Server) useSchemaDirect(c *client, reply string, msg []byte) {
	var asr attachSchemaRequest
	if err := json.Unmarshal(msg, &asr); err != nil {
		errMsg := "failed attaching schema " + asr.Name + ": " + err.Error()
		s.Errorf("useSchemaDirect: At station " + asr.StationName + " " + errMsg)
		respondWithErr(s, reply, errors.New(errMsg))
		return
	}
	stationName, err := StationNameFromStr(asr.StationName)
	if err != nil {
		serv.Warnf("useSchemaDirect: Schema " + asr.Name + " at station " + asr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	exist, _, err := db.GetStationByName(stationName.Ext())
	if err != nil {
		serv.Errorf("useSchemaDirect: Schema " + asr.Name + " at station " + asr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	if !exist {
		errMsg := "Station " + stationName.external + " does not exist"
		serv.Warnf("useSchemaDirect: " + errMsg)
		respondWithErr(s, reply, errors.New("memphis: "+errMsg))
		return
	}
	schemaName := strings.ToLower(asr.Name)
	exist, schema, err := db.GetSchemaByName(schemaName)
	if err != nil {
		serv.Errorf("useSchemaDirect: Schema " + asr.Name + " at station " + asr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}
	if !exist {
		errMsg := "Schema " + schemaName + " does not exist"
		serv.Warnf("useSchemaDirect: " + errMsg)
		respondWithErr(s, reply, errors.New(errMsg))
		return
	}

	schemaVersion, err := getActiveVersionBySchemaId(schema.ID)
	if err != nil {
		serv.Errorf("useSchemaDirect: Schema " + asr.Name + " at station " + asr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	err = db.AttachSchemaToStation(stationName.Ext(), schemaName, schemaVersion.VersionNumber)
	if err != nil {
		serv.Errorf("useSchemaDirect: Schema " + asr.Name + " at station " + asr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	username := c.getClientInfo(true).Name
	message := "Schema " + schemaName + " has been attached to station " + stationName.Ext() + " by user " + asr.Username
	serv.Noticef(message)
	_, user, err := db.GetUserByUsername(asr.Username)
	if err != nil {
		serv.Errorf("useSchemaDirect: Schema " + asr.Name + " at station " + asr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		StationName:       stationName.Intern(),
		Message:           message,
		CreatedBy:         user.ID,
		CreatedByUsername: user.Username,
		CreatedAt:         time.Now(),
	}
	auditLogs = append(auditLogs, newAuditLog)
	err = CreateAuditLogs(auditLogs)
	if err != nil {
		serv.Errorf("useSchemaDirect : Schema " + asr.Name + " at station " + asr.StationName + " - create audit logs: " + err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		param1 := analytics.EventParam{
			Name:  "station-name",
			Value: stationName.Ext(),
		}
		param2 := analytics.EventParam{
			Name:  "schema-name",
			Value: schemaName,
		}
		analyticsParams := []analytics.EventParam{param1, param2}
		analytics.SendEventWithParams(username, analyticsParams, "user-attach-schema-to-station")
	}

	updateContent, err := generateSchemaUpdateInit(schema)
	if err != nil {
		serv.Errorf("useSchemaDirect: Schema " + asr.Name + " at station " + asr.StationName + ": " + err.Error())
		return
	}

	update := models.ProducerSchemaUpdate{
		UpdateType: models.SchemaUpdateTypeInit,
		Init:       *updateContent,
	}

	serv.updateStationProducersOfSchemaChange(stationName, update)
	respondWithErr(s, reply, nil)
}

func removeSchemaFromStation(s *Server, sn StationName, updateDB bool) error {
	exist, _, err := db.GetStationByName(sn.Ext())
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("Station " + sn.external + " does not exist")
	}

	if updateDB {
		err = db.DetachSchemaFromStation(sn.Ext())
		if err != nil {
			return err
		}
	}

	update := models.ProducerSchemaUpdate{
		UpdateType: models.SchemaUpdateTypeDrop,
	}

	s.updateStationProducersOfSchemaChange(sn, update)
	return nil
}

func (s *Server) removeSchemaFromStationDirect(c *client, reply string, msg []byte) {
	var dsr detachSchemaRequest
	if err := json.Unmarshal(msg, &dsr); err != nil {
		s.Errorf("removeSchemaFromStationDirect: failed removing schema at station " + dsr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}
	stationName, err := StationNameFromStr(dsr.StationName)
	if err != nil {
		serv.Warnf("removeSchemaFromStationDirect: At station " + dsr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	err = removeSchemaFromStation(serv, stationName, true)
	if err != nil {
		serv.Errorf("removeSchemaFromStationDirect: At station " + dsr.StationName + ": " + err.Error())
		respondWithErr(s, reply, err)
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(dsr.Username, "user-detach-schema-from-station-sdk")
	}

	respondWithErr(s, reply, nil)
}

func (sh StationsHandler) RemoveSchemaFromStation(c *gin.Context) {
	// if err := DenyForSandboxEnv(c); err != nil {
	//  return
	// }

	var body models.RemoveSchemaFromStation
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("RemoveSchemaFromStation: At station" + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	exist, station, err := db.GetStationByName(stationName.Ext())
	if err != nil {
		serv.Errorf("RemoveSchemaFromStation: At station" + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := "Station " + body.StationName + " does not exist"
		serv.Warnf("RemoveSchemaFromStation: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	err = removeSchemaFromStation(sh.S, stationName, true)
	if err != nil {
		serv.Errorf("RemoveSchemaFromStation: At station" + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveSchemaFromStation: At station" + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}
	message := "Schema " + station.SchemaName + " has been deleted from station " + stationName.Ext() + " by user " + user.Username
	serv.Noticef(message)
	var auditLogs []interface{}
	newAuditLog := models.AuditLog{
		StationName:       stationName.Intern(),
		Message:           message,
		CreatedBy:         user.ID,
		CreatedByUsername: user.Username,
		CreatedAt:         time.Now(),
	}
	auditLogs = append(auditLogs, newAuditLog)
	err = CreateAuditLogs(auditLogs)
	if err != nil {
		serv.Errorf("RemoveSchemaFromStation: At station" + body.StationName + " - create audit logs error: " + err.Error())
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-remove-schema-from-station")
	}

	c.IndentedJSON(200, gin.H{})
}

func (sh StationsHandler) GetUpdatesForSchemaByStation(c *gin.Context) {
	var body models.GetUpdatesForSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("GetUpdatesForSchemaByStation: At station" + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, station, err := db.GetStationByName(stationName.Ext())
	if err != nil {
		serv.Errorf("GetUpdatesForSchemaByStation: At station" + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := "Station " + body.StationName + " does not exist"
		serv.Warnf("GetUpdatesForSchemaByStation: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	exist, schema, err := db.GetSchemaByName(station.SchemaName)
	if err != nil {
		serv.Errorf("GetUpdatesForSchemaByStation: At station" + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !exist {
		errMsg := "Schema " + station.SchemaName + " does not exist"
		serv.Warnf("GetUpdatesForSchemaByStation: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	schemasHandler := SchemasHandler{S: sh.S}
	extedndedSchemaDetails, err := schemasHandler.getExtendedSchemaDetailsUpdateAvailable(station.SchemaVersionNumber, schema)

	if err != nil {
		serv.Errorf("GetUpdatesForSchemaByStation: At station" + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-apply-schema-updates-on-station")
	}

	c.IndentedJSON(200, extedndedSchemaDetails)
}

func (sh StationsHandler) TierdStorageClicked(c *gin.Context) {
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-pushed-tierd-storage-button")
	}

	c.IndentedJSON(200, gin.H{})
}

func (sh StationsHandler) UpdateDlsConfig(c *gin.Context) {
	var body models.UpdateDlsConfigSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("DlsConfiguration: At station" + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, station, err := db.GetStationByName(stationName.Ext())
	if err != nil {
		serv.Errorf("DlsConfiguration: At station" + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := "Station " + body.StationName + " does not exist"
		serv.Warnf("DlsConfiguration: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	poisonConfigChanged := station.DlsConfigurationPoison != body.Poison
	schemaverseConfigChanged := station.DlsConfigurationSchemaverse != body.Schemaverse
	if poisonConfigChanged || schemaverseConfigChanged {
		err = db.UpdateStationDlsConfig(station.Name, body.Poison, body.Schemaverse)
		if err != nil {
			serv.Errorf("DlsConfiguration: At station" + body.StationName + ": " + err.Error())
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
	// if err := DenyForSandboxEnv(c); err != nil {
	//  return
	// }

	var body models.PurgeStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("PurgeStation: station name: " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, station, err := db.GetStationByName(stationName.Ext())
	if err != nil {
		serv.Errorf("PurgeStation: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := "Station " + stationName.external + " does not exist"
		serv.Warnf("PurgeStation: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	if body.PurgeStation {
		err = sh.S.PurgeStream(stationName.Intern())
		if err != nil && !IsNatsErr(err, JSStreamNotFoundErr) {
			serv.Errorf("PurgeStation: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	if body.PurgeDls {
		err := db.PurgeDlsMsgsFromStation(station.ID)
		if err != nil {
			serv.Errorf("PurgeStation dls: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-purge-station")
	}
	c.IndentedJSON(200, gin.H{})
}

func (sh StationsHandler) RemoveMessages(c *gin.Context) {
	// if err := DenyForSandboxEnv(c); err != nil {
	//  return
	// }

	var body models.RemoveMessagesSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("RemoveMessages: station name: " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	exist, _, err := db.GetStationByName(stationName.Ext())
	if err != nil {
		serv.Errorf("RemoveMessages: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := "Station " + stationName.external + " does not exist"
		serv.Warnf("RemoveMessages: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	for _, msg := range body.MessageSeqs {
		err = sh.S.RemoveMsg(stationName, msg)
		if err != nil {
			if IsNatsErr(err, JSStreamNotFoundErr) || IsNatsErr(err, JSStreamMsgDeleteFailedF) {
				continue
			}
			serv.Errorf("RemoveMessages: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-remove-messages")
	}

	c.IndentedJSON(200, gin.H{})
}
