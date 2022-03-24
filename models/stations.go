package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Function struct {
}

type Station struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id"`
	Name               string             `json:"name" bson:"name"`
	FactoryId          primitive.ObjectID `json:"factory_id" bson:"factory_id"`
	RetentionType      string             `json:"retention_type" bson:"retention_type"`
	RetentionValue     string             `json:"retention_value" bson:"retention_value"`
	MaxThroughputType  string             `json:"max_throughput_type" bson:"max_throughput_type"`
	MaxThroughputValue string             `json:"max_throughput_value" bson:"max_throughput_value"`
	StorageType        string             `json:"storage_type" bson:"storage_type"`
	CreatedByUSer      string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate       time.Time          `json:"creation_date" bson:"creation_date"`
	LastUpdate         time.Time          `json:"last_update" bson:"last_update"`
	Functions          []Function
}

type GetStationByIdSchema struct {
	StationId string `form:"station_id" json:"station_id" binding:"required"`
}

type GetFactoryStationsSchema struct {
	FactoryId primitive.ObjectID `json:"factory_id" binding:"required"`
}

type CreateStationSchema struct {
	Name               string             `json:"name" bson:"name"`
	FactoryName        primitive.ObjectID `json:"factory_name" bson:"factory_name"`
	RetentionType      string             `json:"retention_type" bson:"retention_type"`
	RetentionValue     string             `json:"retention_value" bson:"retention_value"`
	MaxThroughputType  string             `json:"max_throughput_type" bson:"max_throughput_type"`
	MaxThroughputValue string             `json:"max_throughput_value" bson:"max_throughput_value"`
	StorageType        string             `json:"storage_type" bson:"storage_type"`
}

type RemoveStationSchema struct {
	StationId string `json:"station_id" binding:"required"`
}
