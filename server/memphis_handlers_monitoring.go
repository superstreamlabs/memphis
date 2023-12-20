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
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/memphisdev/memphis/analytics"
	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"
	"github.com/memphisdev/memphis/utils"

	"github.com/gin-gonic/gin"

	v1 "k8s.io/api/core/v1"
	v1Apimachinery "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type MonitoringHandler struct{ S *Server }

var clientset *kubernetes.Clientset
var metricsclientset *metricsv.Clientset
var config *rest.Config
var noMetricsInstalledLog bool
var noMetricsPermissionLog bool

const (
	healthyStatus                  = "healthy"
	unhealthyStatus                = "unhealthy"
	dangerousStatus                = "dangerous"
	riskyStatus                    = "risky"
	lastProducerCreationReqVersion = 3
	lastConsumerCreationReqVersion = 3
)

func clientSetClusterConfig() error {
	var err error
	// in cluster config
	config, err = rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err = kubernetes.NewForConfig(config)

	if err != nil {
		return err
	}
	if metricsclientset == nil {
		metricsclientset, err = metricsv.NewForConfig(config)
		if err != nil {
			return err
		}
	}

	noMetricsInstalledLog = false
	noMetricsPermissionLog = false

	return nil
}

func (mh MonitoringHandler) GetClusterInfo(c *gin.Context) {
	c.IndentedJSON(200, gin.H{"version": mh.S.MemphisVersion()})
}

func (mh MonitoringHandler) GetMainOverviewData(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetMainOverviewData at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	response, err := mh.getMainOverviewDataDetails(user.TenantName)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "cannot connect to the docker daemon") {
			serv.Warnf("[tenant: %v][user: %v]GetMainOverviewData: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Failed getting system components data: " + err.Error()})
		} else {
			serv.Errorf("[tenant: %v][user: %v]GetMainOverviewData: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		}
	}
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-enter-main-overview")
	}

	c.IndentedJSON(200, response)
}

func getFakeProdsAndConsForPreview() ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, []map[string]interface{}) {
	connectedProducers := make([]map[string]interface{}, 0)
	connectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3d8",
		"name":            "prod.20",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.999Z",
		"station_name":    "idanasulin6",
		"is_active":       true,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	connectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3d6",
		"name":            "prod.19",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.99Z",
		"station_name":    "idanasulin6",
		"is_active":       true,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	connectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3d4",
		"name":            "prod.18",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.982Z",
		"station_name":    "idanasulin6",
		"is_active":       true,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	connectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3d2",
		"name":            "prod.17",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.969Z",
		"station_name":    "idanasulin6",
		"is_active":       true,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})

	disconnectedProducers := make([]map[string]interface{}, 0)
	disconnectedProducers = append(disconnectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3d0",
		"name":            "prod.16",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.959Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(disconnectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3ce",
		"name":            "prod.15",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.951Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(disconnectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3cc",
		"name":            "prod.14",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.941Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(disconnectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3ca",
		"name":            "prod.13",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.93Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(disconnectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3c8",
		"name":            "prod.12",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.92Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(disconnectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3c6",
		"name":            "prod.11",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.911Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(disconnectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3c4",
		"name":            "prod.10",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.902Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(disconnectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3c2",
		"name":            "prod.9",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.892Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(disconnectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3c0",
		"name":            "prod.8",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.882Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(disconnectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3be",
		"name":            "prod.7",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.872Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(disconnectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3bc",
		"name":            "prod.6",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": ROOT_USERNAME,
		"creation_date":   "2023-01-05T08:44:36.862Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})

	connectedCgs := make([]map[string]interface{}, 0)
	connectedCgs = append(connectedCgs, map[string]interface{}{
		"name":                    "cg.20",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               true,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	connectedCgs = append(connectedCgs, map[string]interface{}{
		"name":                    "cg.19",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               true,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	connectedCgs = append(connectedCgs, map[string]interface{}{
		"name":                    "cg.18",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               true,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	connectedCgs = append(connectedCgs, map[string]interface{}{
		"name":                    "cg.17",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               true,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	connectedCgs = append(connectedCgs, map[string]interface{}{
		"name":                    "cg.16",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               true,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})

	disconnectedCgs := make([]map[string]interface{}, 0)
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.15",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.14",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.13",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.12",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.11",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.10",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.9",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})

	return connectedProducers, disconnectedProducers, connectedCgs, disconnectedCgs
}

func (mh MonitoringHandler) GetStationOverviewData(c *gin.Context) {
	stationsHandler := StationsHandler{S: mh.S}
	producersHandler := ProducersHandler{S: mh.S}
	consumersHandler := ConsumersHandler{S: mh.S}
	auditLogsHandler := AuditLogsHandler{}
	poisonMsgsHandler := PoisonMessagesHandler{S: mh.S}
	tagsHandler := TagsHandler{S: mh.S}
	var body models.GetStationOverviewDataSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("GetStationOverviewData at StationNameFromStr: At station %v: %v", body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetStationOverviewData at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	exist, station, err := db.GetStationByName(stationName.Ext(), user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData at GetStationByName: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Station %v does not exist", body.StationName)
		serv.Warnf("[tenant: %v][user: %v]GetStationOverviewData: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	stationsActAsDlsStation, err := db.GetStationsByDlsStationName(stationName.Ext(), user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData at GetStationsByDlsStationName: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	usedAsDlsStations := make([]string, 0)
	if len(stationsActAsDlsStation) > 0 {
		for _, dlsStation := range stationsActAsDlsStation {
			usedAsDlsStations = append(usedAsDlsStations, dlsStation.Name)
		}
	}
	var functionsEnabled bool
	if station.Version >= 2 {
		functionsEnabled = true
	} else {
		functionsEnabled = false
	}

	connectedProducers, disconnectedProducers, deletedProducers := make([]models.ExtendedProducerResponse, 0), make([]models.ExtendedProducerResponse, 0), make([]models.ExtendedProducerResponse, 0)
	if station.IsNative {
		connectedProducers, disconnectedProducers, deletedProducers, err = producersHandler.GetProducersByStation(station)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData at GetProducersByStation: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	auditLogs, err := auditLogsHandler.GetAuditLogsByStation(station.Name, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	var totalMessages int
	if body.PartitionNumber == -1 {
		totalMessages, err = stationsHandler.GetTotalMessages(station.TenantName, station.Name, station.PartitionsList)
		if err != nil {
			if IsNatsErr(err, JSStreamNotFoundErr) {
				serv.Warnf("[tenant: %v][user: %v]GetStationOverviewData at GetTotalMessages: nats error At station %v: does not exist", user.TenantName, user.Username, body.StationName)
				c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station " + body.StationName + " does not exist"})
			} else {
				serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			}
			return
		}
	} else {
		totalMessages, err = stationsHandler.GetTotalPartitionMessages(station.TenantName, station.Name, body.PartitionNumber)
		if err != nil {
			if IsNatsErr(err, JSStreamNotFoundErr) {
				serv.Warnf("[tenant: %v][user: %v]GetStationOverviewData at GetTotalPartitionMessages: nats error At station %v: does not exist", user.TenantName, user.Username, body.StationName)
				c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station " + body.StationName + " does not exist"})
			} else {
				serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			}
			return
		}
		valid := validatePartitionNumber(station.PartitionsList, body.PartitionNumber)
		if !valid {
			errMsg := fmt.Sprintf("Partition number %v does not exist", body.PartitionNumber)
			serv.Warnf("[tenant: %v][user: %v]GetStationOverviewData: %v", user.TenantName, user.Username, errMsg)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		}
	}

	var avgMsgSize int64
	if body.PartitionNumber == -1 {
		avgMsgSize, err = stationsHandler.GetAvgMsgSize(station)
		if err != nil {
			if IsNatsErr(err, JSStreamNotFoundErr) {
				serv.Warnf("[tenant: %v][user: %v]GetStationOverviewData at GetAvgMsgSize: At station %v: does not exist", user.TenantName, user.Username, body.StationName)
				c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station " + body.StationName + " does not exist"})
			} else {
				serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData at GetAvgMsgSize: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			}
			return
		}
	} else {
		avgMsgSize, err = stationsHandler.GetPartitionAvgMsgSize(station.TenantName, fmt.Sprintf("%v$%v", stationName.Intern(), body.PartitionNumber))
		if err != nil {
			if IsNatsErr(err, JSStreamNotFoundErr) {
				serv.Warnf("[tenant: %v][user: %v]GetStationOverviewData at GetAvgMsgSize: At station %v: does not exist", user.TenantName, user.Username, body.StationName)
				c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station " + body.StationName + " does not exist"})
			} else {
				serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData at GetAvgMsgSize: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			}
			return
		}
	}

	usageLimit, err := getUsageLimitProduersLimitPerStation(user.TenantName, body.StationName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData: At getUsageLimitProduersLimitPerStation %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	messagesToFetch := 1000
	messages := make([]models.MessageDetails, 0)
	if body.PartitionNumber == -1 {
		messages, err = stationsHandler.GetMessages(station, messagesToFetch)
		if err != nil {
			if IsNatsErr(err, JSStreamNotFoundErr) {
				serv.Warnf("[tenant: %v][user: %v]GetStationOverviewData at GetMessages: nats error At station %v: does not exist", user.TenantName, user.Username, body.StationName)
				c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station " + body.StationName + " does not exist"})
			} else {
				serv.Errorf("GetStationOverviewData at GetMessages: At station " + body.StationName + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			}
			return
		}
	} else {
		messages, err = stationsHandler.GetMessagesFromPartition(station, fmt.Sprintf("%v$%v", stationName.Intern(), body.PartitionNumber), messagesToFetch, body.PartitionNumber)
		if err != nil {
			if IsNatsErr(err, JSStreamNotFoundErr) {
				serv.Warnf("[tenant: %v][user: %v]GetStationOverviewData at GetMessagesFromPartition: nats error At station %v: does not exist", user.TenantName, user.Username, body.StationName)
				c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station " + body.StationName + " does not exist"})
			} else {
				serv.Errorf("GetStationOverviewData at GetMessagesFromPartition: At station " + body.StationName + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			}
			return
		}
	}

	poisonMessages, schemaFailedMessages, functionsMessages, totalDlsAmount, err := poisonMsgsHandler.GetDlsMsgsByStationLight(station, body.PartitionNumber)
	if err != nil {
		if IsNatsErr(err, JSStreamNotFoundErr) {
			serv.Warnf("[tenant: %v][user: %v]GetStationOverviewData at GetDlsMsgsByStationLight: nats error At station %v: does not exist", user.TenantName, user.Username, body.StationName)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station " + body.StationName + " does not exist"})
		} else {
			serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData at GetDlsMsgsByStationLight: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		}
		return
	}

	connectedCgs, disconnectedCgs, deletedCgs := make([]models.Cg, 0), make([]models.Cg, 0), make([]models.Cg, 0)

	// Only native stations have CGs
	if station.IsNative {
		connectedCgs, disconnectedCgs, deletedCgs, err = consumersHandler.GetCgsByStation(stationName, station)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData at GetCgsByStation: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	tags, err := tagsHandler.GetTagsByEntityWithID("station", station.ID)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData at GetTagsByEntityWithID: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	leader, followers, err := stationsHandler.GetLeaderAndFollowers(station, body.PartitionNumber)
	if err != nil {
		if IsNatsErr(err, JSStreamNotFoundErr) {
			serv.Warnf("[tenant: %v][user: %v]GetStationOverviewData at GetLeaderAndFollowers: nats error At station %v: does not exist", user.TenantName, user.Username, body.StationName)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station " + body.StationName + " does not exist"})
		} else {
			serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData at GetLeaderAndFollowers: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		}
		return
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

	sourceConnectors, err := mh.S.GetSourceConnectorsByStationAndPartition(station.ID, body.PartitionNumber, len(station.PartitionsList))
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData at GetSourceConnectorsByStationAndPartition: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	sinkConnectors, err := consumersHandler.GetSinkConnectorsByStation(stationName, station, body.PartitionNumber, station.PartitionsList)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData at GetSinkConnectorsByStation: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	var response gin.H

	// Check when the schema object in station is not empty, not optional for non native stations
	if station.SchemaName != "" && station.SchemaVersionNumber != 0 {
		var schemaDetails models.StationOverviewSchemaDetails
		exist, schema, err := db.GetSchemaByName(station.SchemaName, station.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData at GetSchemaByName: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			schemaDetails = models.StationOverviewSchemaDetails{}
		} else {
			_, schemaVersion, err := db.GetSchemaVersionByNumberAndID(station.SchemaVersionNumber, schema.ID)
			if err != nil {
				serv.Errorf("[tenant: %v][user: %v]GetStationOverviewData at GetSchemaVersionByNumberAndID: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			updatesAvailable := !schemaVersion.Active
			schemaDetails = models.StationOverviewSchemaDetails{
				SchemaName:       schema.Name,
				VersionNumber:    station.SchemaVersionNumber,
				UpdatesAvailable: updatesAvailable,
				SchemaType:       schema.Type,
			}
		}
		response = gin.H{
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
			"schema_failed_messages":          schemaFailedMessages,
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
			"source_connectors":               sourceConnectors,
			"sink_connectors":                 sinkConnectors,
			"act_as_dls_station_in_stations":  usedAsDlsStations,
		}
	} else {
		var emptyResponse struct{}
		if !station.IsNative {
			cp, dp, cc, dc := getFakeProdsAndConsForPreview()
			response = gin.H{
				"connected_producers":             cp,
				"disconnected_producers":          dp,
				"deleted_producers":               deletedProducers,
				"connected_cgs":                   cc,
				"disconnected_cgs":                dc,
				"deleted_cgs":                     deletedCgs,
				"total_messages":                  totalMessages,
				"average_message_size":            avgMsgSize,
				"audit_logs":                      auditLogs,
				"messages":                        messages,
				"poison_messages":                 poisonMessages,
				"schema_failed_messages":          schemaFailedMessages,
				"functions_failed_messages":       functionsMessages,
				"tags":                            tags,
				"leader":                          leader,
				"followers":                       followers,
				"schema":                          emptyResponse,
				"idempotency_window_in_ms":        station.IdempotencyWindow,
				"dls_configuration_poison":        station.DlsConfigurationPoison,
				"dls_configuration_schemaverse":   station.DlsConfigurationSchemaverse,
				"total_dls_messages":              totalDlsAmount,
				"tiered_storage_enabled":          station.TieredStorageEnabled,
				"created_by_username":             station.CreatedByUsername,
				"resend_disabled":                 station.ResendDisabled,
				"functions_enabled":               functionsEnabled,
				"max_amount_of_allowed_producers": usageLimit,
				"source_connectors":               sourceConnectors,
				"sink_connectors":                 sinkConnectors,
				"act_as_dls_station_in_stations":  usedAsDlsStations,
			}
		} else {
			response = gin.H{
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
				"schema_failed_messages":          schemaFailedMessages,
				"functions_failed_messages":       functionsMessages,
				"tags":                            tags,
				"leader":                          leader,
				"followers":                       followers,
				"schema":                          emptyResponse,
				"idempotency_window_in_ms":        station.IdempotencyWindow,
				"dls_configuration_poison":        station.DlsConfigurationPoison,
				"dls_configuration_schemaverse":   station.DlsConfigurationSchemaverse,
				"total_dls_messages":              totalDlsAmount,
				"tiered_storage_enabled":          station.TieredStorageEnabled,
				"created_by_username":             station.CreatedByUsername,
				"resend_disabled":                 station.ResendDisabled,
				"functions_enabled":               functionsEnabled,
				"max_amount_of_allowed_producers": usageLimit,
				"source_connectors":               sourceConnectors,
				"sink_connectors":                 sinkConnectors,
				"act_as_dls_station_in_stations":  usedAsDlsStations,
			}
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-enter-station-overview")
	}

	c.IndentedJSON(200, response)
}

func min(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}

func (s *Server) GetSystemLogs(amount uint64,
	timeout time.Duration,
	fromLast bool,
	lastKnownSeq uint64,
	filterSubject string,
	getAll bool) (models.SystemLogsResponse, error) {
	uid := s.memphis.nuid.Next()
	durableName := "$memphis_fetch_logs_consumer_" + uid
	var msgs []StoredMsg

	streamInfo, err := s.memphisStreamInfo(s.MemphisGlobalAccountString(), syslogsStreamName)
	if err != nil {
		return models.SystemLogsResponse{}, err
	}

	amount = min(streamInfo.State.Msgs, amount)
	startSeq := lastKnownSeq - amount + 1

	if getAll {
		startSeq = streamInfo.State.FirstSeq
		amount = streamInfo.State.Msgs
	} else if fromLast {
		startSeq = streamInfo.State.LastSeq - amount + 1

		//handle uint wrap around
		if amount >= streamInfo.State.LastSeq {
			startSeq = 1
		}
		lastKnownSeq = streamInfo.State.LastSeq

	} else if amount >= lastKnownSeq {
		startSeq = 1
		amount = lastKnownSeq
	}

	cc := ConsumerConfig{
		OptStartSeq:   startSeq,
		DeliverPolicy: DeliverByStartSequence,
		AckPolicy:     AckExplicit,
		Durable:       durableName,
		Replicas:      1,
	}

	if filterSubject != _EMPTY_ {
		cc.FilterSubject = filterSubject
	}

	err = s.memphisAddConsumer(s.MemphisGlobalAccountString(), syslogsStreamName, &cc)
	if err != nil {
		return models.SystemLogsResponse{}, err
	}

	responseChan := make(chan StoredMsg)
	subject := fmt.Sprintf(JSApiRequestNextT, syslogsStreamName, durableName)
	reply := durableName + "_reply"
	req := []byte(strconv.FormatUint(amount, 10))
	sub, err := s.subscribeOnAcc(s.MemphisGlobalAccount(), reply, reply+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(respCh chan StoredMsg, subject, reply string, msg []byte) {
			// ack
			s.sendInternalAccountMsg(s.MemphisGlobalAccount(), reply, []byte(_EMPTY_))
			rawTs := tokenAt(reply, 8)
			seq, _, _ := ackReplyInfo(reply)

			intTs, err := strconv.Atoi(rawTs)
			if err != nil {
				s.Errorf("GetSystemLogs: %v", err.Error())
				return
			}

			respCh <- StoredMsg{
				Subject:  subject,
				Sequence: uint64(seq),
				Data:     msg,
				Time:     time.Unix(0, int64(intTs)),
			}
		}(responseChan, subject, reply, copyBytes(msg))
	})
	if err != nil {
		return models.SystemLogsResponse{}, err
	}

	s.sendInternalAccountMsgWithReply(s.MemphisGlobalAccount(), subject, reply, nil, req, true)

	timer := time.NewTimer(timeout)
	for i := uint64(0); i < amount; i++ {
		select {
		case <-timer.C:
			goto cleanup
		case msg := <-responseChan:
			msgs = append(msgs, msg)
		}
	}

cleanup:
	timer.Stop()
	s.unsubscribeOnAcc(s.MemphisGlobalAccount(), sub)
	time.AfterFunc(500*time.Millisecond, func() { serv.memphisRemoveConsumer(s.MemphisGlobalAccountString(), syslogsStreamName, durableName) })

	var resMsgs []models.Log
	if uint64(len(msgs)) < amount && streamInfo.State.Msgs > amount && streamInfo.State.FirstSeq < startSeq {
		return s.GetSystemLogs(amount*2, timeout, false, lastKnownSeq, filterSubject, getAll)
	}
	for _, msg := range msgs {
		if err != nil {
			return models.SystemLogsResponse{}, err
		}

		splittedSubj := strings.Split(msg.Subject, tsep)
		var (
			logSource string
			logType   string
		)

		if len(splittedSubj) == 2 {
			// old version's logs
			logSource = "broker"
			logType = splittedSubj[1]
		} else if len(splittedSubj) == 3 {
			// old version's logs
			logSource, logType = splittedSubj[1], splittedSubj[2]
		} else {
			logSource, logType = splittedSubj[1], splittedSubj[3]
		}

		data := string(msg.Data)
		resMsgs = append(resMsgs, models.Log{
			MessageSeq: int(msg.Sequence),
			Type:       logType,
			Data:       data,
			Source:     logSource,
			TimeSent:   msg.Time,
		})
	}

	if getAll {
		sort.Slice(resMsgs, func(i, j int) bool {
			return resMsgs[i].MessageSeq < resMsgs[j].MessageSeq
		})
	} else {
		sort.Slice(resMsgs, func(i, j int) bool {
			return resMsgs[i].MessageSeq > resMsgs[j].MessageSeq
		})

		if len(resMsgs) > 100 {
			resMsgs = resMsgs[:100]
		}
	}

	return models.SystemLogsResponse{Logs: resMsgs}, nil
}

func checkCompStatus(components models.Components) string {
	if len(components.UnhealthyComponents) > 0 {
		return unhealthyStatus
	}
	if len(components.DangerousComponents) > 0 {
		return dangerousStatus
	}
	if len(components.RiskyComponents) > 0 {
		return riskyStatus
	}
	return healthyStatus
}

func getDbStorageUsage() (float64, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), db.DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := db.MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	var dbStorageSize float64
	query := `SELECT pg_database_size($1) AS db_size`
	stmt, err := conn.Conn().Prepare(ctx, "get_db_storage_size", query)
	if err != nil {
		return 0, err
	}
	err = conn.Conn().QueryRow(ctx, stmt.Name, configuration.METADATA_DB_DBNAME).Scan(&dbStorageSize)
	if err != nil {
		return 0, err
	}

	return dbStorageSize, nil
}

func getUnixStorageSize() float64 {
	out, err := exec.Command("df", "-h", "/").Output()
	if err != nil {
		serv.Errorf("getUnixStorageSize: " + err.Error())
		return 0
	}
	var storageSize float64
	output := string(out[:])
	splitted_output := strings.Split(output, "\n")
	parsedline := strings.Fields(splitted_output[1])
	if len(parsedline) > 0 {
		re := regexp.MustCompile(`^([\d.]+)([A-Za-z]+)$`)
		matches := re.FindStringSubmatch(parsedline[1])
		if len(matches) != 3 {
			serv.Warnf("getUnixStorageSize: invalid size format")
			return 0
		}
		sizeStr := matches[1]
		unit := matches[2]
		storageSize, err = strconv.ParseFloat(sizeStr, 64)
		if err != nil {
			serv.Errorf("getUnixStorageSize: " + err.Error())
			return 0
		}
		switch unit {
		case "T":
			storageSize *= 1024 * 1024 * 1024 * 1024 // Terabytes to bytes
		case "G":
			storageSize *= 1024 * 1024 * 1024 // Gigabytes to bytes
		case "M":
			storageSize *= 1024 * 1024 // Megabytes to bytes
		case "K":
			storageSize *= 1024 // Kilobytes to bytes
		case "Ti":
			storageSize *= 1024 * 1024 * 1024 * 1024 // Tebibytes to bytes
		case "Gi":
			storageSize *= 1024 * 1024 * 1024 // Gibibytes to bytes
		case "Mi":
			storageSize *= 1024 * 1024 // Mebibytes to bytes
		case "Ki":
			storageSize *= 1024 // Kibibytes to bytes
		default:
			storageSize = 0
			serv.Warnf("getUnixStorageSize: unsupported unit: %s", unit)
		}
	} else {
		serv.Warnf("getUnixStorageSize: invalid size format")
		return 0
	}
	return storageSize
}

func getUnixMemoryUsage() (float64, error) {
	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)
	out, err := exec.Command("ps", "-o", "vsz", "-p", pidStr).Output()
	if err != nil {
		return 0, err
	}
	memUsage := float64(0)
	output := string(out[:])
	splitted_output := strings.Split(output, "\n")
	parsedline := strings.Fields(splitted_output[1])
	if len(parsedline) > 0 {
		memUsage, err = strconv.ParseFloat(parsedline[0], 64)
		if err != nil {
			return 0, err
		}
	}
	return memUsage, nil
}

func defaultSystemComp(compName string, healthy bool) models.SysComponent {
	defaultStat := models.CompStats{
		Total:      0,
		Current:    0,
		Percentage: 0,
	}
	status := healthyStatus
	if !healthy {
		status = unhealthyStatus
	}
	return models.SysComponent{
		Name:    compName,
		CPU:     defaultStat,
		Memory:  defaultStat,
		Storage: defaultStat,
		Healthy: healthy,
		Status:  status,
	}
}

func getRelevantComponents(name string, components []models.SysComponent, desired int) models.Components {
	healthyComps := []models.SysComponent{}
	unhealthyComps := []models.SysComponent{}
	dangerousComps := []models.SysComponent{}
	riskyComps := []models.SysComponent{}
	for _, comp := range components {
		if name == "memphis" {
			regexMatch, _ := regexp.MatchString(`^memphis-\d*[0-9]\d*$`, comp.Name)
			if regexMatch {
				switch comp.Status {
				case unhealthyStatus:
					unhealthyComps = append(unhealthyComps, comp)
				case dangerousStatus:
					dangerousComps = append(dangerousComps, comp)
				case riskyStatus:
					riskyComps = append(riskyComps, comp)
				default:
					healthyComps = append(healthyComps, comp)
				}
			}
		} else if name == "memphis-metadata" {
			regexMatch, _ := regexp.MatchString(`^memphis-metadata-\d*[0-9]\d*$`, comp.Name)
			if regexMatch {
				switch comp.Status {
				case unhealthyStatus:
					unhealthyComps = append(unhealthyComps, comp)
				case dangerousStatus:
					dangerousComps = append(dangerousComps, comp)
				case riskyStatus:
					riskyComps = append(riskyComps, comp)
				default:
					healthyComps = append(healthyComps, comp)
				}
			}
		} else if name == "memphis-rest-gateway" || name == "memphis-metadata-coordinator" {
			if strings.Contains(comp.Name, name) {
				switch comp.Status {
				case unhealthyStatus:
					unhealthyComps = append(unhealthyComps, comp)
				case dangerousStatus:
					dangerousComps = append(dangerousComps, comp)
				case riskyStatus:
					riskyComps = append(riskyComps, comp)
				default:
					healthyComps = append(healthyComps, comp)
				}
			}
		}
	}
	missingComps := desired - (len(unhealthyComps) + len(dangerousComps) + len(riskyComps) + len(healthyComps))
	if missingComps > 0 {
		for i := 0; i < missingComps; i++ {
			unhealthyComps = append(unhealthyComps, defaultSystemComp(name, false))
		}
	}
	return models.Components{
		UnhealthyComponents: unhealthyComps,
		DangerousComponents: dangerousComps,
		RiskyComponents:     riskyComps,
		HealthyComponents:   healthyComps,
	}
}

func getRelevantPorts(name string, portsMap map[string][]int) []int {
	res := []int{}
	mPorts := make(map[int]bool)
	for key, ports := range portsMap {
		if name == "memphis" {
			keyMatchBroker, err := regexp.MatchString(`^memphis-\d*[0-9]\d*$`, key)
			if err != nil {
				return []int{}
			}
			if keyMatchBroker {
				for _, port := range ports {
					if !mPorts[port] {
						mPorts[port] = true
						res = append(res, port)
					}
				}
			}
		} else if strings.Contains(key, name) {
			for _, port := range ports {
				if !mPorts[port] {
					mPorts[port] = true
					res = append(res, port)
				}
			}
		}
	}
	return res
}

func getContainerStorageUsage(config *rest.Config, mountPath string, container string, pod string, namespace string) (float64, error) {
	usage := float64(0)
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Container: container,
		Command:   []string{"df", mountPath},
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return 0, err
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return 0, err
	}
	splitted_output := strings.Split(stdout.String(), "\n")
	parsedline := strings.Fields(splitted_output[1])
	if stderr.String() != "" {
		return usage, errors.New(stderr.String())
	}
	if len(parsedline) > 1 {
		usage, err = strconv.ParseFloat(parsedline[2], 64)
		if err != nil {
			return usage, err
		}
		usage = usage * 1024
	}

	return usage, nil
}

func shortenFloat(f float64) float64 {
	// round up very small number
	if f < float64(0.01) && f > float64(0) {
		return float64(0.01)
	}
	// shorten float to 2 decimal places
	return math.Floor(f*100) / 100
}

func (mh MonitoringHandler) GetAvailableReplicas(c *gin.Context) {
	v, err := serv.Varz(nil)
	if err != nil {
		serv.Errorf("GetAvailableReplicas: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	replicas := v.Routes + 1
	replicas = GetAvailableReplicas(replicas)
	c.IndentedJSON(200, gin.H{
		"available_replicas": replicas})
}

func checkIsMinikube(labels map[string]string) bool {
	for key := range labels {
		if strings.Contains(strings.ToLower(key), "minikube") {
			return true
		}
	}
	return false
}

func checkPodStatus(cpu int, memory int, storage int) string {
	if cpu > 99 || memory > 99 || storage > 99 {
		return unhealthyStatus
	}
	if cpu > 94 || memory > 94 || storage > 94 {
		return dangerousStatus
	}
	if cpu > 84 || memory > 84 || storage > 84 {
		return riskyStatus
	}
	return healthyStatus
}

func getComponentsStructByOneComp(comp models.SysComponent) models.Components {
	if comp.Status == unhealthyStatus {
		return models.Components{
			UnhealthyComponents: []models.SysComponent{comp},
		}
	}
	if comp.Status == dangerousStatus {
		return models.Components{
			DangerousComponents: []models.SysComponent{comp},
		}
	}
	if comp.Status == riskyStatus {
		return models.Components{
			RiskyComponents: []models.SysComponent{comp},
		}
	}
	return models.Components{
		HealthyComponents: []models.SysComponent{comp},
	}
}

func getK8sClusterTimestamp() (string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	clusterInfo, err := clientset.CoreV1().Namespaces().Get(ctx, "kube-system", v1Apimachinery.GetOptions{})
	if err != nil {
		return "", err
	}

	creationTime := clusterInfo.CreationTimestamp.Time
	unixTime := creationTime.Unix()

	return strconv.FormatInt(unixTime, 10), nil
}

func getDockerMacAddress() (string, error) {
	var macAdress string
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.HardwareAddr == nil {
			continue
		} else {
			macAdress = iface.HardwareAddr.String()
			break
		}
	}
	return macAdress, nil
}

func (mh MonitoringHandler) GetGraphOverview(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetGraphOverview at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	res, err := mh.getGraphOverview(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]%v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
	}

	c.IndentedJSON(200, res)
}

func (mh MonitoringHandler) GetSystemGeneralInfo(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetSystemGeneralInfo at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	totalAmountBrokers := 1

	v, err := serv.Varz(nil)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetSystemGeneralInfo at serv.Varz : %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if len(v.Cluster.URLs) > 0 {
		totalAmountBrokers = len(v.Cluster.URLs)
	}

	stationsCount, err := db.CountStationsByTenant(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetSystemGeneralInfo at CountStationsByTenant: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	usersCount, err := db.CountAllUsersByTenant(user.TenantName, false)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetSystemGeneralInfo at CountAllUsersByTenant: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	schemasCount, err := db.CountAllSchemasByTenant(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetSystemGeneralInfo at CountAllSchemasByTenant: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, gin.H{"total_amount_brokers": totalAmountBrokers, "total_stations": stationsCount, "total_users": usersCount, "total_schemas": schemasCount})
}
