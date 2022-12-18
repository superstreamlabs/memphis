// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package models
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

type MessagePayloadDb struct {
	TimeSent time.Time   `json:"time_sent" bson:"time_sent"`
	Size     int         `json:"size" bson:"size"`
	Data     string      `json:"data" bson:"data"`
	Headers  []MsgHeader `json:"headers" bson:"headers"`
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

type PoisonMessage struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id"`
	StationName  string             `json:"station_name" bson:"station_name"`
	MessageSeq   int                `json:"message_seq" bson:"message_seq"`
	Producer     ProducerDetails    `json:"producer" bson:"producer"`
	PoisonedCgs  []PoisonedCg       `json:"poisoned_cgs" bson:"poisoned_cgs"`
	Message      MessagePayloadDb   `json:"message" bson:"message"`
	CreationDate time.Time          `json:"creation_date" bson:"creation_date"`
}

type PoisonMessageResponse struct {
	ID           primitive.ObjectID `json:"_id"`
	StationName  string             `json:"station_name"`
	MessageSeq   int                `json:"message_seq"`
	Producer     ProducerDetails    `json:"producer"`
	PoisonedCgs  []PoisonedCg       `json:"poisoned_cgs"`
	Message      MessagePayload     `json:"message"`
	CreationDate time.Time          `json:"creation_date"`
}

type DlqMessage struct {
	ID           string           `json:"id"`
	StationName  string           `json:"station_name"`
	MessageSeq   int              `json:"message_seq"`
	Producer     ProducerDetails  `json:"producer"`
	PoisonedCg   PoisonedCg       `json:"poisoned_cg"`
	Message      MessagePayloadDb `json:"message"`
	CreationDate time.Time        `json:"creation_date"`
}

type DlqMessageResponse struct {
	ID           string           `json:"id"`
	StationName  string           `json:"station_name"`
	MessageSeq   int              `json:"message_seq"`
	Producer     ProducerDetails  `json:"producer"`
	PoisonedCg   PoisonedCg       `json:"poisoned_cg"`
	Message      MessagePayloadDb `json:"message"`
	CreationDate time.Time        `json:"creation_date"`
}

type LightPoisonMessage struct {
	ID      primitive.ObjectID `json:"_id" bson:"_id"`
	Message MessagePayloadDb   `json:"message" bson:"message"`
}

type LightPoisonMessageResponse struct {
	ID      primitive.ObjectID `json:"_id"`
	Message MessagePayload     `json:"message"`
}

type LightweightPoisonMessage struct {
	ID   primitive.ObjectID `json:"_id" bson:"_id"`
	Data string             `json:"data" bson:"data"`
}

type PmAckMsg struct {
	ID     string `json:"id" binding:"required"`
	CgName string `json:"cg_name" binding:"required"`
}
