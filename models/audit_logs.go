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

type AuditLog struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	StationName   string             `json:"station_name" bson:"station_name"`
	Message       string             `json:"message" bson:"message"`
	CreatedByUser string             `json:"created_by_user" bson:"created_by_user"`
	UserType      string             `json:"user_type" bson:"user_type"`
	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
}

type GetAllAuditLogsByStationSchema struct {
	StationName string `form:"station_name" binding:"required"`
}
