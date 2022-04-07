package middlewares

import (
	"context"
	"errors"
	"fmt"
	"memphis-server/config"
	"memphis-server/db"
	"memphis-server/logger"
	"memphis-server/models"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var noNeedAuthRoutes = []string{
	"/api-gw/usermgmt/login",
	"/api-gw/usermgmt/refreshtoken",
	"/api-gw/status",
}

var refreshTokenRoute string = "/api-gw/usermgmt/refreshtoken"

var configuration = config.GetConfig()
var tokensCollection *mongo.Collection = db.GetCollection("tokens")

func isAuthNeeded(path string) bool {
	if strings.HasPrefix(path, "/api-gw/usermgmt/nats/authenticate") {
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
		tokenString, err := extractToken(c.GetHeader("authorization"))
		if err != nil {
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
