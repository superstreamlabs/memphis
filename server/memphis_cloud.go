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
package server

import (
	"memphis/analytics"
	"memphis/conf"
	"memphis/db"
	"memphis/models"
	"memphis/utils"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type CloudHandler struct{ S *Server }

// routes
func InitializeTenantsRoutes(router *gin.RouterGroup, h *Handlers) {
	tenantsHandler := h.Tenants
	tenantsRoutes := router.Group("/tenants")
	config := SetCors()
	tenantsRoutes.POST("/createTenant", cors.New(config), tenantsHandler.CreateTenant)
}

func SetCors() cors.Config {
	config := cors.Config{}
	return config
}

func validateTenantName(tenantName string) error {
	return nil
}

func (cl CloudHandler) CreateTenant(c *gin.Context) {
	c.IndentedJSON(404, gin.H{})
}

func (cl CloudHandler) Login(c *gin.Context) {
	var body models.LoginSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	username := strings.ToLower(body.Username)
	authenticated, user, err := authenticateUser(username, body.Password)
	if err != nil {
		serv.Errorf("Login : User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !authenticated || user.UserType == "application" {
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	token, refreshToken, err := CreateTokens(user)
	if err != nil {
		serv.Errorf("Login: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !user.AlreadyLoggedIn {
		err = db.UpdateUserAlreadyLoggedIn(user.ID)
		if err != nil {
			serv.Errorf("Login: User " + body.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-login")
	}

	env := "K8S"
	if configuration.DOCKER_ENV != "" || configuration.LOCAL_CLUSTER_ENV {
		env = "docker"
	}
	exist, tenant, err := db.GetTenantByName(user.TenantName)
	if err != nil {
		serv.Errorf("Login: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("Login: User " + body.Username + ": tenant " + user.TenantName + " does not exist")
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	domain := ""
	secure := false
	c.SetCookie("jwt-refresh-token", refreshToken, REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
	c.IndentedJSON(200, gin.H{
		"jwt":                     token,
		"expires_in":              JWT_EXPIRES_IN_MINUTES * 60 * 1000,
		"user_id":                 user.ID,
		"username":                user.Username,
		"user_type":               user.UserType,
		"created_at":              user.CreatedAt,
		"already_logged_in":       user.AlreadyLoggedIn,
		"avatar_id":               user.AvatarId,
		"send_analytics":          shouldSendAnalytics,
		"env":                     env,
		"full_name":               user.FullName,
		"skip_get_started":        user.SkipGetStarted,
		"broker_host":             serv.opts.BrokerHost,
		"rest_gw_host":            serv.opts.RestGwHost,
		"ui_host":                 serv.opts.UiHost,
		"tiered_storage_time_sec": serv.opts.TieredStorageUploadIntervalSec,
		"ws_port":                 serv.opts.Websocket.Port,
		"http_port":               serv.opts.UiPort,
		"clients_port":            serv.opts.Port,
		"rest_gw_port":            serv.opts.RestGwPort,
		"user_pass_based_auth":    configuration.USER_PASS_BASED_AUTH,
		"connection_token":        configuration.CONNECTION_TOKEN,
		"account_id":              tenant.ID,
	})
}

func (cl CloudHandler) AddUser(c *gin.Context) {
	var body models.AddUserSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	var subscription, pending bool
	team := strings.ToLower(body.Team)
	position := strings.ToLower(body.Position)
	fullName := strings.ToLower(body.FullName)

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("AddUser: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	if user.TenantName != conf.GlobalAccountName {
		user.TenantName = strings.ToLower(user.TenantName)
	}

	username := strings.ToLower(body.Username)
	usernameError := validateUsername(username)
	if usernameError != nil {
		serv.Warnf("AddUser: " + usernameError.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": usernameError.Error()})
		return
	}
	exist, _, err := db.GetUserByUsername(username, user.TenantName)
	if err != nil {
		serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		errMsg := "A user with the name " + body.Username + " already exists"
		serv.Warnf("CreateUser: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	userType := strings.ToLower(body.UserType)
	userTypeError := validateUserType(userType)
	if userTypeError != nil {
		serv.Warnf("AddUser: " + userTypeError.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": userTypeError.Error()})
		return
	}

	var password string
	var avatarId int
	if userType == "management" {
		if body.Password == "" {
			serv.Warnf("AddUser: Password was not provided for user " + username)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Password was not provided"})
			return
		}

		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
		if err != nil {
			serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		password = string(hashedPwd)

		avatarId = 1
		if body.AvatarId > 0 {
			avatarId = body.AvatarId
		}
	}

	var brokerConnectionCreds string
	if userType == "application" {
		fullName = ""
		subscription = false
		pending = false
		if configuration.USER_PASS_BASED_AUTH {
			if body.Password == "" {
				serv.Warnf("AddUser: Password was not provided for user " + username)
				c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Password was not provided"})
				return
			}
			password, err = EncryptAES([]byte(body.Password))
			if err != nil {
				serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			avatarId = 1
			if body.AvatarId > 0 {
				avatarId = body.AvatarId
			}
		} else {
			brokerConnectionCreds = configuration.CONNECTION_TOKEN
		}
	}
	newUser, err := db.CreateUser(username, userType, password, fullName, subscription, avatarId, user.TenantName, pending, team, position)
	if err != nil {
		if strings.Contains(err.Error(), "already exist") {
			serv.Warnf("CreateUserManagement: " + err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
		serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-add-user")
	}

	if userType == "application" && configuration.USER_PASS_BASED_AUTH {
		// send signal to reload config
		err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), CONFIGURATIONS_RELOAD_SIGNAL_SUBJ, _EMPTY_, nil, _EMPTY_, true)
		if err != nil {
			serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	serv.Noticef("User " + username + " has been created")
	c.IndentedJSON(200, gin.H{
		"id":                      newUser.ID,
		"username":                username,
		"user_type":               userType,
		"created_at":              newUser.CreatedAt,
		"already_logged_in":       false,
		"avatar_id":               body.AvatarId,
		"broker_connection_creds": brokerConnectionCreds,
		"position":                newUser.Position,
		"team":                    newUser.Team,
		"pending":                 newUser.Pending,
	})
}

func InitializeApprovedInvitation(router *gin.RouterGroup) {
}
