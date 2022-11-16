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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Integration struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	Name       string             `json:"name" bson:"name"`
	Keys       map[string]string  `json:"keys" bson:"keys"`
	Properties map[string]bool    `json:"properties" bson:"properties"`
}

type SlackIntegrationRequest struct {
	AuthToken                 string `json:"auth_token"`
	ChannelID                 string `json:"channel_id"`
	PoisonMessageAlert        bool   `json:"poison_message_alert"`
	SchemaValidationFailAlert bool   `json:"schema_validation_fail_alert"`
}

type GetIntegrationDetailsRequest struct {
	Name string `form:"name" json:"name" binding:"required"`
}
