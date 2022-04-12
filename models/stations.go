package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FunctionParam struct {
	ParamName  string `json:"param_name" bson:"param_name"`
	ParamType  string `json:"param_type" bson:"param_type"`
	ParamValue string `json:"param_value" bson:"param_value"`
}

type Function struct {
	StepNumber     int                `json:"step_number" bson:"step_number"`
	FunctionId     primitive.ObjectID `json:"function_id" bson:"function_id"`
	FunctionParams []FunctionParam    `json:"function_params" bson:"function_params"`
}

type Station struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	Name            string             `json:"name" bson:"name"`
	FactoryId       primitive.ObjectID `json:"factory_id" bson:"factory_id"`
	RetentionType   string             `json:"retention_type" bson:"retention_type"`
	RetentionValue  int                `json:"retention_value" bson:"retention_value"`
	StorageType     string             `json:"storage_type" bson:"storage_type"`
	Replicas        int                `json:"replicas" bson:"replicas"`
	DedupEnabled    bool               `json:"dedup_enabled" bson:"dedup_enabled"`
	DedupWindowInMs int                `json:"dedup_window_in_ms" bson:"dedup_window_in_ms"`
	CreatedByUser   string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate    time.Time          `json:"creation_date" bson:"creation_date"`
	LastUpdate      time.Time          `json:"last_update" bson:"last_update"`
	Functions       []Function         `json:"functions" bson:"functions"`
}

type GetStationSchema struct {
	StationName string `form:"station_name" json:"station_name" binding:"required"`
}

type CreateStationSchema struct {
	Name            string `json:"name" binding:"required,min=1,max=32"`
	FactoryName     string `json:"factory_name" binding:"required"`
	RetentionType   string `json:"retention_type"`
	RetentionValue  int    `json:"retention_value"`
	Replicas        int    `json:"replicas"`
	StorageType     string `json:"storage_type"`
	DedupEnabled    bool   `json:"dedup_enabled"`
	DedupWindowInMs int    `json:"dedup_window_in_ms" binding:"min=0"`
}

type RemoveStationSchema struct {
	StationName string `json:"station_name" binding:"required"`
}
