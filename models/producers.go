package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Producer struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	StationId     primitive.ObjectID `json:"station_id" bson:"station_id"`
	FactoryId     primitive.ObjectID `json:"factory_id" bson:"factory_id"`
	Type          string             `json:"type" bson:"type"`
	ConnectionId  primitive.ObjectID `json:"connection_id" bson:"connection_id"`
	CreatedByUser string             `json:"created_by_user" bson:"created_by_user"`
	IsActive      bool               `json:"is_active" bson:"is_active"`
	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
}

type GetAllProducersByStationSchema struct {
	StationName string `form:"station_name" binding:"required" bson:"station_name"`
}
