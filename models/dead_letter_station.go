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

type ProducerDetails struct {
	Name              string `json:"name"`
	ConnectionId      string `json:"connection_id"`
	CreatedBy         int    `json:"created_by"`
	IsActive          bool   `json:"is_active"`
	IsDeleted         bool   `json:"is_deleted"`
	CreatedByUsername string `json:"created_by_username"`
}

type ProducerDetailsResp struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}
type MsgHeader struct {
	HeaderKey   string `json:"header_key"`
	HeaderValue string `json:"header_value"`
}

type MessagePayload struct {
	TimeSent time.Time         `json:"time_sent"`
	Size     int               `json:"size"`
	Data     string            `json:"data"`
	Headers  map[string]string `json:"headers"`
}

type PoisonedCg struct {
	CgName              string     `json:"cg_name"`
	UnprocessedMessages int        `json:"unprocessed_messages"`
	MaxAckTimeMs        int64      `json:"max_ack_time_ms"`
	InProcessMessages   int        `json:"in_process_messages"`
	TotalPoisonMessages int        `json:"total_poison_messages"`
	MaxMsgDeliveries    int        `json:"max_msg_deliveries"`
	CgMembers           []CgMember `json:"cg_members"`
	IsActive            bool       `json:"is_active"`
	IsDeleted           bool       `json:"is_deleted"`
}

type PoisonedCgResponse struct {
	CgName []string `json:"cg_name"`
}

type SchemaVerseDlsMessageSdk struct {
	StationName     string          `json:"station_name"`
	Producer        ProducerDetails `json:"producer"`
	Message         MessagePayload  `json:"message"`
	ValidationError string          `json:"validation_error"`
	PartitionNumber int             `json:"partition_number"`
}

type FunctionsMessagePayload struct {
	TimeSent time.Time         `json:"time_sent"`
	Size     int               `json:"size"`
	Data     []byte            `json:"data"`
	Headers  map[string]string `json:"headers"`
}

type FunctionsDlsMessage struct {
	StationID       int            `json:"station_id"`
	Message         MessagePayload `json:"message"`
	Err             string         `json:"err"`
	PartitionNumber int            `json:"partition_number"`
}

type DlsMessage struct {
	ID              int            `json:"id"`
	StationId       int            `json:"station_id"`
	MessageSeq      int            `json:"message_seq"`
	PoisonedCgs     []string       `json:"poisoned_cgs"`
	MessageDetails  MessagePayload `json:"message_details"`
	UpdatedAt       time.Time      `json:"updated_at"`
	MessageType     string         `json:"message_type"`
	ValidationError string         `json:"validation_error"`
	TenantName      string         `json:"tenant_name"`
	ProducerName    string         `json:"producer_name"`
	PartitionNumber int            `json:"partition_number"`
	FunctionId      int            `json:"function_id"`
}

type DlsMsgResendAll struct {
	MinId int `json:"min_id"`
	MaxId int `json:"max_id"`
}

type DlsMessageResponse struct {
	ID              int                 `json:"id"`
	StationName     string              `json:"station_name"`
	SchemaType      string              `json:"schema_type"`
	MessageSeq      int                 `json:"message_seq"`
	Producer        ProducerDetailsResp `json:"producer"`
	PoisonedCgs     []PoisonedCg        `json:"poisoned_cgs"`
	Message         MessagePayload      `json:"message"`
	UpdatedAt       time.Time           `json:"updated_at"`
	ValidationError string              `json:"validation_error"`
}

type PmAckMsg struct {
	ID     int    `json:"id" binding:"required"`
	CgName string `json:"cg_name"`
}

type LightDlsMessage struct {
	MessageSeq int            `json:"message_seq"`
	ID         string         `json:"_id"`
	Message    MessagePayload `json:"message"`
}

type LightDlsMessageResponse struct {
	MessageSeq   int            `json:"message_seq"`
	ID           int            `json:"id"`
	Message      MessagePayload `json:"message"`
	FunctionName string         `json:"function_name"`
}

type RetentionIntervalData struct {
	Updated_at time.Time `json:"updated_at"`
}
