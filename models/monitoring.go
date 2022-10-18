// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package models

import "time"

type SystemComponent struct {
	Component   string `json:"component"`
	DesiredPods int    `json:"desired_pods"`
	ActualPods  int    `json:"actual_pods"`
}

type MainOverviewData struct {
	TotalStations    int               `json:"total_stations"`
	TotalMessages    int               `json:"total_messages"`
	SystemComponents []SystemComponent `json:"system_components"`
	Stations         []ExtendedStation `json:"stations"`
}

type StationOverviewData struct {
	ConnectedProducers    []ExtendedProducer   `json:"connected_producers"`
	DisconnectedProducers []ExtendedProducer   `json:"disconnected_producers"`
	DeletedProducers      []ExtendedProducer   `json:"deleted_producers"`
	ConnectedCgs          []Cg                 `json:"connected_cgs"`
	DisconnectedCgs       []Cg                 `json:"disconnected_cgs"`
	DeletedCgs            []Cg                 `json:"deleted_cgs"`
	TotalMessages         int                  `json:"total_messages"`
	AvgMsgSize            int64                `json:"average_message_size"`
	AuditLogs             []AuditLog           `json:"audit_logs"`
	Messages              []MessageDetails     `json:"messages"`
	PoisonMessages        []LightPoisonMessage `json:"poison_messages"`
	Tags                  []Tag                `json:"tags"`
	Leader                string               `json:"leader"`
	Followers             []string             `json:"followers"`
	SchemaName            string               `json:"schema_name"`
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
