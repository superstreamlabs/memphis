package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Username     string             `json:"username" binding:"required,min=1,max=25" bson:"username"`
	Password     string             `json:"password" binding:"required,min=6" bson:"password"`
	HubUsername  string             `json:"hub_username" bson:"hub_username"`
	HubPassword  string             `json:"hub_password" bson:"hub_password"`
	UserType     string             `json:"user_type" binding:"required" bson:"user_type"`
	CreationDate time.Time          `json:"creation_date" bson:"creation_date"`
}

func (u User) GetUserWithoutPassword() User {
	user := User{
		ID:           u.ID,
		Username:     u.Username,
		Password:     "",
		HubUsername:  u.HubUsername,
		HubPassword:  "",
		UserType:     u.UserType,
		CreationDate: u.CreationDate,
	}

	return user
}

type AddUserSchema struct {
	Username    string `json:"username" binding:"required,min=1,max=25" bson:"username"`
	Password    string `json:"password" binding:"required,min=6" bson:"password"`
	HubUsername string `json:"hub_username" bson:"hub_username"`
	HubPassword string `json:"hub_password" bson:"hub_password"`
	UserType    string `json:"user_type" binding:"required" bson:"user_type"`
}

type CreateRootUserSchema struct {
	Username    string `json:"username" binding:"required,min=1,max=25" bson:"username"`
	Password    string `json:"password" binding:"required,min=6" bson:"password"`
	HubUsername string `json:"hub_username" bson:"hub_username"`
	HubPassword string `json:"hub_password" bson:"hub_password"`
}

type AuthenticateNatsSchema struct {
	Username string `json:"username" binding:"required" bson:"username"`
	Password string `json:"password" binding:"required" bson:"password"`
}
