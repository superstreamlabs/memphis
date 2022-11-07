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

type Tag struct {
	ID       primitive.ObjectID   `json:"id" bson:"_id"`
	Name     string               `json:"name" bson:"name"`
	Color    string               `json:"color" bson:"color"`
	Users    []primitive.ObjectID `json:"users" bson:"users"`
	Stations []primitive.ObjectID `json:"stations" bson:"stations"`
	Schemas  []primitive.ObjectID `json:"schemas" bson:"schemas"`
}

type CreateTag struct {
	Name  string `json:"name" binding:"required,min=1,max=20"`
	Color string `json:"color"`
}

type RemoveTagSchema struct {
	Name       string `json:"name"`
	EntityType string `json:"entity_type"`
	EntityName string `json:"entity_name"`
}

type UpdateTagsForEntitySchema struct {
	TagsToAdd    []CreateTag `json:"tags_to_add"`
	TagsToRemove []string    `json:"tags_to_remove"`
	EntityType   string      `json:"entity_type"`
	EntityName   string      `json:"entity_name"`
}

type GetTagsSchema struct {
	EntityType string `json:"entity_type"`
}
