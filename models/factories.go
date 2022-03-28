package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Factory struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	Description   string             `json:"description" bson:"description"`
	CreatedByUSer string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
}

type CreateFactorySchema struct {
	Name        string `json:"name" binding:"required,min=1,max=25"`
	Description string `json:"description"`
}

type GetFactorySchema struct {
	FactoryName string `form:"factory_name" json:"factory_name"  binding:"required"`
}

type RemoveFactorySchema struct {
	FactoryName string `json:"factory_name"  binding:"required"`
}

type EditFactorySchema struct {
	FactoryName    string `json:"factory_name"  binding:"required"`
	NewName        string `json:"factory_new_name"`
	NewDescription string `json:"factory_new_description"`
}
