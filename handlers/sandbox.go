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

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"memphis-broker/logger"
	"memphis-broker/models"
	"memphis-broker/utils"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SandboxHandler struct{}

func (sbh SandboxHandler) Login(c *gin.Context) {
	var body models.SandboxLoginSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	google_token := body.Google_token
	claims, err := ValidateGoogleJWT(google_token)
	if err != nil {
		logger.Error("ValidateGoogleJWT: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	exist, user, err := IsUserExist(claims.Email)
	if err != nil {
		logger.Error("IsUserExist: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !exist {
		user = models.User{
			ID:              primitive.NewObjectID(),
			Username:        claims.Email,
			Password:        claims.FirstName + "." + claims.LastName,
			HubUsername:     "",
			HubPassword:     "",
			UserType:        "management",
			CreationDate:    time.Now(),
			AlreadyLoggedIn: false,
			AvatarId:        1,
		}
		_, err = usersCollection.InsertOne(context.TODO(), user)
		if err != nil {
			logger.Error("InsertOne: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

	}

	token, refreshToken, err := createTokens(user)
	if err != nil {
		logger.Error("createTokens: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	// authenticated, user, err := authenticateUser(claims.Email, claims.FirstName+"."+claims.LastName)
	// if err != nil {
	// 	logger.Error("Login error: " + err.Error())
	// 	c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
	// 	return
	// }
	// if !authenticated {
	// 	c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	// 	return
	// }

	opts := options.Update().SetUpsert(true)
	_, err = tokensCollection.UpdateOne(context.TODO(),
		bson.M{"username": user.Username},
		bson.M{"$set": bson.M{"jwt_token": token, "refresh_token": refreshToken}},
		opts,
	)
	if err != nil {
		logger.Error("Login error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !user.AlreadyLoggedIn {
		usersCollection.UpdateOne(context.TODO(),
			bson.M{"_id": user.ID},
			bson.M{"$set": bson.M{"already_logged_in": true}},
		)
	}
	domain := ""
	secure := false
	c.SetCookie("jwt-refresh-token", refreshToken, configuration.REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
	c.IndentedJSON(200, gin.H{
		"jwt":               token,
		"expires_in":        configuration.JWT_EXPIRES_IN_MINUTES * 60 * 1000,
		"user_id":           user.ID,
		"username":          user.Username,
		"user_type":         user.UserType,
		"creation_date":     user.CreationDate,
		"already_logged_in": user.AlreadyLoggedIn,
		"avatar_id":         user.AvatarId,
	})
}

type GoogleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
	jwt.StandardClaims
}

func getGooglePublicKey(keyID string) (string, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v1/certs")
	if err != nil {
		return "", err
	}
	dat, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	myResp := map[string]string{}
	err = json.Unmarshal(dat, &myResp)
	if err != nil {
		return "", err
	}
	key, ok := myResp[keyID]
	if !ok {
		return "", errors.New("key not found")
	}
	return key, nil
}

func ValidateGoogleJWT(tokenString string) (GoogleClaims, error) {
	claimsStruct := GoogleClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) {
			pem, err := getGooglePublicKey(fmt.Sprintf("%s", token.Header["kid"]))
			if err != nil {
				return nil, err
			}
			key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pem))
			if err != nil {
				return nil, err
			}
			return key, nil
		},
	)
	if err != nil {
		return GoogleClaims{}, err
	}

	claims, ok := token.Claims.(*GoogleClaims)
	if !ok {
		return GoogleClaims{}, errors.New("invalid Google JWT")
	}

	if claims.Issuer != "accounts.google.com" && claims.Issuer != "https://accounts.google.com" {
		return GoogleClaims{}, errors.New("iss is invalid")
	}

	if claims.Audience != configuration.GOOGLE_CLIENT_ID {
		return GoogleClaims{}, errors.New("aud is invalid")
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return GoogleClaims{}, errors.New("JWT is expired")
	}

	return *claims, nil
}
