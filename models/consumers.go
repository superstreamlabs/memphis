// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Consumer struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	Name           string             `json:"name" bson:"name"`
	StationId      primitive.ObjectID `json:"station_id" bson:"station_id"`
	FactoryId      primitive.ObjectID `json:"factory_id" bson:"factory_id"`
	Type           string             `json:"type" bson:"type"`
	ConnectionId   primitive.ObjectID `json:"connection_id" bson:"connection_id"`
	ConsumersGroup string             `json:"consumers_group" bson:"consumers_group"`
	MaxAckTimeMs   int64              `json:"max_ack_time_ms" bson:"max_ack_time_ms"`
	CreatedByUser  string             `json:"created_by_user" bson:"created_by_user"`
	IsActive       bool               `json:"is_active" bson:"is_active"`
	CreationDate   time.Time          `json:"creation_date" bson:"creation_date"`
}

type GetAllConsumersByStationSchema struct {
	StationName string `form:"station_name" binding:"required" bson:"station_name"`
}

type CreateConsumerSchema struct {
	Name           string `json:"name" binding:"required"`
	StationName    string `json:"station_name" binding:"required"`
	ConnectionId   string `json:"connection_id" binding:"required"`
	ConsumerType   string `json:"consumer_type" binding:"required"`
	ConsumersGroup string `json:"consumers_group"`
	MaxAckTimeMs   int64  `json:"max_ack_time_ms"`
}

type DestroyConsumerSchema struct {
	Name        string `json:"name" binding:"required"`
	StationName string `json:"station_name" binding:"required"`
}
