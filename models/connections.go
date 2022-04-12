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
}
