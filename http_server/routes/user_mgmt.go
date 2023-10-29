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
package routes

import (
	"github.com/memphisdev/memphis/server"

	"github.com/gin-gonic/gin"
)

func InitializeUserMgmtRoutes(router *gin.RouterGroup) {
	userMgmtHandler := server.UserMgmtHandler{}
	userMgmtRoutes := router.Group("/usermgmt")
	userMgmtRoutes.POST("/login", userMgmtHandler.Login)
	userMgmtRoutes.POST("/doneNextSteps", userMgmtHandler.DoneNextSteps)
	userMgmtRoutes.POST("/refreshToken", userMgmtHandler.RefreshToken)
	userMgmtRoutes.POST("/addUser", userMgmtHandler.AddUser)
	userMgmtRoutes.POST("/addUserSignUp", userMgmtHandler.AddUserSignUp)
	userMgmtRoutes.GET("/getSignUpFlag", userMgmtHandler.GetSignUpFlag)
	userMgmtRoutes.GET("/getAllUsers", userMgmtHandler.GetAllUsers)
	userMgmtRoutes.GET("/getApplicationUsers", userMgmtHandler.GetApplicationUsers)
	userMgmtRoutes.DELETE("/removeUser", userMgmtHandler.RemoveUser)
	// TODO: change the name to removeAccount
	userMgmtRoutes.DELETE("/removeMyUser", userMgmtHandler.RemoveMyUser)
	userMgmtRoutes.PUT("/editAvatar", userMgmtHandler.EditAvatar)
	userMgmtRoutes.PUT("/editCompanyLogo", userMgmtHandler.EditCompanyLogo)
	userMgmtRoutes.DELETE("/removeCompanyLogo", userMgmtHandler.RemoveCompanyLogo)
	userMgmtRoutes.GET("/getCompanyLogo", userMgmtHandler.GetCompanyLogo)
	userMgmtRoutes.PUT("/editAnalytics", userMgmtHandler.EditAnalytics)
	userMgmtRoutes.POST("/skipGetStarted", userMgmtHandler.SkipGetStarted)
	userMgmtRoutes.GET("/getFilterDetails", userMgmtHandler.GetFilterDetails)
	userMgmtRoutes.PUT("/changePassword", userMgmtHandler.ChangePassword)
	userMgmtRoutes.POST("/sendTrace", userMgmtHandler.SendTrace)
	server.AddUsrMgmtCloudRoutes(userMgmtRoutes, userMgmtHandler)
}
