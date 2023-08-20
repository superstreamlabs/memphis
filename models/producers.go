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
package models

import (
	"time"
)

type Producer struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	StationId      int       `json:"station_id"`
	Type           string    `json:"type"`
	ConnectionId   string    `json:"connection_id"`
	IsActive       bool      `json:"is_active"`
	UpdatedAt      time.Time `json:"updated_at"`
	TenantName     string    `json:"tenant_name"`
	PartitionsList []int     `json:"partitions"`
	Version        int       `json:"version"`
	Sdk            string    `json:"sdk"`
	AppId          string    `json:"app_id"`
}

type ExtendedProducer struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type,omitempty"`
	ConnectionId string    `json:"connection_id,omitempty"`
	UpdatedAt    time.Time `json:"created_at"`
	StationName  string    `json:"station_name"`
	IsActive     bool      `json:"is_active"`
	Count        int       `json:"count"`
}

type LightProducer struct {
	Name        string `json:"name"`
	StationName string `json:"station_name"`
	Count       int    `json:"count"`
}

type GetAllProducersByStationSchema struct {
	StationName string `form:"station_name" binding:"required"`
}

type CreateProducerSchema struct {
	Name         string `json:"name" binding:"required"`
	StationName  string `json:"station_name" binding:"required"`
	ConnectionId string `json:"connection_id" binding:"required"`
	ProducerType string `json:"producer_type" binding:"required"`
}

type DestroyProducerSchema struct {
	Name        string `json:"name" binding:"required"`
	StationName string `json:"station_name" binding:"required"`
}

type ProducerForGraph struct {
	Name      string `json:"name"`
	StationId int    `json:"station_id"`
	AppId     string `json:"app_id"`
	IsActive  string `json:"is_active"`
}

type ProducerForGraphWithCount struct {
	Name      string `json:"name"`
	StationId int    `json:"station_id"`
	AppId     string `json:"app_id"`
	Count     int    `json:"count"`
}
