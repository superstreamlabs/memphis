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

type CreateTagsSchema struct {
	Tags       []CreateTag `json:"tags"`
	EntityType string      `json:"entity_type"`
	EntityName string      `json:"entity_name"`
}

type RemoveTagsSchema struct {
	Names      []string `json:"names"`
	EntityType string   `json:"entity_type"`
	EntityName string   `json:"entity_name"`
}

type GetTagsSchema struct {
	EntityType string `json:"entity_type"`
}
