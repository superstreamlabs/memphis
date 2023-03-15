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

type Producer struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	StationId     primitive.ObjectID `json:"station_id" bson:"station_id"`
	Type          string             `json:"type" bson:"type"`
	ConnectionId  primitive.ObjectID `json:"connection_id" bson:"connection_id"`
	CreatedByUser string             `json:"created_by_user" bson:"created_by_user"`
	IsActive      bool               `json:"is_active" bson:"is_active"`
	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
	IsDeleted     bool               `json:"is_deleted" bson:"is_deleted"`
}

type ProducerPg struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	StationId    int       `json:"station_id"`
	Type         string    `json:"type"`
	ConnectionId int       `json:"connection_id"`
	CreatedBy    int       `json:"created_by"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	IsDeleted    bool      `json:"is_deleted"`
}

type ExtendedProducer struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	Type          string             `json:"type" bson:"type"`
	ConnectionId  primitive.ObjectID `json:"connection_id" bson:"connection_id"`
	CreatedByUser string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
	StationName   string             `json:"station_name" bson:"station_name"`
	IsActive      bool               `json:"is_active" bson:"is_active"`
	IsDeleted     bool               `json:"is_deleted" bson:"is_deleted"`
	ClientAddress string             `json:"client_address" bson:"client_address"`
}

type GetAllProducersByStationSchema struct {
	StationName string `form:"station_name" binding:"required" bson:"station_name"`
}

type CreateProducerSchema struct {
	Name         string `json:"name" binding:"required"`
	StationName  string `json:"station_name" binding:"required"`
	ConnectionId string `json:"connection_id" binding:"required"`
	ProducerType string `json:"producer_type" binding:"required"`
}

type DestroyProducerSchema struct {
	Name        string `json:"name" binding:"required"`
	StationName string `json:"station_name" binding:"required"`
}
