package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Factory struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	ApplicationId string             `json:"application_id" bson:"application_id"`
	
	// RetentionType
	// RetentionValue
	// MaxThroughputType
	// MaxThroughputValue

	// Publishers
	// Consumers
	// Functions

	CreatedByUSer string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
	LastUpdate  time.Time          `json:"last_update" bson:"last_update"`
}

type GetFactoryByIdSchema struct {
	FactoryId        string `json:"factory_id" binding:"required"`
}

type GetApplicationFactoriesSchema struct {
	ApplicationId primitive.ObjectID `json:"application_id" binding:"required"`
}

type CreateFactorySchema struct {
	BoxId       primitive.ObjectID `json:"box_id" binding:"required"`
	Name        string             `json:"box_name"`
	Description string             `json:"box_description"`
}

type RemoveFactorySchema struct {
	FactoryId        string `json:"factory_id" binding:"required"`
}
