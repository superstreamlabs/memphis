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

type Consumer struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	StationId           int       `json:"station_id"`
	Type                string    `json:"type"`
	ConnectionId        string    `json:"connection_id"`
	ConsumersGroup      string    `json:"consumers_group"`
	MaxAckTimeMs        int64     `json:"max_ack_time_ms"`
	IsActive            bool      `json:"is_active"`
	UpdatedAt           time.Time `json:"updated_at"`
	MaxMsgDeliveries    int       `json:"max_msg_deliveries"`
	StartConsumeFromSeq uint64    `json:"start_consume_from_seq"`
	LastMessages        int64     `json:"last_messages"`
	TenantName          string    `json:"tenant_name"`
	PartitionsList      []int     `json:"partitions_list"`
	Version             int       `json:"version"`
	Sdk                 string    `json:"sdk"`
	AppId               string    `json:"app_id"`
}

type ExtendedConsumer struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	UpdatedAt        time.Time `json:"updated_at"`
	IsActive         bool      `json:"is_active"`
	ConsumersGroup   string    `json:"consumers_group"`
	MaxAckTimeMs     int64     `json:"max_ack_time_ms"`
	MaxMsgDeliveries int       `json:"max_msg_deliveries"`
	StationName      string    `json:"station_name,omitempty"`
	PartitionsList   []int     `json:"partitions_list"`
	Count            int       `json:"count"`
	Version          int       `json:"version"`
	Sdk              string    `json:"sdk"`
}

type ExtendedConsumerResponse struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	UpdatedAt        time.Time `json:"updated_at"`
	IsActive         bool      `json:"is_active"`
	ConsumersGroup   string    `json:"consumers_group"`
	MaxAckTimeMs     int64     `json:"max_ack_time_ms"`
	MaxMsgDeliveries int       `json:"max_msg_deliveries"`
	StationName      string    `json:"station_name,omitempty"`
	PartitionsList   []int     `json:"partitions_list"`
	Count            int       `json:"count"`
	SdkLanguage      string    `json:"sdk_language"`
	UpdateAvailable  bool      `json:"update_available"`
}

type LightConsumer struct {
	Name        string `json:"name"`
	StationName string `json:"station_name"`
	Count       int    `json:"count"`
}

type Cg struct {
	Name                  string                     `json:"name"`
	UnprocessedMessages   int                        `json:"unprocessed_messages"`
	PoisonMessages        int                        `json:"poison_messages"`
	IsActive              bool                       `json:"is_active"`
	InProcessMessages     int                        `json:"in_process_messages"`
	MaxAckTimeMs          int64                      `json:"max_ack_time_ms"`
	MaxMsgDeliveries      int                        `json:"max_msg_deliveries"`
	ConnectedConsumers    []ExtendedConsumerResponse `json:"connected_consumers"`
	DisconnectedConsumers []ExtendedConsumerResponse `json:"disconnected_consumers"`
	DeletedConsumers      []ExtendedConsumerResponse `json:"deleted_consumers"`
	LastStatusChangeDate  time.Time                  `json:"last_status_change_date"`
	PartitionsList        []int                      `json:"partitions_list"`
	SdkLanguage           string                     `json:"sdk_language"`
	UpdateAvailable       bool                       `json:"update_available"`
}

type GetAllConsumersByStationSchema struct {
	StationName string `form:"station_name" binding:"required"`
}

type CreateConsumerSchema struct {
	Name             string `json:"name" binding:"required"`
	StationName      string `json:"station_name" binding:"required"`
	ConnectionId     string `json:"connection_id" binding:"required"`
	ConsumerType     string `json:"consumer_type" binding:"required"`
	ConsumersGroup   string `json:"consumers_group"`
	MaxAckTimeMs     int64  `json:"max_ack_time_ms"`
	MaxMsgDeliveries int    `json:"max_msg_deliveries"`
}

type DestroyConsumerSchema struct {
	Name        string `json:"name" binding:"required"`
	StationName string `json:"station_name" binding:"required"`
}

type CgMember struct {
	Name             string `json:"name"`
	ConnectionID     string `json:"Connection_id"`
	IsActive         bool   `json:"is_active"`
	MaxMsgDeliveries int    `json:"max_msg_deliveries"`
	MaxAckTimeMs     int64  `json:"max_ack_time_ms"`
	PartitionsList   []int  `json:"partitions_list"`
	Count            int    `json:"count"`
}

type DelayedCg struct {
	CGName           string `json:"cg_name"`
	NumOfDelayedMsgs uint64 `json:"num_of_delayed_msgs"`
}

type DelayedCgResp struct {
	StationName string      `json:"station_name"`
	CGS         []DelayedCg `json:"cgs"`
}

type LightCG struct {
	CGName         string `json:"cg_name"`
	StationName    string `json:"station_name"`
	StationId      int    `json:"station_id"`
	TenantName     string `json:"tenant_name"`
	PartitionsList []int  `json:"partitions_list"`
}

type ConsumerForGraph struct {
	Name      string `json:"name"`
	CGName    string `json:"cg_name"`
	StationId int    `json:"station_id"`
	AppId     string `json:"app_id"`
}
