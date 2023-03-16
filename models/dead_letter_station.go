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

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProducerDetails struct {
	Name          string             `json:"name" bson:"name"`
	ClientAddress string             `json:"client_address" bson:"client_address"`
	ConnectionId  primitive.ObjectID `json:"connection_id" bson:"connection_id"`
	CreatedByUser string             `json:"created_by_user" bson:"created_by_user"`
	IsActive      bool               `json:"is_active" bson:"is_active"`
	IsDeleted     bool               `json:"is_deleted" bson:"is_deleted"`
}

type MsgHeader struct {
	HeaderKey   string `json:"header_key" bson:"header_key"`
	HeaderValue string `json:"header_value" bson:"header_value"`
}

type MessagePayload struct {
	TimeSent time.Time         `json:"time_sent"`
	Size     int               `json:"size"`
	Data     string            `json:"data"`
	Headers  map[string]string `json:"headers"`
}

type MessagePayloadDls struct {
	TimeSent time.Time         `json:"time_sent"`
	Size     int               `json:"size"`
	Data     string            `json:"data"`
	Headers  map[string]string `json:"headers"`
}

type PoisonedCg struct {
	CgName              string     `json:"cg_name" bson:"cg_name"`
	PoisoningTime       time.Time  `json:"poisoning_time" bson:"poisoning_time"`
	DeliveriesCount     int        `json:"deliveries_count" bson:"deliveries_count"`
	UnprocessedMessages int        `json:"unprocessed_messages" bson:"unprocessed_messages"`
	MaxAckTimeMs        int64      `json:"max_ack_time_ms" bson:"max_ack_time_ms"`
	InProcessMessages   int        `json:"in_process_messages" bson:"in_process_messages"`
	TotalPoisonMessages int        `json:"total_poison_messages" bson:"total_poison_messages"`
	MaxMsgDeliveries    int        `json:"max_msg_deliveries" bson:"max_msg_deliveries"`
	CgMembers           []CgMember `json:"cg_members" bson:"cg_members"`
	IsActive            bool       `json:"is_active" bson:"is_active"`
	IsDeleted           bool       `json:"is_deleted" bson:"is_deleted"`
}

type DlsMessage struct {
	ID           string            `json:"_id"`
	StationName  string            `json:"station_name"`
	MessageSeq   int               `json:"message_seq"`
	Producer     ProducerDetails   `json:"producer"`
	PoisonedCg   PoisonedCg        `json:"poisoned_cg"`
	Message      MessagePayloadDls `json:"message"`
	CreationDate time.Time         `json:"creation_date"`
	CreationUnix int64             `json:"creation_unix"`
}

type DlsMessageResponse struct {
	ID           string            `json:"_id"`
	StationName  string            `json:"station_name"`
	SchemaType   string            `json:"station_type"`
	MessageSeq   int               `json:"message_seq"`
	Producer     ProducerDetails   `json:"producer"`
	PoisonedCgs  []PoisonedCg      `json:"poisoned_cgs"`
	Message      MessagePayloadDls `json:"message"`
	CreationDate time.Time         `json:"creation_date"`
}

type PmAckMsg struct {
	ID       string `json:"id" binding:"required"`
	CgName   string `json:"cg_name"`
	Sequence string `json:"sequence"`
}

type LightDlsMessage struct {
	MessageSeq int               `json:"message_seq"`
	ID         string            `json:"_id"`
	Message    MessagePayloadDls `json:"message" bson:"message"`
}

type LightDlsMessageResponse struct {
	MessageSeq int               `json:"message_seq"`
	ID         string            `json:"_id"`
	Message    MessagePayloadDls `json:"message"`
}
