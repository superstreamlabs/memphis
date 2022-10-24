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

type MessagePayload struct {
	TimeSent time.Time `json:"time_sent" bson:"time_sent"`
	Size     int       `json:"size" bson:"size"`
	Data     string    `json:"data" bson:"data"`
	Headers  string    `json:"headers" bson:"headers"`
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
	Message      MessagePayload     `json:"message" bson:"message"`
	CreationDate time.Time          `json:"creation_date" bson:"creation_date"`
}

type LightPoisonMessage struct {
	ID      primitive.ObjectID `json:"_id" bson:"_id"`
	Message MessagePayload     `json:"message" bson:"message"`
}

type LightweightPoisonMessage struct {
	ID   primitive.ObjectID `json:"_id" bson:"_id"`
	Data string             `json:"data" bson:"data"`
}
