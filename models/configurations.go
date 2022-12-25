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

import "go.mongodb.org/mongo-driver/bson/primitive"

type EditClusterConfigSchema struct {
	PMRetention   int `json:"pm_retention" binding:"required"`
	LogsRetention int `json:"logs_retention" binding:"required"`
}

type ConfigurationsUpdate struct {
	StationName string `json:"station_name"`
	Type        string `json:"type"`
	Update      any    `json:"update"`
}

type ConfigurationsIntValue struct {
	ID    primitive.ObjectID `json:"id" bson:"_id"`
	Key   string             `json:"key" bson:"key"`
	Value int                `json:"value" bson:"value"`
}

type ConfigurationsStringValue struct {
	ID    primitive.ObjectID `json:"id" bson:"_id"`
	Key   string             `json:"key" bson:"key"`
	Value string             `json:"value" bson:"value"`
}
