// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package models

import (
	"time"
)

type SandboxUser struct {
	ID              int       `json:"id" bson:"_id"`
	Username        string    `json:"username" bson:"username"`
	Email           string    `json:"email" bson:"email"`
	FirstName       string    `json:"first_name" bson:"first_name"`
	LastName        string    `json:"last_name" bson:"last_name"`
	Password        string    `json:"password" bson:"password"`
	UserType        string    `json:"user_type" bson:"user_type"`
	AlreadyLoggedIn bool      `json:"already_logged_in" bson:"already_logged_in"`
	CreationDate    time.Time `json:"creation_date" bson:"creation_date"`
	AvatarId        int       `json:"avatar_id" bson:"avatar_id"`
	ProfilePic      string    `json:"profile_pic" bson:"profile_pic"`
	SkipGetStarted  bool      `json:"skip_get_started" bson:"skip_get_started"`
}

type SandboxLoginSchema struct {
	LoginType string `json:"login_type" binding:"required"`
	Token     string `json:"token" binding:"required"`
}
