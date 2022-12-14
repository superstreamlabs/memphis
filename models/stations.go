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

type MessageResponse struct {
	MessageSeq  int             `json:"message_seq" bson:"message_seq"`
	Producer    ProducerDetails `json:"producer" bson:"producer"`
	PoisonedCgs []PoisonedCg    `json:"poisoned_cgs" bson:"poisoned_cgs"`
	Message     MessagePayload  `json:"message" bson:"message"`
}

type MessageDetails struct {
	MessageSeq   int               `json:"message_seq" bson:"message_seq"`
	ProducedBy   string            `json:"produced_by" bson:"produced_by"`
	Data         string            `json:"data" bson:"data"`
	TimeSent     time.Time         `json:"creation_date" bson:"creation_date"`
	ConnectionId string            `json:"connection_id" bson:"connection_id"`
	Size         int               `json:"size" bson:"size"`
	Headers      map[string]string `json:"headers" bson:"headers"`
}

type Station struct {
	ID                primitive.ObjectID `json:"id" bson:"_id"`
	Name              string             `json:"name" bson:"name"`
	RetentionType     string             `json:"retention_type" bson:"retention_type"`
	RetentionValue    int                `json:"retention_value" bson:"retention_value"`
	StorageType       string             `json:"storage_type" bson:"storage_type"`
	Replicas          int                `json:"replicas" bson:"replicas"`
	DedupEnabled      bool               `json:"dedup_enabled" bson:"dedup_enabled"`           // TODO deprecated
	DedupWindowInMs   int                `json:"dedup_window_in_ms" bson:"dedup_window_in_ms"` // TODO deprecated
	CreatedByUser     string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate      time.Time          `json:"creation_date" bson:"creation_date"`
	LastUpdate        time.Time          `json:"last_update" bson:"last_update"`
	Functions         []Function         `json:"functions" bson:"functions"`
	IsDeleted         bool               `json:"is_deleted" bson:"is_deleted"`
	Schema            SchemaDetails      `json:"schema" bson:"schema"`
	IdempotencyWindow int                `json:"idempotency_window_in_ms" bson:"idempotency_window_in_ms"`
	DlsConfiguration  DlsConfiguration   `json:"dls_configuration" bson:"dls_configuration"`
}

type GetStationResponseSchema struct {
	ID                primitive.ObjectID `json:"id" bson:"_id"`
	Name              string             `json:"name" bson:"name"`
	RetentionType     string             `json:"retention_type" bson:"retention_type"`
	RetentionValue    int                `json:"retention_value" bson:"retention_value"`
	StorageType       string             `json:"storage_type" bson:"storage_type"`
	Replicas          int                `json:"replicas" bson:"replicas"`
	DedupEnabled      bool               `json:"dedup_enabled" bson:"dedup_enabled"`           // TODO deprecated
	DedupWindowInMs   int                `json:"dedup_window_in_ms" bson:"dedup_window_in_ms"` // TODO deprecated
	CreatedByUser     string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate      time.Time          `json:"creation_date" bson:"creation_date"`
	LastUpdate        time.Time          `json:"last_update" bson:"last_update"`
	Functions         []Function         `json:"functions" bson:"functions"`
	IsDeleted         bool               `json:"is_deleted" bson:"is_deleted"`
	Tags              []CreateTag        `json:"tags"`
	IdempotencyWindow int                `json:"idempotency_window_in_ms" bson:"idempotency_window_in_ms"`
	DlsConfiguration  DlsConfiguration   `json:"dls_configuration" bson:"dls_configuration"`
}

type ExtendedStation struct {
	ID                primitive.ObjectID `json:"id" bson:"_id"`
	Name              string             `json:"name" bson:"name"`
	RetentionType     string             `json:"retention_type" bson:"retention_type"`
	RetentionValue    int                `json:"retention_value" bson:"retention_value"`
	StorageType       string             `json:"storage_type" bson:"storage_type"`
	Replicas          int                `json:"replicas" bson:"replicas"`
	DedupEnabled      bool               `json:"dedup_enabled" bson:"dedup_enabled"`           // TODO deprecated
	DedupWindowInMs   int                `json:"dedup_window_in_ms" bson:"dedup_window_in_ms"` // TODO deprecated
	CreatedByUser     string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate      time.Time          `json:"creation_date" bson:"creation_date"`
	LastUpdate        time.Time          `json:"last_update" bson:"last_update"`
	Functions         []Function         `json:"functions" bson:"functions"`
	TotalMessages     int                `json:"total_messages"`
	PoisonMessages    int                `json:"posion_messages"`
	Tags              []CreateTag        `json:"tags"`
	IdempotencyWindow int                `json:"idempotency_window_in_ms" bson:"idempotency_window_in_ms"`
	DlsConfiguration  DlsConfiguration   `json:"dls_configuration" bson:"dls_configuration"`
}

type ExtendedStationDetails struct {
	Station        Station     `json:"station"`
	TotalMessages  int         `json:"total_messages"`
	PoisonMessages int         `json:"posion_messages"`
	Tags           []CreateTag `json:"tags"`
}

type GetStationSchema struct {
	StationName string `form:"station_name" json:"station_name" binding:"required"`
}

type CreateStationSchema struct {
	Name              string           `json:"name" binding:"required,min=1,max=32"`
	RetentionType     string           `json:"retention_type"`
	RetentionValue    int              `json:"retention_value"`
	Replicas          int              `json:"replicas"`
	StorageType       string           `json:"storage_type"`
	DedupEnabled      bool             `json:"dedup_enabled"`                      // TODO deprecated
	DedupWindowInMs   int              `json:"dedup_window_in_ms" binding:"min=0"` // TODO deprecated
	Tags              []CreateTag      `json:"tags"`
	SchemaName        string           `json:"schema_name"`
	IdempotencyWindow int              `json:"idempotency_window_in_ms"`
	DlsConfiguration  DlsConfiguration `json:"dls_configuration"`
}

type DlsConfiguration struct {
	Poison      bool `json:"poison" bson:"poison"`
	Schemaverse bool `json:"schemaverse" bson:"schemaverse"`
}

type DlsConfigurationSchema struct {
	StationName string `json:"station_name" binding:"required"`
	Poison      bool   `json:"poison" binding:"required"`
	Schemaverse bool   `json:"schemaverse" binding:"required"`
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
	StationNames []string `json:"station_names" binding:"required"`
	SchemaName   string   `json:"schema_name" binding:"required"`
}

type RemoveSchemaFromStation struct {
	StationName string `json:"station_name" binding:"required"`
}

type SchemaDetails struct {
	SchemaName    string `json:"name" bson:"name"`
	VersionNumber int    `json:"version_number" bson:"version_number"`
}

type StationOverviewSchemaDetails struct {
	SchemaName       string `json:"name" bson:"name"`
	VersionNumber    int    `json:"version_number" bson:"version_number"`
	UpdatesAvailable bool   `json:"updates_available"`
}

type GetUpdatesForSchema struct {
	StationName string `form:"station_name" json:"station_name" binding:"required"`
}
