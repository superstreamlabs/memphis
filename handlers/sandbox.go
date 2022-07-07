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
	"bytes"
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
	"net/url"
	"strings"
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
	Picture       string `json: "picture"`
	jwt.StandardClaims
}

type googleOauthToken struct {
	Access_token string
	Id_token     string
}

type githubAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

var sandboxUsersCollection *mongo.Collection = db.GetCollection("sandbox_users")

func (sbh SandboxHandler) Login(c *gin.Context) {
	var body models.SandboxLoginSchema
	var firstName string
	var lastName string
	var email string
	var profilePic string
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	token := body.Token
	loginType := body.LoginType
	if loginType == "google" {
		gOuth, err := getGoogleAuthToken(token)
		if err != nil {
			logger.Error("Login(Sandbox) error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		claims, err := GetGoogleUser(*gOuth)
		if err != nil {
			logger.Error("Login(Sandbox) error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		firstName = claims.FirstName
		lastName = claims.LastName
		email = claims.Email
		profilePic = claims.Picture
	} else if loginType == "github" {
		gitAccessToken, err := getGithubAccessToken(token)
		if err != nil {
			logger.Error("Login(Sandbox) error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		claims, err := getGithubData(gitAccessToken)
		if err != nil {
			logger.Error("Login(Sandbox) error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		email, _ = claims["email"].(string)
		if email == "" {
			temp, _ := claims["repos_url"].(string)
			temp2 := strings.Split(temp, "https://api.github.com/users/")
			temp3 := strings.Split(temp2[1], "/")
			email = temp3[0]
		}
		fullName := strings.Split(claims["name"].(string), " ")
		firstName = fullName[0]
		lastName = fullName[1]
		profilePic = claims["avatar_url"].(string)
	} else {
		logger.Error("Wrong login type")
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	exist, user, err := isSandboxUserExist(email)
	if err != nil {
		logger.Error("Login(Sandbox) error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !exist {
		user = models.SandboxUser{
			ID:              primitive.NewObjectID(),
			Username:        email,
			Password:        "",
			FirstName:       firstName,
			LastName:        lastName,
			HubUsername:     "",
			HubPassword:     "",
			UserType:        "",
			CreationDate:    time.Now(),
			AlreadyLoggedIn: false,
			AvatarId:        1,
			ProfilePic:      profilePic,
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
		"profile_pic":       profilePic,
	})
}

func getGoogleAuthToken(code string) (*googleOauthToken, error) {
	const googleTokenURl = "https://oauth2.googleapis.com/token"

	values := url.Values{}
	decodedValue, err := url.QueryUnescape(code)
	if err != nil {
		decodedValue = code
	}

	values.Add("grant_type", "authorization_code")
	values.Add("code", decodedValue)
	values.Add("client_id", configuration.GOOGLE_CLIENT_ID)
	values.Add("client_secret", configuration.GOOGLE_CLIENT_SECRET)
	values.Add("redirect_uri", configuration.SANDBOX_REDIRECT_URI)

	query := values.Encode()

	req, err := http.NewRequest("POST", googleTokenURl, bytes.NewBufferString(query))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := http.Client{
		Timeout: time.Second * 30,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New("could not retrieve token")
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var GoogleOauthTokenRes map[string]interface{}

	if err := json.Unmarshal(resBody, &GoogleOauthTokenRes); err != nil {
		return nil, err
	}

	tokenBody := &googleOauthToken{
		Access_token: GoogleOauthTokenRes["access_token"].(string),
		Id_token:     GoogleOauthTokenRes["id_token"].(string),
	}

	return tokenBody, nil
}

func GetGoogleUser(gOauthToken googleOauthToken) (*googleClaims, error) {
	googleTokenURl := fmt.Sprintf("https://www.googleapis.com/oauth2/v1/userinfo?alt=json&access_token=%s", gOauthToken.Access_token)
	req, err := http.NewRequest("GET", googleTokenURl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gOauthToken.Id_token))

	client := http.Client{
		Timeout: time.Second * 30,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("could not retrieve user")
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var GoogleUserRes map[string]interface{}

	if err := json.Unmarshal(resBody, &GoogleUserRes); err != nil {
		return nil, err
	}

	claims := &googleClaims{
		Email:         GoogleUserRes["email"].(string),
		EmailVerified: GoogleUserRes["verified_email"].(bool),
		FirstName:     GoogleUserRes["given_name"].(string),
		LastName:      GoogleUserRes["family_name"].(string),
		Picture:       GoogleUserRes["picture"].(string),
	}

	return claims, nil
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

func getGithubAccessToken(code string) (string, error) {

	requestBodyMap := map[string]string{
		"client_id":     configuration.GITHUB_CLIENT_ID,
		"client_secret": configuration.GITHUB_CLIENT_SECRET,
		"code":          code,
	}
	requestJSON, _ := json.Marshal(requestBodyMap)

	req, reqerr := http.NewRequest(
		"POST",
		"https://github.com/login/oauth/access_token",
		bytes.NewBuffer(requestJSON),
	)
	if reqerr != nil {
		return "", reqerr
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		return "", resperr
	}

	respbody, _ := ioutil.ReadAll(resp.Body)

	var ghresp githubAccessTokenResponse
	json.Unmarshal(respbody, &ghresp)

	return ghresp.AccessToken, nil
}

func getGithubData(accessToken string) (map[string]any, error) {

	req, reqerr := http.NewRequest(
		"GET",
		"https://api.github.com/user",
		nil,
	)
	if reqerr != nil {
		return nil, reqerr
	}

	authorizationHeaderValue := fmt.Sprintf("token %s", accessToken)
	req.Header.Set("Authorization", authorizationHeaderValue)

	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		return nil, resperr
	}

	respbody, _ := ioutil.ReadAll(resp.Body)

	data := make(map[string]any)
	err := json.Unmarshal(respbody, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func DenyForSandboxEnv(c *gin.Context) error {
	if configuration.SANDBOX_ENV == "true" {
		c.AbortWithStatusJSON(666, gin.H{"message": "You are in a sandbox environment, this operation is not allowed"})
		return errors.New("Sandbox environment")
	}
	return nil
}
