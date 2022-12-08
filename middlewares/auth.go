// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package middlewares
package middlewares

import (
	"errors"
	"fmt"

	"memphis-broker/conf"
	"memphis-broker/models"

	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var noNeedAuthRoutes = []string{
	"/api/usermgmt/login",
	"/api/usermgmt/refreshtoken",
	"/api/usermgmt/addusersignup",
	"/api/usermgmt/getsignupflag",
	"/api/status",
	"/api/sandbox/login",
	"/api/monitoring/getclusterinfo",
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
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return models.User{}, errors.New("f")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok && !token.Valid {
		return models.User{}, errors.New("f")
	}

	userId, _ := primitive.ObjectIDFromHex(claims["user_id"].(string))
	creationDate, _ := time.Parse("2006-01-02T15:04:05.000Z", claims["creation_date"].(string))
	user := models.User{
		ID:              userId,
		Username:        claims["username"].(string),
		UserType:        claims["user_type"].(string),
		CreationDate:    creationDate,
		AlreadyLoggedIn: claims["already_logged_in"].(bool),
		AvatarId:        int(claims["avatar_id"].(float64)),
	}

	return user, nil
}

func Authenticate(c *gin.Context) {
	path := strings.ToLower(c.Request.URL.Path)
	needToAuthenticate := isAuthNeeded(path)
	if needToAuthenticate {
		var tokenString string
		var err error
		tokenString, err = extractToken(c.GetHeader("authorization"))

		if err != nil || tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
			return
		}

		user, err := verifyToken(tokenString, configuration.JWT_SECRET)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
			return
		}

		c.Set("user", user)
	} else if path == refreshTokenRoute {
		tokenString, err := c.Cookie("jwt-refresh-token")
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
			return
		}

		user, err := verifyToken(tokenString, configuration.REFRESH_JWT_SECRET)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
			return
		}

		c.Set("user", user)
	}

	c.Next()
}
