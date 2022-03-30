package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	Username        string             `json:"username" bson:"username"`
	Password        string             `json:"password" bson:"password"`
	HubUsername     string             `json:"hub_username" bson:"hub_username"`
	HubPassword     string             `json:"hub_password" bson:"hub_password"`
	UserType        string             `json:"user_type" bson:"user_type"`
	AlreadyLoggedIn bool               `json:"already_logged_in" bson:"already_logged_in"`
	CreationDate    time.Time          `json:"creation_date" bson:"creation_date"`
	AvatarId        int                `json:"avatar_id" bson:"avatar_id"`
}

type Token struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Username     string             `json:"username" bson:"username"`
	JwtToken     string             `json:"jwt_token" bson:"jwt_token"`
	RefreshToken string             `json:"refresh_token" bson:"refresh_token"`
}

type AddUserSchema struct {
	Username    string `json:"username" binding:"required,min=1,max=25"`
	Password    string `json:"password" binding:"required,min=6"`
	HubUsername string `json:"hub_username"`
	HubPassword string `json:"hub_password"`
	UserType    string `json:"user_type" binding:"required"`
	AvatarId    int    `json:"avatar_id" bson:"avatar_id"`
}

type CreateRootUserSchema struct {
	Username    string `json:"username" binding:"required,min=1,max=25"`
	Password    string `json:"password" binding:"required,min=6"`
	HubUsername string `json:"hub_username"`
	HubPassword string `json:"hub_password"`
	AvatarId    int    `json:"avatar_id" bson:"avatar_id"`
}

type AuthenticateNatsSchema struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginSchema struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RemoveUserSchema struct {
	Username string `json:"username" binding:"required"`
}

type EditHubCredsSchema struct {
	HubUsername string `json:"hub_username" binding:"required"`
	HubPassword string `json:"hub_password" binding:"required"`
}

type EditAvatarSchema struct {
	AvatarId int `json:"avatar_id" binding:"required"`
}
