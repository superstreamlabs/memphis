package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Application struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	Description   string             `json:"description" bson:"description"`
	CreatedByUSer string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
}

type CreateApplicationSchema struct {
	Name        string `json:"name" binding:"required,min=1,max=25"`
	Description string `json:"description"`
}

type RemoveApplicationSchema struct {
	ApplicationName string `json:"application_name"  binding:"required"`
}

type EditApplicationSchema struct {
	ApplicationName string `json:"application_name"  binding:"required"`
	NewName         string `json:"application_new_name"`
	NewDescription  string `json:"application_new_description"`
}
