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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"memphis/analytics"
	"memphis/db"
	"memphis/models"
	"memphis/utils"
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
	Picture       string `json:"picture"`
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
			serv.Errorf("Login(Sandbox) with Google: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		claims, err := GetGoogleUser(*gOuth)
		if err != nil {
			serv.Errorf("Login(Sandbox) with Google: " + err.Error())
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
			serv.Errorf("Login(Sandbox) with GitHub: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		claims, err := getGithubData(gitAccessToken)
		if err != nil {
			serv.Errorf("Login(Sandbox) with GitHub: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		email, ok = claims["email"].(string)
		if email == "" || !ok {
			temp, _ := claims["repos_url"].(string)
			temp2 := strings.Split(temp, "https://api.github.com/users/")
			temp3 := strings.Split(temp2[1], "/")
			email = temp3[0]
		}
		name, ok := claims["name"].(string)
		if name == "" || !ok {
			firstName = ""
			lastName = ""
		} else {
			fullName := strings.Split(name, " ")
			firstName = fullName[0]
			lastName = fullName[1]
		}
		profilePic = claims["avatar_url"].(string)
	} else {
		serv.Errorf("Login(Sandbox): Wrong login type")
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	var username string
	if strings.Contains(email, "@") {
		username = email[:strings.IndexByte(email, '@')]
	} else {
		username = email
	}
	exist, user, err := IsSandboxUserExist(username)
	if err != nil {
		serv.Errorf("Login(Sandbox): With user " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !exist {
		user = models.SandboxUser{
			ID:              primitive.NewObjectID(),
			Username:        username,
			Email:           email,
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

		if !strings.Contains(email, "@") {
			email = email + "@github.memphis"
		}

		param := analytics.EventParam{
			Name:  "email",
			Value: email,
		}
		analyticsParams := []analytics.EventParam{param}
		analytics.SendEventWithParams("", analyticsParams, "new-sandbox-user")

		var sandboxUsersCollection *mongo.Collection = db.GetCollection("sandbox_users", serv.memphis.dbClient)
		_, err = sandboxUsersCollection.InsertOne(context.TODO(), user)
		if err != nil {
			serv.Errorf("Login(Sandbox): With user " + username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		serv.Noticef("New sandbox user was created: " + username)
	}

	token, refreshToken, err := CreateTokens(user)
	if err != nil {
		serv.Errorf("Login(Sandbox): With user " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !user.AlreadyLoggedIn {
		sandboxUsersCollection.UpdateOne(context.TODO(),
			bson.M{"_id": user.ID},
			bson.M{"$set": bson.M{"already_logged_in": true}},
		)
	}
	serv.Noticef("Sandbox user logged in: " + username)
	domain := ""
	secure := false
	c.SetCookie("jwt-refresh-token", refreshToken, configuration.REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-sandbox-login")
	}

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
		"env":               "K8S",
		"skip_get_started":  user.SkipGetStarted,
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

func IsSandboxUserExist(username string) (bool, models.SandboxUser, error) {
	filter := bson.M{"username": username}
	var user models.SandboxUser
	var sandboxUsersCollection *mongo.Collection = db.GetCollection("sandbox_users", serv.memphis.dbClient)
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
	if err := json.Unmarshal(respbody, &ghresp); err != nil {
		return "", err
	}

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
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("DenyForSandboxEnv: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return err
	}

	if configuration.SANDBOX_ENV == "true" && user.UserType != "root" {
		sandboxErrCode := 665
		c.AbortWithStatusJSON(sandboxErrCode, gin.H{"message": "You are in a sandbox environment, this operation is not allowed"})
		return errors.New("Sandbox environment")
	}
	return nil
}
