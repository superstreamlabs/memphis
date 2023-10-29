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

type User struct {
	ID              int       `json:"id"`
	Username        string    `json:"username"`
	Password        string    `json:"password"`
	UserType        string    `json:"user_type"`
	AlreadyLoggedIn bool      `json:"already_logged_in"`
	CreatedAt       time.Time `json:"created_at"`
	AvatarId        int       `json:"avatar_id"`
	FullName        string    `json:"full_name"`
	Subscribtion    bool      `json:"subscription"`
	SkipGetStarted  bool      `json:"skip_get_started"`
	TenantName      string    `json:"tenant_name"`
	Pending         bool      `json:"pending"`
	Position        string    `json:"position"`
	Team            string    `json:"team"`
	Owner           string    `json:"owner"`
	Description     string    `json:"description"`
}

type Image struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

type AddUserSchema struct {
	Username     string `json:"username" binding:"required,min=1"`
	Password     string `json:"password"`
	UserType     string `json:"user_type" binding:"required"`
	AvatarId     int    `json:"avatar_id"`
	FullName     string `json:"full_name"`
	Subscribtion bool   `json:"subscription"`
	Team         string `json:"team"`
	Position     string `json:"position"`
	Owner        string `json:"owner"`
	Description  string `json:"description"`
}

type AuthenticateNatsSchema struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RemoveUserSchema struct {
	Username string `json:"username" binding:"required"`
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
	Username string `json:"_id"` //_id holds username, returning value from query.
}

type ChangePasswordSchema struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SendTraceSchema struct {
	TraceName   string                 `json:"trace_name" binding:"required"`
	TraceParams map[string]interface{} `json:"trace_params" binding:"required"`
}

type FilteredGenericUser struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	UserType    string    `json:"user_type"`
	CreatedAt   time.Time `json:"created_at"`
	AvatarId    int       `json:"avatar_id"`
	FullName    string    `json:"full_name"`
	Pending     bool      `json:"pending"`
	Position    string    `json:"position"`
	Team        string    `json:"team"`
	Owner       string    `json:"owner"`
	Description string    `json:"description"`
}

type FilteredApplicationUser struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}
