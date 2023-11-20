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
package middlewares

import (
	"errors"
	"fmt"

	"github.com/memphisdev/memphis/conf"
	"github.com/memphisdev/memphis/memphis_cache"
	"github.com/memphisdev/memphis/models"

	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var noNeedAuthRoutes = []string{
	"/api/usermgmt/login",
	"/api/usermgmt/refreshtoken",
	"/api/usermgmt/addusersignup",
	"/api/usermgmt/getsignupflag",
	"/api/status",
	"/api/monitoring/getclusterinfo",
	"/api/tenants/createtenant",
	"/api/usermgmt/approveinvitation",
}

var refreshTokenRoute string = "/api/usermgmt/refreshtoken"

var configuration = conf.GetConfig()

func isAuthNeeded(path string) bool {
	for _, route := range noNeedAuthRoutes {
		if route == path {
			return false
		}
	}

	return true
}

func extractToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("unsupported auth header")
	}

	splited := strings.Split(authHeader, " ")
	if len(splited) != 2 {
		return "", errors.New("unsupported auth header")
	}

	tokenString := splited[1]
	return tokenString, nil
}

func verifyToken(tokenString string, secret string) (models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("verifyToken: unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return models.User{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok && !token.Valid {
		return models.User{}, err
	}

	if claims["tenant_name"] == nil {
		claims["tenant_name"] = conf.MemphisGlobalAccountName
	}

	userId := int(claims["user_id"].(float64))
	creationDate, _ := time.Parse("2006-01-02T15:04:05.000Z", claims["creation_date"].(string))
	user := models.User{
		ID:              userId,
		Username:        claims["username"].(string),
		UserType:        claims["user_type"].(string),
		CreatedAt:       creationDate,
		AlreadyLoggedIn: claims["already_logged_in"].(bool),
		AvatarId:        int(claims["avatar_id"].(float64)),
		TenantName:      claims["tenant_name"].(string),
	}

	return user, nil
}

func Authenticate(c *gin.Context) {
	path := strings.ToLower(c.Request.URL.Path)
	needToAuthenticate := isAuthNeeded(path)
	var tokenString string
	var err error
	var user models.User
	shouldCheckUser := false
	if needToAuthenticate {
		tokenString, err = extractToken(c.GetHeader("authorization"))
		if err != nil || tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
			return
		}

		user, err = verifyToken(tokenString, configuration.JWT_SECRET)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
			return
		}

		shouldCheckUser = true
	} else if path == refreshTokenRoute {
		tokenString, err = c.Cookie("memphis-jwt-refresh-token")
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
			return
		}

		user, err = verifyToken(tokenString, configuration.REFRESH_JWT_SECRET)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
			return
		}

		shouldCheckUser = true
	}

	if shouldCheckUser {
		username := strings.ToLower(user.Username)
		if user.TenantName != conf.GlobalAccount {
			user.TenantName = strings.ToLower(user.TenantName)
		}

		exists, _, err := memphis_cache.GetUser(username, user.TenantName, false)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exists {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
			return
		}
	}

	c.Set("user", user)
	c.Next()
}
