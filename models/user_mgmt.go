// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
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
	FullName		string 			   `json:"full_name" bson:"full_name"`
	Subscribtion	bool 			   `json:"subscription" bson:"subscription"`
}

type Image struct {
	ID    primitive.ObjectID `json:"id" bson:"_id"`
	Name  string             `json:"name" bson:"name"`
	Image string             `json:"image" bson:"image"`
}

type AddUserSchema struct {
	Username    	string `json:"username" binding:"required,min=1,max=25"`
	Password    	string `json:"password"`
	HubUsername 	string `json:"hub_username"`
	HubPassword 	string `json:"hub_password"`
	UserType    	string `json:"user_type" binding:"required"`
	AvatarId    	int    `json:"avatar_id"`
	FullName		string `json:"full_name"`
	Subscribtion	bool   `json:"subscription"`
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

type EditAnalyticsSchema struct {
	SendAnalytics bool `json:"send_analytics"`
}
