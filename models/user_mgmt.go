// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
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
	FullName        string             `json:"full_name" bson:"full_name"`
	Subscribtion    bool               `json:"subscription" bson:"subscription"`
	SkipGetStarted  bool               `json:"skip_get_started" bson:"skip_get_started"`
}

type Image struct {
	ID    primitive.ObjectID `json:"id" bson:"_id"`
	Name  string             `json:"name" bson:"name"`
	Image string             `json:"image" bson:"image"`
}

type AddUserSchema struct {
	Username     string `json:"username" binding:"required,min=1,max=60"`
	Password     string `json:"password"`
	HubUsername  string `json:"hub_username"`
	HubPassword  string `json:"hub_password"`
	UserType     string `json:"user_type" binding:"required"`
	AvatarId     int    `json:"avatar_id"`
	FullName     string `json:"full_name"`
	Subscribtion bool   `json:"subscription"`
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

type GetFilterDetailsSchema struct {
	Route string `form:"route" json:"route"`
}

type FilteredUser struct {
	Username string `json:"_id" bson:"_id"` //_id holds username, returning value from query.
}

type ChangePasswordSchema struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
