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

type Schema struct {
	ID   primitive.ObjectID `json:"id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
	Type string             `json:"type" bson:"type"`
}

type SchemaPg struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type SchemaVersion struct {
	ID                primitive.ObjectID `json:"id" bson:"_id"`
	VersionNumber     int                `json:"version_number" bson:"version_number"`
	Active            bool               `json:"active" bson:"active"`
	CreatedByUser     string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate      time.Time          `json:"creation_date" bson:"creation_date"`
	SchemaContent     string             `json:"schema_content" bson:"schema_content"`
	SchemaId          primitive.ObjectID `json:"schema_id" bson:"schema_id"`
	MessageStructName string             `json:"message_struct_name" bson:"message_struct_name"`
	Descriptor        string             `json:"-" bson:"descriptor"`
}

type SchemaVersionPg struct {
	ID                int       `json:"id" `
	VersionNumber     int       `json:"version_number"`
	Active            bool      `json:"active"`
	CreatedByUser     int       `json:"created_by_user"`
	CreationDate      time.Time `json:"creation_date"`
	SchemaContent     string    `json:"schema_content"`
	SchemaId          int       `json:"schema_id"`
	MessageStructName string    `json:"message_struct_name"`
	Descriptor        string    `json:"-"`
}

type CreateNewSchema struct {
	Name              string      `json:"name" binding:"required,min=1,max=32"`
	Type              string      `json:"type"`
	SchemaContent     string      `json:"schema_content"`
	Tags              []CreateTag `json:"tags"`
	MessageStructName string      `json:"message_struct_name"`
}

type ExtendedSchema struct {
	ID                  primitive.ObjectID `json:"id" bson:"_id"`
	Name                string             `json:"name" bson:"name"`
	Type                string             `json:"type" bson:"type"`
	CreatedByUser       string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate        time.Time          `json:"creation_date" bson:"creation_date"`
	ActiveVersionNumber int                `json:"active_version_number" bson:"version_number"`
	Used                bool               `json:"used"`
	Tags                []CreateTag        `json:"tags"`
}

type ExtendedSchemaDetails struct {
	ID           primitive.ObjectID `json:"id"`
	SchemaName   string             `json:"schema_name"`
	Type         string             `json:"type"`
	Versions     []SchemaVersion    `json:"versions"`
	UsedStations []string           `json:"used_stations"`
	Tags         []CreateTag        `json:"tags"`
}

type ProducerSchemaUpdateType int

const (
	SchemaUpdateTypeInit ProducerSchemaUpdateType = iota + 1
	SchemaUpdateTypeDrop
)

type ProducerSchemaUpdate struct {
	UpdateType ProducerSchemaUpdateType
	Init       ProducerSchemaUpdateInit `json:"init,omitempty"`
}

type ProducerSchemaUpdateInit struct {
	SchemaName    string                      `json:"schema_name"`
	ActiveVersion ProducerSchemaUpdateVersion `json:"active_version"`
	SchemaType    string                      `json:"type"`
}

type ProducerSchemaUpdateVersion struct {
	VersionNumber     int    `json:"version_number"`
	Descriptor        string `json:"descriptor"`
	Content           string `json:"schema_content"`
	MessageStructName string `json:"message_struct_name"`
}

type GetSchemaDetails struct {
	SchemaName string `form:"schema_name" json:"schema_name"`
}

type RemoveSchema struct {
	SchemaNames []string `json:"schema_names" binding:"required"`
}

type CreateNewVersion struct {
	SchemaName        string `json:"schema_name"`
	SchemaContent     string `json:"schema_content"`
	MessageStructName string `json:"message_struct_name"`
}

type RollBackVersion struct {
	SchemaName    string `json:"schema_name"`
	VersionNumber int    `json:"version_number"`
}

type ValidateSchema struct {
	SchemaType    string `json:"schema_type"`
	SchemaContent string `json:"schema_content"`
}
