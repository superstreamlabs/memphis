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
// limitations under the License.

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Connection struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	CreatedByUser string             `json:"created_by_user" bson:"created_by_user"`
	IsActive      bool               `json:"is_active" bson:"is_active"`
	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
	LastPing      time.Time          `json:"last_ping" bson:"last_ping"`
	ClientAddress string             `json:"client_address" bson:"client_address"`
}
