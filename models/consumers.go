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

type Consumer struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	Name             string             `json:"name" bson:"name"`
	StationId        primitive.ObjectID `json:"station_id" bson:"station_id"`
	FactoryId        primitive.ObjectID `json:"factory_id" bson:"factory_id"`
	Type             string             `json:"type" bson:"type"`
	ConnectionId     primitive.ObjectID `json:"connection_id" bson:"connection_id"`
	ConsumersGroup   string             `json:"consumers_group" bson:"consumers_group"`
	MaxAckTimeMs     int64              `json:"max_ack_time_ms" bson:"max_ack_time_ms"`
	CreatedByUser    string             `json:"created_by_user" bson:"created_by_user"`
	IsActive         bool               `json:"is_active" bson:"is_active"`
	CreationDate     time.Time          `json:"creation_date" bson:"creation_date"`
	IsDeleted        bool               `json:"is_deleted" bson:"is_deleted"`
	MaxMsgDeliveries int                `json:"max_msg_deliveries" bson:"max_msg_deliveries"`
}

type ExtendedConsumer struct {
	Name             string    `json:"name" bson:"name"`
	CreatedByUser    string    `json:"created_by_user" bson:"created_by_user"`
	CreationDate     time.Time `json:"creation_date" bson:"creation_date"`
	IsActive         bool      `json:"is_active" bson:"is_active"`
	IsDeleted        bool      `json:"is_deleted" bson:"is_deleted"`
	ClientAddress    string    `json:"client_address" bson:"client_address"`
	ConsumersGroup   string    `json:"consumers_group" bson:"consumers_group"`
	MaxAckTimeMs     int64     `json:"max_ack_time_ms" bson:"max_ack_time_ms"`
	MaxMsgDeliveries int       `json:"max_msg_deliveries" bson:"max_msg_deliveries"`
}

type Cg struct {
	Name                  string             `json:"name" bson:"name"`
	UnprocessedMessages   int                `json:"unprocessed_messages" bson:"unprocessed_messages"`
	PoisonMessages        int                `json:"poison_messages" bson:"poison_messages"`
	IsActive              bool               `json:"is_active" bson:"is_active"`
	IsDeleted             bool               `json:"is_deleted" bson:"is_deleted"`
	InProcessMessages     int                `json:"in_process_messages" bson:"in_process_messages"`
	MaxAckTimeMs          int64              `json:"max_ack_time_ms" bson:"max_ack_time_ms"`
	MaxMsgDeliveries      int                `json:"max_msg_deliveries" bson:"max_msg_deliveries"`
	ConnectedConsumers    []ExtendedConsumer `json:"connected_consumers" bson:"connected_consumers"`
	DisconnectedConsumers []ExtendedConsumer `json:"disconnected_consumers" bson:"disconnected_consumers"`
	DeletedConsumers      []ExtendedConsumer `json:"deleted_consumers" bson:"deleted_consumers"`
	LastStatusChangeDate  time.Time          `json:"last_status_change_date" bson:"last_status_change_date"`
}

type GetAllConsumersByStationSchema struct {
	StationName string `form:"station_name" binding:"required" bson:"station_name"`
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
	Name             string `json:"name" bson:"name"`
	ClientAddress    string `json:"client_address" bson:"client_address"`
	IsActive         bool   `json:"is_active" bson:"is_active"`
	IsDeleted        bool   `json:"is_deleted" bson:"is_deleted"`
	CreatedByUser    string `json:"created_by_user" bson:"created_by_user"`
	MaxMsgDeliveries int    `json:"max_msg_deliveries" bson:"max_msg_deliveries"`
	MaxAckTimeMs     int64  `json:"max_ack_time_ms" bson:"max_ack_time_ms"`
}
