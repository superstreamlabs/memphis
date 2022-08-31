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
	"/api/status",
	"/api/sandbox/login",
}

var refreshTokenRoute string = "/api/usermgmt/refreshtoken"

var configuration = conf.GetConfig()

func isAuthNeeded(path string) bool {
	if strings.HasPrefix(path, "/api/usermgmt/nats/authenticate") {
		return false
	}

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
		if strings.Contains(path, "socket.io") {
			tokenString = c.Request.URL.Query().Get("authorization")
		} else {
			tokenString, err = extractToken(c.GetHeader("authorization"))
		}

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
