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
package routes

import (
	"memphis-broker/server"
	"github.com/gin-gonic/gin"
)

func InitializeUserMgmtRoutes(router *gin.RouterGroup) {
	userMgmtHandler := server.UserMgmtHandler{}
	userMgmtRoutes := router.Group("/usermgmt")
	userMgmtRoutes.GET("/nats/authenticate", userMgmtHandler.AuthenticateNatsUser)
	userMgmtRoutes.GET("/nats/authenticate/:publicKey", userMgmtHandler.AuthenticateNatsUser)
	userMgmtRoutes.POST("/login", userMgmtHandler.Login)
	userMgmtRoutes.POST("/doneNextSteps", userMgmtHandler.DoneNextSteps)
	userMgmtRoutes.POST("/refreshToken", userMgmtHandler.RefreshToken)
	userMgmtRoutes.POST("/addUser", userMgmtHandler.AddUser)
	userMgmtRoutes.GET("/getAllUsers", userMgmtHandler.GetAllUsers)
	userMgmtRoutes.DELETE("/removeUser", userMgmtHandler.RemoveUser)
	userMgmtRoutes.DELETE("/removeMyUser", userMgmtHandler.RemoveMyUser)
	userMgmtRoutes.PUT("/editAvatar", userMgmtHandler.EditAvatar)
	userMgmtRoutes.PUT("/editHubCreds", userMgmtHandler.EditHubCreds)
	userMgmtRoutes.PUT("/editCompanyLogo", userMgmtHandler.EditCompanyLogo)
	userMgmtRoutes.DELETE("/removeCompanyLogo", userMgmtHandler.RemoveCompanyLogo)
	userMgmtRoutes.GET("/getCompanyLogo", userMgmtHandler.GetCompanyLogo)
	userMgmtRoutes.PUT("/editAnalytics", userMgmtHandler.EditAnalytics)
}
