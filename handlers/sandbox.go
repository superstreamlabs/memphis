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
	"memphis-broker/db"
	"memphis-broker/logger"
	"memphis-broker/models"
	"memphis-broker/utils"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SandboxHandler struct{}

type googleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
	jwt.StandardClaims
}

var sandboxUsersCollection *mongo.Collection = db.GetCollection("sandbox_users")

func (sbh SandboxHandler) Login(c *gin.Context) {
	var body models.SandboxLoginSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	google_token := body.Google_token
	claims, err := validateGoogleJWT(google_token)
	if err != nil {
		logger.Error("Login(Sandbox) error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	exist, user, err := isSandboxUserExist(claims.Email)
	if err != nil {
		logger.Error("Login(Sandbox) error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !exist {
		user = models.SandboxUser{
			ID:              primitive.NewObjectID(),
			Username:        claims.Email,
			Password:        claims.FirstName + "." + claims.LastName,
			FirstName:       claims.FirstName,
			LastName:        claims.LastName,
			HubUsername:     "",
			HubPassword:     "",
			UserType:        "management",
			CreationDate:    time.Now(),
			AlreadyLoggedIn: false,
			AvatarId:        1,
		}
		_, err = sandboxUsersCollection.InsertOne(context.TODO(), user)
		if err != nil {
			logger.Error("Login(Sandbox) error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

	}

	token, refreshToken, err := CreateTokens(user)
	if err != nil {
		logger.Error("Login(Sandbox) error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !user.AlreadyLoggedIn {
		sandboxUsersCollection.UpdateOne(context.TODO(),
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

func validateGoogleJWT(tokenString string) (googleClaims, error) {
	claimsStruct := googleClaims{}

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
		return googleClaims{}, err
	}

	claims, ok := token.Claims.(*googleClaims)
	if !ok {
		return googleClaims{}, errors.New("invalid Google JWT")
	}

	if claims.Issuer != "accounts.google.com" && claims.Issuer != "https://accounts.google.com" {
		return googleClaims{}, errors.New("iss is invalid")
	}

	if claims.Audience != configuration.GOOGLE_CLIENT_ID {
		return googleClaims{}, errors.New("aud is invalid")
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return googleClaims{}, errors.New("JWT is expired")
	}

	return *claims, nil
}

func isSandboxUserExist(username string) (bool, models.SandboxUser, error) {
	filter := bson.M{"username": username}
	var user models.SandboxUser
	err := sandboxUsersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return false, user, nil
	} else if err != nil {
		return false, user, err
	}
	return true, user, nil
}
