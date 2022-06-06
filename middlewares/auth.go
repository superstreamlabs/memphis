// Copyright 2021-2022 The Memphis Authors
// Licensed under the GNU General Public License v3.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middlewares

import (
	"context"
	"errors"
	"fmt"

	"memphis-control-plane/config"
	"memphis-control-plane/db"
	"memphis-control-plane/logger"
	"memphis-control-plane/models"

	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var noNeedAuthRoutes = []string{
	"/api/usermgmt/login",
	"/api/usermgmt/refreshtoken",
	"/api/status",
}

var refreshTokenRoute string = "/api/usermgmt/refreshtoken"

var configuration = config.GetConfig()
var tokensCollection *mongo.Collection = db.GetCollection("tokens")

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

func isRefreshTokenExist(username string, tokenString string) (bool, error) {
	filter := bson.M{"username": username, "refresh_token": tokenString}
	var token models.Token
	err := tokensCollection.FindOne(context.TODO(), filter).Decode(&token)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func isTokenExist(username string, tokenString string) (bool, error) {
	filter := bson.M{"username": username, "jwt_token": tokenString}
	var token models.Token
	err := tokensCollection.FindOne(context.TODO(), filter).Decode(&token)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
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

		exist, err := isTokenExist(user.Username, tokenString)
		if err != nil {
			logger.Error("Authenticate error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
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

		exist, err := isRefreshTokenExist(user.Username, tokenString)
		if err != nil {
			logger.Error("Authenticate error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
			return
		}

		c.Set("user", user)
	}

	c.Next()
}
