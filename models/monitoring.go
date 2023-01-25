// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package models

import "time"

type SysComponent struct {
	Name    string    `json:"name"`
	CPU     CompStats `json:"cpu"`
	Memory  CompStats `json:"memory"`
	Storage CompStats `json:"storage"`
	Healthy bool      `json:"healthy"`
}

type CompStats struct {
	Total      float64 `json:"total"`
	Current    float64 `json:"current"`
	Percentage int   `json:"percentage"`
}

type SystemComponents struct {
	Name        string         `json:"name"`
	Components  []SysComponent `json:"components"`
	Status      string         `json:"status"`
	Ports       []int          `json:"ports"`
	DesiredPods int            `json:"desired_pods"`
	ActualPods  int            `json:"actual_pods"`
	Host        string         `json:"host"`
}

type MainOverviewData struct {
	TotalStations     int                `json:"total_stations"`
	TotalMessages     int                `json:"total_messages"`
	SystemComponents  []SystemComponents `json:"system_components"`
	Stations          []ExtendedStation  `json:"stations"`
	K8sEnv            bool               `json:"k8s_env"`
	BrokersThroughput []BrokerThroughput `json:"brokers_throughput"`
}

type GetStationOverviewDataSchema struct {
	StationName string `form:"station_name" json:"station_name"  binding:"required"`
}

type SystemLogsRequest struct {
	LogType  string `form:"log_type" json:"log_type"  binding:"required"`
	StartIdx int    `form:"start_index" json:"start_index"  binding:"required"`
}

type Log struct {
	MessageSeq int       `json:"message_seq"`
	Type       string    `json:"type"`
	Source     string    `json:"source"`
	Data       string    `json:"data"`
	TimeSent   time.Time `json:"creation_date"`
}

type SystemLogsResponse struct {
	Logs []Log `json:"logs"`
}

type ProxyMonitoringResponse struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Storage float64 `json:"storage"`
}

type BrokerThroughput struct {
	Name  string `json:"name"`
	Read  int64  `json:"read"`
	Write int64  `json:"write"`
}

type Throughput struct {
	Bytes       int64 `json:"bytes"`
	BytesPerSec int64 `json:"bytes_per_sec"`
}
