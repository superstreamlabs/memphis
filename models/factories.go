package models

// import (
// 	"time"

// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// type Box struct {
// 	ID            primitive.ObjectID `json:"id" bson:"_id"`
// 	Name          string             `json:"name" binding:"required,min=2,max=25" bson:"name"`
// 	Description   string             `json:"description" bson:"description"`
// 	CreatedByUSer string             `json:"created_by_user" binding:"required" bson:"created_by_user"`
// 	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
// }

// type CreateBoxSchema struct {
// 	Name        string `json:"name"`
// 	Description string `json:"description"`
// }

// type RemoveBoxSchema struct {
// 	BoxId primitive.ObjectID `json:"box_id"`
// }

// type EditBoxSchema struct {
// 	BoxId       primitive.ObjectID `json:"box_id"`
// 	Name        string             `json:"box_name"`
// 	Description string             `json:"box_description"`
// }
