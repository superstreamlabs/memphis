package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Connection struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	CreationDate time.Time          `json:"creation_date" bson:"creation_date"`
}
