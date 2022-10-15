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

type FunctionParam struct {
	ParamName  string `json:"param_name" bson:"param_name"`
	ParamType  string `json:"param_type" bson:"param_type"`
	ParamValue string `json:"param_value" bson:"param_value"`
}

type Function struct {
	StepNumber     int                `json:"step_number" bson:"step_number"`
	FunctionId     primitive.ObjectID `json:"function_id" bson:"function_id"`
	FunctionParams []FunctionParam    `json:"function_params" bson:"function_params"`
}

type Message struct {
	MessageSeq  int             `json:"message_seq" bson:"message_seq"`
	Producer    ProducerDetails `json:"producer" bson:"producer"`
	PoisonedCgs []PoisonedCg    `json:"poisoned_cgs" bson:"poisoned_cgs"`
	Message     MessagePayload  `json:"message" bson:"message"`
}

type MessageDetails struct {
	MessageSeq   int       `json:"message_seq" bson:"message_seq"`
	ProducedBy   string    `json:"produced_by" bson:"produced_by"`
	Data         string    `json:"data" bson:"data"`
	TimeSent     time.Time `json:"creation_date" bson:"creation_date"`
	ConnectionId string    `json:"connection_id" bson:"connection_id"`
	Size         int       `json:"size" bson:"size"`
}

type Station struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	Name            string             `json:"name" bson:"name"`
	RetentionType   string             `json:"retention_type" bson:"retention_type"`
	RetentionValue  int                `json:"retention_value" bson:"retention_value"`
	StorageType     string             `json:"storage_type" bson:"storage_type"`
	Replicas        int                `json:"replicas" bson:"replicas"`
	DedupEnabled    bool               `json:"dedup_enabled" bson:"dedup_enabled"`
	DedupWindowInMs int                `json:"dedup_window_in_ms" bson:"dedup_window_in_ms"`
	CreatedByUser   string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate    time.Time          `json:"creation_date" bson:"creation_date"`
	LastUpdate      time.Time          `json:"last_update" bson:"last_update"`
	Functions       []Function         `json:"functions" bson:"functions"`
	IsDeleted       bool               `json:"is_deleted" bson:"is_deleted"`
	SchemaName      string             `json:"schema_name" bson:"schema_name"`
}

type GetStationResponseSchema struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	Name            string             `json:"name" bson:"name"`
	RetentionType   string             `json:"retention_type" bson:"retention_type"`
	RetentionValue  int                `json:"retention_value" bson:"retention_value"`
	StorageType     string             `json:"storage_type" bson:"storage_type"`
	Replicas        int                `json:"replicas" bson:"replicas"`
	DedupEnabled    bool               `json:"dedup_enabled" bson:"dedup_enabled"`
	DedupWindowInMs int                `json:"dedup_window_in_ms" bson:"dedup_window_in_ms"`
	CreatedByUser   string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate    time.Time          `json:"creation_date" bson:"creation_date"`
	LastUpdate      time.Time          `json:"last_update" bson:"last_update"`
	Functions       []Function         `json:"functions" bson:"functions"`
	IsDeleted       bool               `json:"is_deleted" bson:"is_deleted"`
	Tags            []Tag              `json:"tags"`
}

type ExtendedStation struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	Name            string             `json:"name" bson:"name"`
	RetentionType   string             `json:"retention_type" bson:"retention_type"`
	RetentionValue  int                `json:"retention_value" bson:"retention_value"`
	StorageType     string             `json:"storage_type" bson:"storage_type"`
	Replicas        int                `json:"replicas" bson:"replicas"`
	DedupEnabled    bool               `json:"dedup_enabled" bson:"dedup_enabled"`
	DedupWindowInMs int                `json:"dedup_window_in_ms" bson:"dedup_window_in_ms"`
	CreatedByUser   string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate    time.Time          `json:"creation_date" bson:"creation_date"`
	LastUpdate      time.Time          `json:"last_update" bson:"last_update"`
	Functions       []Function         `json:"functions" bson:"functions"`
	TotalMessages   int                `json:"total_messages"`
	PoisonMessages  int                `json:"posion_messages"`
	Tags            []Tag              `json:"tags"`
}

type ExtendedStationDetails struct {
	Station        Station `json:"station"`
	TotalMessages  int     `json:"total_messages"`
	PoisonMessages int     `json:"posion_messages"`
	Tags           []Tag   `json:"tags"`
}

type GetStationSchema struct {
	StationName string `form:"station_name" json:"station_name" binding:"required"`
}

type CreateStationSchema struct {
	Name            string      `json:"name" binding:"required,min=1,max=32"`
	RetentionType   string      `json:"retention_type"`
	RetentionValue  int         `json:"retention_value"`
	Replicas        int         `json:"replicas"`
	StorageType     string      `json:"storage_type"`
	DedupEnabled    bool        `json:"dedup_enabled"`
	DedupWindowInMs int         `json:"dedup_window_in_ms" binding:"min=0"`
	Tags            []CreateTag `json:"tags"`
	SchemaName      string      `json:"schema_name"`
}

type AckPoisonMessagesSchema struct {
	PoisonMessageIds []primitive.ObjectID `json:"poison_message_ids" binding:"required"`
}

type ResendPoisonMessagesSchema struct {
	PoisonMessageIds []primitive.ObjectID `json:"poison_message_ids" binding:"required"`
}

type RemoveStationSchema struct {
	StationNames []string `json:"station_names" binding:"required"`
}

type GetPoisonMessageJourneySchema struct {
	MessageId string `form:"message_id" json:"message_id" binding:"required"`
}

type GetMessageDetailsSchema struct {
	IsPoisonMessage bool   `form:"is_poison_message" json:"is_poison_message"`
	MessageId       string `form:"message_id" json:"message_id"`
	MessageSeq      int    `form:"message_seq" json:"message_seq"`
	StationName     string `form:"station_name" json:"station_name" binding:"required"`
}

type UseSchema struct {
	StationName string `json:"station_name" binding:"required"`
	SchemaName  string `json:"schema_name" binding:"required"`
}
