package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Application struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" binding:"required,min=2,max=25" bson:"name"`
	Description   string             `json:"description" bson:"description"`
	CreatedByUSer string             `json:"created_by_user" binding:"required" bson:"created_by_user"`
	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
}

type CreateApplicationSchema struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RemoveApplicationSchema struct {
	ApplicationId primitive.ObjectID `json:"application_id"`
}

type EditApplicationSchema struct {
	ApplicationId       primitive.ObjectID `json:"application_id"`
	Name        string             `json:"application_name"`
	Description string             `json:"application_description"`
}
