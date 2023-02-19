// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
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
	ID                   primitive.ObjectID `json:"id" bson:"_id"`
	Name                 string             `json:"name" bson:"name"`
	RetentionType        string             `json:"retention_type" bson:"retention_type"`
	RetentionValue       int                `json:"retention_value" bson:"retention_value"`
	StorageType          string             `json:"storage_type" bson:"storage_type"`
	Replicas             int                `json:"replicas" bson:"replicas"`
	DedupEnabled         bool               `json:"dedup_enabled" bson:"dedup_enabled"`           // TODO deprecated
	DedupWindowInMs      int                `json:"dedup_window_in_ms" bson:"dedup_window_in_ms"` // TODO deprecated
	CreatedByUser        string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate         time.Time          `json:"creation_date" bson:"creation_date"`
	LastUpdate           time.Time          `json:"last_update" bson:"last_update"`
	Functions            []Function         `json:"functions" bson:"functions"`
	IsDeleted            bool               `json:"is_deleted" bson:"is_deleted"`
	Schema               SchemaDetails      `json:"schema" bson:"schema"`
	IdempotencyWindow    int64              `json:"idempotency_window_in_ms" bson:"idempotency_window_in_ms"`
	IsNative             bool               `json:"is_native" bson:"is_native"`
	DlsConfiguration     DlsConfiguration   `json:"dls_configuration" bson:"dls_configuration"`
	TieredStorageEnabled bool               `json:"tiered_storage_enabled" bson:"tiered_storage_enabled"`
}

type GetStationResponseSchema struct {
	ID                   primitive.ObjectID `json:"id" bson:"_id"`
	Name                 string             `json:"name" bson:"name"`
	RetentionType        string             `json:"retention_type" bson:"retention_type"`
	RetentionValue       int                `json:"retention_value" bson:"retention_value"`
	StorageType          string             `json:"storage_type" bson:"storage_type"`
	Replicas             int                `json:"replicas" bson:"replicas"`
	DedupEnabled         bool               `json:"dedup_enabled" bson:"dedup_enabled"`           // TODO deprecated
	DedupWindowInMs      int                `json:"dedup_window_in_ms" bson:"dedup_window_in_ms"` // TODO deprecated
	CreatedByUser        string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate         time.Time          `json:"creation_date" bson:"creation_date"`
	LastUpdate           time.Time          `json:"last_update" bson:"last_update"`
	Functions            []Function         `json:"functions" bson:"functions"`
	IsDeleted            bool               `json:"is_deleted" bson:"is_deleted"`
	Tags                 []CreateTag        `json:"tags"`
	IdempotencyWindow    int                `json:"idempotency_window_in_ms" bson:"idempotency_window_in_ms"`
	IsNative             bool               `json:"is_native" bson:"is_native"`
	DlsConfiguration     DlsConfiguration   `json:"dls_configuration" bson:"dls_configuration"`
	TieredStorageEnabled bool               `json:"tiered_storage_enabled" bson:"tiered_storage_enabled"`
}

type ExtendedStation struct {
	ID                   primitive.ObjectID `json:"id" bson:"_id"`
	Name                 string             `json:"name" bson:"name"`
	RetentionType        string             `json:"retention_type" bson:"retention_type"`
	RetentionValue       int                `json:"retention_value" bson:"retention_value"`
	StorageType          string             `json:"storage_type" bson:"storage_type"`
	Replicas             int                `json:"replicas" bson:"replicas"`
	DedupEnabled         bool               `json:"dedup_enabled" bson:"dedup_enabled"`           // TODO deprecated
	DedupWindowInMs      int                `json:"dedup_window_in_ms" bson:"dedup_window_in_ms"` // TODO deprecated
	CreatedByUser        string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate         time.Time          `json:"creation_date" bson:"creation_date"`
	LastUpdate           time.Time          `json:"last_update" bson:"last_update"`
	Functions            []Function         `json:"functions" bson:"functions"`
	TotalMessages        int                `json:"total_messages"`
	PoisonMessages       int                `json:"posion_messages"`
	Tags                 []CreateTag        `json:"tags"`
	IdempotencyWindow    int                `json:"idempotency_window_in_ms" bson:"idempotency_window_in_ms"`
	IsNative             bool               `json:"is_native" bson:"is_native"`
	DlsConfiguration     DlsConfiguration   `json:"dls_configuration" bson:"dls_configuration"`
	HasDlsMsgs           bool               `json:"has_dls_messages"`
	Activity             bool               `json:"activity"`
	Producers            []Producer         `json:"producers"`
	Consumers            []Consumer         `json:"consumers"`
	TieredStorageEnabled bool               `json:"tiered_storage_enabled" bson:"tiered_storage_enabled"`
}

type ExtendedStationDetails struct {
	Station        Station     `json:"station"`
	TotalMessages  int         `json:"total_messages"`
	PoisonMessages int         `json:"posion_messages"`
	Tags           []CreateTag `json:"tags"`
	HasDlsMsgs     bool        `json:"has_dls_messages"`
	Activity       bool        `json:"activity"`
}

type GetStationSchema struct {
	StationName string `form:"station_name" json:"station_name" binding:"required"`
}

type CreateStationSchema struct {
	Name                 string           `json:"name" binding:"required,min=1,max=128"`
	RetentionType        string           `json:"retention_type"`
	RetentionValue       int              `json:"retention_value"`
	Replicas             int              `json:"replicas"`
	StorageType          string           `json:"storage_type"`
	DedupEnabled         bool             `json:"dedup_enabled"`                      // TODO deprecated
	DedupWindowInMs      int              `json:"dedup_window_in_ms" binding:"min=0"` // TODO deprecated
	Tags                 []CreateTag      `json:"tags"`
	SchemaName           string           `json:"schema_name"`
	IdempotencyWindow    int64            `json:"idempotency_window_in_ms"`
	DlsConfiguration     DlsConfiguration `json:"dls_configuration"`
	TieredStorageEnabled bool             `json:"tiered_storage_enabled"`
}

type DlsConfiguration struct {
	Poison      bool `json:"poison" bson:"poison"`
	Schemaverse bool `json:"schemaverse" bson:"schemaverse"`
}

type UpdateDlsConfigSchema struct {
	StationName string `json:"station_name" binding:"required"`
	Poison      bool   `json:"poison"`
	Schemaverse bool   `json:"schemaverse"`
}

type DropDlsMessagesSchema struct {
	DlsMsgType    string   `json:"dls_type" binding:"required"`
	DlsMessageIds []string `json:"dls_message_ids" binding:"required"`
}

type ResendPoisonMessagesSchema struct {
	PoisonMessageIds []string `json:"poison_message_ids" binding:"required"`
}

type RemoveStationSchema struct {
	StationNames []string `json:"station_names" binding:"required"`
}

type GetPoisonMessageJourneySchema struct {
	MessageId string `form:"message_id" json:"message_id" binding:"required"`
}

type GetMessageDetailsSchema struct {
	IsDls       bool   `form:"is_dls" json:"is_dls"`
	DlsType     string `form:"dls_type" json:"dls_type"`
	MessageId   string `form:"message_id" json:"message_id"`
	MessageSeq  int    `form:"message_seq" json:"message_seq"`
	StationName string `form:"station_name" json:"station_name" binding:"required"`
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

type StationMsgsDetails struct {
	HasDlsMsgs    bool `json:"has_dls_messages"`
	TotalMessages int  `json:"total_messages"`
}
