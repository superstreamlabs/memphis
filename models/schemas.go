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

type Schema struct {
	ID   primitive.ObjectID `json:"id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
	Type string             `json:"type" bson:"type"`
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
	SchemaUpdateTypeEnableDLS
	SchemaUpdateTypeDisableDLS
)

func SchemaverseDLSConfToSchemaUpdateType(enable bool) ProducerSchemaUpdateType {
	if enable {
		return SchemaUpdateTypeEnableDLS
	}
	return SchemaUpdateTypeDisableDLS
}

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
