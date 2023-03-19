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
	"encoding/base64"
	"errors"
	"memphis/analytics"
	"memphis/db"
	"memphis/models"
	"memphis/utils"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserMgmtHandler struct{}

func isRootUserExist() (bool, error) {
	exist, _, err := db.GetRootUser()
	if !exist {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func isRootUserLoggedIn() (bool, error) {
	exist, user, err := db.GetRootUser()
	if !exist {
		return false, errors.New("Root user does not exist")
	} else if err != nil {
		return false, err
	}

	if user.AlreadyLoggedIn {
		return true, nil
	} else {
		return false, nil
	}
}

func authenticateUser(username string, password string) (bool, models.User, error) {
	exist, user, err := db.GetUserByUsername(username)
	if !exist {
		return false, models.User{}, nil
	} else if err != nil {
		return false, models.User{}, err
	}

	hashedPwd := []byte(user.Password)
	err = bcrypt.CompareHashAndPassword(hashedPwd, []byte(password))
	if err != nil {
		return false, models.User{}, nil
	}

	return true, user, nil
}

func validateUserType(userType string) error {
	if userType != "application" && userType != "management" {
		return errors.New("user type has to be application/management")
	}
	return nil
}

func updateDeletedUserResources(user models.User) error {
	if user.UserType == "application" {
		err := RemoveUser(user.Username)
		if err != nil {
			return err
		}
	}

	err := db.UpdateStationsOfDeletedUser(user.ID)
	if err != nil {
		return err
	}

	err = db.UpdateConncetionsOfDeletedUser(user.ID)
	if err != nil {
		return err
	}

	err = db.UpdateProducersOfDeletedUser(user.ID)
	if err != nil {
		return err
	}

	err = db.UpdateConsumersOfDeletedUser(user.ID)
	if err != nil {
		return err
	}

	err = db.UpdateSchemasOfDeletedUser(user.ID)
	if err != nil {
		return err
	}

	err = db.UpdateAuditLogsOfDeletedUser(user.ID)
	if err != nil {
		return err
	}

	return nil
}

func validateUsername(username string) error {
	re := regexp.MustCompile("^[a-z0-9_.]*$")

	validName := re.MatchString(username)
	if !validName || len(username) == 0 {
		return errors.New("username has to include only letters/numbers/./_ ")
	}
	return nil
}

func validateEmail(email string) error {
	re := regexp.MustCompile("^[a-z0-9._%+-]+@[a-z0-9_.-]+.[a-z]{2,4}$")
	validateEmail := re.MatchString(email)
	if !validateEmail || len(email) == 0 {
		return errors.New("email is not valid")
	}
	return nil
}

type userToTokens interface {
	models.User | models.SandboxUser
}

func CreateTokens[U userToTokens](user U) (string, string, error) {
	atClaims := jwt.MapClaims{}
	var at *jwt.Token
	switch u := any(user).(type) {
	case models.User:
		atClaims["user_id"] = u.ID
		atClaims["username"] = u.Username
		atClaims["user_type"] = u.UserType
		atClaims["creation_date"] = u.CreatedAt
		atClaims["already_logged_in"] = u.AlreadyLoggedIn
		atClaims["avatar_id"] = u.AvatarId
		atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(configuration.JWT_EXPIRES_IN_MINUTES)).Unix()
		at = jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	case models.SandboxUser:
		atClaims["user_id"] = u.ID
		atClaims["username"] = u.Username
		atClaims["user_type"] = u.UserType
		atClaims["creation_date"] = u.CreatedAt
		atClaims["already_logged_in"] = u.AlreadyLoggedIn
		atClaims["avatar_id"] = u.AvatarId
		atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(configuration.JWT_EXPIRES_IN_MINUTES)).Unix()
		at = jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	}
	token, err := at.SignedString([]byte(configuration.JWT_SECRET))
	if err != nil {
		return "", "", err
	}

	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(configuration.REFRESH_JWT_EXPIRES_IN_MINUTES)).Unix()

	at = jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	refreshToken, err := at.SignedString([]byte(configuration.REFRESH_JWT_SECRET))
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

func imageToBase64(imagePath string) (string, error) {
	bytes, err := os.ReadFile(imagePath)
	if err != nil {
		return "", err
	}

	fileExt := filepath.Ext(imagePath)
	var base64Encoding string

	switch fileExt {
	case ".jpeg":
		base64Encoding += "data:image/jpeg;base64,"
	case ".png":
		base64Encoding += "data:image/png;base64,"
	case ".jpg":
		base64Encoding += "data:image/jpg;base64,"
	}

	base64Encoding += base64.StdEncoding.EncodeToString(bytes)
	return base64Encoding, nil
}

func CreateRootUserOnFirstSystemLoad() error {
	exist, err := isRootUserExist()
	if err != nil {
		return err
	}
	password := configuration.ROOT_PASSWORD
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return err
	}
	hashedPwdString := string(hashedPwd)

	if !exist {
		_, err = db.CreateUser("root", "root", hashedPwdString, "", false, 1)
		if err != nil {
			return err
		}

		if configuration.ANALYTICS == "true" {
			installationType := "stand-alone-k8s"
			if serv.JetStreamIsClustered() {
				installationType = "cluster"
			} else if configuration.DOCKER_ENV == "true" {
				installationType = "stand-alone-docker"
			}

			param := analytics.EventParam{
				Name:  "installation-type",
				Value: installationType,
			}
			analyticsParams := []analytics.EventParam{param}
			analytics.SendEventWithParams("", analyticsParams, "installation")

			if configuration.EXPORTER {
				analytics.SendEventWithParams("", analyticsParams, "enable-exporter")
			}
		}
	} else {
		err = db.ChangeUserPassword("root", hashedPwdString)
		if err != nil {
			return err
		}
	}

	return nil
}

func (umh UserMgmtHandler) ChangePassword(c *gin.Context) {
	var body models.ChangePasswordSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	username := strings.ToLower(body.Username)
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("EditPassword: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if username == "root" && user.UserType != "root" {
		errMsg := "Change root password: This operation can be done only by the root user"
		serv.Warnf("EditPassword: " + errMsg)
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	} else if username != strings.ToLower(user.Username) && strings.ToLower(user.Username) != "root" {
		errMsg := "Change user password: This operation can be done only by the user or the root user"
		serv.Warnf("EditPassword: " + errMsg)
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
	if err != nil {
		serv.Errorf("EditPassword: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	hashedPwdString := string(hashedPwd)
	err = db.ChangeUserPassword(username, hashedPwdString)
	if err != nil {
		serv.Errorf("EditPassword: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) Login(c *gin.Context) {
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
		db.UpdateUserAlreadyLoggedIn(user.ID)
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-login")
	}

	brokerHost := BROKER_HOST
	restGWHost := REST_GW_HOST
	uiHost := UI_HOST
	var env string
	if configuration.DOCKER_ENV != "" || configuration.LOCAL_CLUSTER_ENV {
		env = "docker"
	} else {
		env = "K8S"
		if BROKER_HOST == "" {
			brokerHost = "memphis." + configuration.K8S_NAMESPACE + ".svc.cluster.local"
		}
		if UI_HOST == "" {
			uiHost = "memphis." + configuration.K8S_NAMESPACE + ".svc.cluster.local"
		}
		if REST_GW_HOST == "" {
			restGWHost = "http://memphis-rest-gateway." + configuration.K8S_NAMESPACE + ".svc.cluster.local"
		}
	}

	domain := ""
	secure := false
	c.SetCookie("jwt-refresh-token", refreshToken, configuration.REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
	c.IndentedJSON(200, gin.H{
		"jwt":                     token,
		"expires_in":              configuration.JWT_EXPIRES_IN_MINUTES * 60 * 1000,
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
		"broker_host":             brokerHost,
		"rest_gw_host":            restGWHost,
		"ui_host":                 uiHost,
		"tiered_storage_time_sec": TIERED_STORAGE_TIME_FRAME_SEC,
		"ws_port":                 configuration.WS_PORT,
		"http_port":               configuration.HTTP_PORT,
		"clients_port":            configuration.CLIENTS_PORT,
		"rest_gw_port":            configuration.REST_GW_PORT,
	})
}

func (umh UserMgmtHandler) RefreshToken(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("refreshToken: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}
	username := user.Username
	_, systemKey, err := db.GetSystemKey("analytics")
	if err != nil {
		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	sendAnalytics, _ := strconv.ParseBool(systemKey.Value)
	exist, user, err := db.GetUserByUsername(username)
	if err != nil {
		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		// exist, sandboxUser, err := IsSandboxUserExist(username)
		// if exist {
		// 	if err != nil {
		// 		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		// 		return
		// 	}

		// 	token, refreshToken, err := CreateTokens(sandboxUser)
		// 	if err != nil {
		// 		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		// 		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		// 		return
		// 	}
		// 	domain := ""
		// 	secure := true
		// 	c.SetCookie("jwt-refresh-token", refreshToken, configuration.REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
		// 	c.IndentedJSON(200, gin.H{
		// 		"jwt":                     token,
		// 		"expires_in":              configuration.JWT_EXPIRES_IN_MINUTES * 60 * 1000,
		// 		"user_id":                 sandboxUser.ID,
		// 		"username":                sandboxUser.Username,
		// 		"user_type":               sandboxUser.UserType,
		// 		"creation_date":           sandboxUser.CreationDate,
		// 		"already_logged_in":       sandboxUser.AlreadyLoggedIn,
		// 		"avatar_id":               sandboxUser.AvatarId,
		// 		"send_analytics":          true,
		// 		"env":                     "K8S",
		// 		"namespace":               configuration.K8S_NAMESPACE,
		// 		"skip_get_started":        sandboxUser.SkipGetStarted,
		// 		"broker_host":             BROKER_HOST,
		// 		"rest_gw_host":            REST_GW_HOST,
		// 		"ui_host":                 UI_HOST,
		// 		"tiered_storage_time_sec": TIERED_STORAGE_TIME_FRAME_SEC,
		// "ws_port":                 configuration.WS_PORT,
		// 		"http_port":               configuration.HTTP_PORT,
		// 		"clients_port":            configuration.CLIENTS_PORT,
		// 		"rest_gw_port":            configuration.REST_GW_PORT,
		// 	})
		// 	return
		// }
	}

	token, refreshToken, err := CreateTokens(user)
	if err != nil {
		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	brokerHost := BROKER_HOST
	restGWHost := REST_GW_HOST
	uiHost := UI_HOST
	var env string
	if configuration.DOCKER_ENV != "" || configuration.LOCAL_CLUSTER_ENV {
		env = "docker"
	} else {
		env = "K8S"
		if BROKER_HOST == "" {
			brokerHost = "memphis." + configuration.K8S_NAMESPACE + ".svc.cluster.local"
		}
		if UI_HOST == "" {
			uiHost = "memphis." + configuration.K8S_NAMESPACE + ".svc.cluster.local"
		}
		if REST_GW_HOST == "" {
			restGWHost = "http://memphis-rest-gateway." + configuration.K8S_NAMESPACE + ".svc.cluster.local"
		}
	}

	domain := ""
	secure := true
	c.SetCookie("jwt-refresh-token", refreshToken, configuration.REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
	c.IndentedJSON(200, gin.H{
		"jwt":                     token,
		"expires_in":              configuration.JWT_EXPIRES_IN_MINUTES * 60 * 1000,
		"user_id":                 user.ID,
		"username":                user.Username,
		"user_type":               user.UserType,
		"created_at":              user.CreatedAt,
		"already_logged_in":       user.AlreadyLoggedIn,
		"avatar_id":               user.AvatarId,
		"send_analytics":          sendAnalytics,
		"env":                     env,
		"namespace":               configuration.K8S_NAMESPACE,
		"full_name":               user.FullName,
		"skip_get_started":        user.SkipGetStarted,
		"broker_host":             brokerHost,
		"rest_gw_host":            restGWHost,
		"ui_host":                 uiHost,
		"tiered_storage_time_sec": TIERED_STORAGE_TIME_FRAME_SEC,
		"ws_port":                 configuration.WS_PORT,
		"http_port":               configuration.HTTP_PORT,
		"clients_port":            configuration.CLIENTS_PORT,
		"rest_gw_port":            configuration.REST_GW_PORT,
	})
}

func (umh UserMgmtHandler) GetSignUpFlag(c *gin.Context) {
	if configuration.SANDBOX_ENV == "true" {
		c.IndentedJSON(200, gin.H{"show_signup": false})
		return
	}

	loggedIn, err := isRootUserLoggedIn()
	if err != nil {
		serv.Errorf("GetSignUpFlag: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent("", "user-open-ui")
	}
	c.IndentedJSON(200, gin.H{"show_signup": !loggedIn})
}

func (umh UserMgmtHandler) AddUserSignUp(c *gin.Context) {
	var body models.AddUserSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	username := strings.ToLower(body.Username)
	usernameError := validateEmail(username)
	if usernameError != nil {
		serv.Warnf(usernameError.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": usernameError.Error()})
		return
	}
	fullName := strings.ToLower(body.FullName)

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
	if err != nil {
		serv.Errorf("CreateUserSignUp: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	hashedPwdString := string(hashedPwd)
	subscription := body.Subscribtion

	newUser, err := db.CreateUser(username, "management", hashedPwdString, fullName, subscription, 1)
	if err != nil {
		serv.Errorf("CreateUserSignUp error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	serv.Noticef("User " + username + " has been signed up")
	token, refreshToken, err := CreateTokens(newUser)
	if err != nil {
		serv.Errorf("CreateUserSignUp error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	brokerHost := BROKER_HOST
	restGWHost := REST_GW_HOST
	uiHost := UI_HOST
	var env string
	if configuration.DOCKER_ENV != "" || configuration.LOCAL_CLUSTER_ENV {
		env = "docker"
	} else {
		env = "K8S"
		if BROKER_HOST == "" {
			brokerHost = "memphis." + configuration.K8S_NAMESPACE + ".svc.cluster.local"
		}
		if UI_HOST == "" {
			uiHost = "memphis." + configuration.K8S_NAMESPACE + ".svc.cluster.local"
		}
		if REST_GW_HOST == "" {
			restGWHost = "http://memphis-rest-gateway." + configuration.K8S_NAMESPACE + ".svc.cluster.local"
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		param1 := analytics.EventParam{
			Name:  "email",
			Value: username,
		}
		param2 := analytics.EventParam{
			Name:  "newsletter",
			Value: strconv.FormatBool(subscription),
		}
		analyticsParams := []analytics.EventParam{param1, param2}
		analytics.SendEventWithParams(username, analyticsParams, "user-signup")
	}

	domain := ""
	secure := false
	c.SetCookie("jwt-refresh-token", refreshToken, configuration.REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
	c.IndentedJSON(200, gin.H{
		"jwt":                     token,
		"expires_in":              configuration.JWT_EXPIRES_IN_MINUTES * 60 * 1000,
		"user_id":                 newUser.ID,
		"username":                newUser.Username,
		"user_type":               newUser.UserType,
		"created_at":              newUser.CreatedAt,
		"already_logged_in":       newUser.AlreadyLoggedIn,
		"avatar_id":               newUser.AvatarId,
		"send_analytics":          shouldSendAnalytics,
		"env":                     env,
		"namespace":               configuration.K8S_NAMESPACE,
		"full_name":               newUser.FullName,
		"skip_get_started":        newUser.SkipGetStarted,
		"broker_host":             brokerHost,
		"rest_gw_host":            restGWHost,
		"ui_host":                 uiHost,
		"tiered_storage_time_sec": TIERED_STORAGE_TIME_FRAME_SEC,
		"ws_port":                 configuration.WS_PORT,
		"http_port":               configuration.HTTP_PORT,
		"clients_port":            configuration.CLIENTS_PORT,
		"rest_gw_port":            configuration.REST_GW_PORT,
	})
}

func (umh UserMgmtHandler) AddUser(c *gin.Context) {
	var body models.AddUserSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	username := strings.ToLower(body.Username)
	exist, _, err := db.GetUserByUsername(username)
	if err != nil {
		serv.Errorf("CreateUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		errMsg := "A user with the name " + body.Username + " already exists"
		serv.Warnf("CreateUser: " + errMsg)
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	userType := strings.ToLower(body.UserType)
	userTypeError := validateUserType(userType)
	if userTypeError != nil {
		serv.Warnf("CreateUser: " + userTypeError.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": userTypeError.Error()})
		return
	}

	usernameError := validateUsername(username)
	if usernameError != nil {
		serv.Warnf("CreateUser: " + usernameError.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": usernameError.Error()})
		return
	}

	var hashedPwdString string
	var avatarId int
	if userType == "management" {
		if body.Password == "" {
			serv.Warnf("CreateUser: Password was not provided for user " + username)
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Password was not provided"})
			return
		}

		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
		if err != nil {
			serv.Errorf("CreateUser: User " + body.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		hashedPwdString = string(hashedPwd)

		avatarId = 1
		if body.AvatarId > 0 {
			avatarId = body.AvatarId
		}
	}

	var brokerConnectionCreds string
	if userType == "application" {
		brokerConnectionCreds, err = AddUser(username)
		if err != nil || len(username) == 0 {
			serv.Errorf("CreateUser: User " + body.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
			return
		}
	}
	newUser, err := db.CreateUser(username, userType, hashedPwdString, "", false, avatarId)
	if err != nil || len(username) == 0 {
		serv.Errorf("CreateUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-add-user")
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
	})
}

func (umh UserMgmtHandler) GetAllUsers(c *gin.Context) {
	users, err := db.GetAllUsers()
	if err != nil {
		serv.Errorf("GetAllUsers: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-enter-users-page")
	}

	if len(users) == 0 {
		c.IndentedJSON(200, []models.User{})
	} else {
		c.IndentedJSON(200, users)
	}
}

func (umh UserMgmtHandler) GetApplicationUsers(c *gin.Context) {
	users, err := db.GetAllApplicationUsers()
	if err != nil {
		serv.Errorf("GetApplicationUsers: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(users) == 0 {
		c.IndentedJSON(200, []models.User{})
	} else {
		c.IndentedJSON(200, users)
	}
}

func (umh UserMgmtHandler) RemoveUser(c *gin.Context) {
	// if err := DenyForSandboxEnv(c); err != nil {
	// 	return
	// }
	var body models.RemoveUserSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	username := strings.ToLower(body.Username)
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}
	if user.Username == username {
		serv.Warnf("RemoveUser: You can not remove your own user")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can not remove your own user"})
		return
	}

	exist, userToRemove, err := db.GetUserByUsername(username)
	if err != nil {
		serv.Errorf("RemoveUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("RemoveUser: User does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "User does not exist"})
		return
	}
	if userToRemove.UserType == "root" {
		serv.Warnf("RemoveUser: You can not remove the root user")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can not remove the root user"})
		return
	}

	err = updateDeletedUserResources(userToRemove)
	if err != nil {
		serv.Errorf("RemoveUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	err = db.DeleteUser(username)
	if err != nil {
		serv.Errorf("RemoveUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-remove-user")
	}

	serv.Noticef("User " + username + " has been deleted by user " + user.Username)
	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) RemoveMyUser(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveMyUser: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}

	if user.UserType == "root" {
		c.AbortWithStatusJSON(500, gin.H{"message": "Root user can not be deleted"})
		return
	}

	err = updateDeletedUserResources(user)
	if err != nil {
		serv.Errorf("RemoveMyUser: User " + user.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	err = db.DeleteUser(user.Username)
	if err != nil {
		serv.Errorf("RemoveMyUser: User " + user.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-remove-himself")
	}

	serv.Noticef("User " + user.Username + " has been deleted")
	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) EditAvatar(c *gin.Context) {
	var body models.EditAvatarSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	avatarId := 1
	if body.AvatarId > 0 {
		avatarId = body.AvatarId
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("EditAvatar: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}

	err = db.EditAvatar(user.Username, avatarId)
	if err != nil {
		serv.Errorf("EditAvatar: User " + user.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{
		"id":                user.ID,
		"username":          user.Username,
		"user_type":         user.UserType,
		"created_at":        user.CreatedAt,
		"already_logged_in": user.AlreadyLoggedIn,
		"avatar_id":         avatarId,
	})
}

func (umh UserMgmtHandler) EditCompanyLogo(c *gin.Context) {
	var file multipart.FileHeader
	ok := utils.Validate(c, nil, true, &file)
	if !ok {
		return
	}

	fileName := "company_logo" + filepath.Ext(file.Filename)
	if err := c.SaveUploadedFile(&file, fileName); err != nil {
		serv.Errorf("EditCompanyLogo: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	base64Encoding, err := imageToBase64(fileName)
	if err != nil {
		serv.Errorf("EditCompanyLogo: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	_ = os.Remove(fileName)

	err = db.InsertImage("company_logo", base64Encoding)
	if err != nil {
		serv.Errorf("EditCompanyLogo: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{"image": base64Encoding})
}

func (umh UserMgmtHandler) RemoveCompanyLogo(c *gin.Context) {
	err := db.DeleteImage("company_logo")
	if err != nil {
		serv.Errorf("RemoveCompanyLogo: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) GetCompanyLogo(c *gin.Context) {
	exist, image, err := db.GetImage("company_logo")
	if !exist {
		c.IndentedJSON(200, gin.H{"image": ""})
		return
	}
	if err != nil {
		serv.Errorf("GetCompanyLogo: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{"image": image.Image})
}

func (umh UserMgmtHandler) EditAnalytics(c *gin.Context) {
	// if err := DenyForSandboxEnv(c); err != nil {
	// 	return
	// }
	var body models.EditAnalyticsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	flag := "false"
	if body.SendAnalytics {
		flag = "true"
	}

	err := db.EditConfigurationValue("analytics", flag)
	if err != nil {
		serv.Errorf("EditAnalytics: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !body.SendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-disable-analytics")
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) DoneNextSteps(c *gin.Context) {
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-done-next-steps")
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) SkipGetStarted(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("SkipGetStarted: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}

	username := strings.ToLower(user.Username)
	err = db.UpdateSkipGetStarted(username)
	if err != nil {
		err2 := db.UpdateSkipGetStartedSandbox(username)
		if err2 != nil {
			serv.Errorf("SkipGetStarted: User " + user.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-skip-get-started")
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) GetActiveUsers() ([]string, error) {
	userList, err := db.GetAllActiveUsers()
	if err != nil {
		return []string{}, err
	}

	var users []string
	for _, user := range userList {
		if user.Username != "" {
			users = append(users, user.Username)
		}
	}

	return users, nil
}

func (umh UserMgmtHandler) GetActiveTags() ([]models.CreateTag, error) {
	tags, err := db.GetAllUsedTags()
	if err != nil {
		return []models.CreateTag{}, err
	}

	tagsRes := []models.CreateTag{}
	for _, tag := range tags {
		tagRes := models.CreateTag{
			Name:  tag.Name,
			Color: tag.Color,
		}
		tagsRes = append(tagsRes, tagRes)
	}
	return tagsRes, nil
}

func (umh UserMgmtHandler) GetFilterDetails(c *gin.Context) {
	var body models.GetFilterDetailsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	switch body.Route {
	case "stations":
		users, err := umh.GetActiveUsers()
		if err != nil {
			serv.Errorf("GetFilterDetails: GetActiveUsers: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		tags, err := umh.GetActiveTags()
		if err != nil {
			serv.Errorf("GetFilterDetails: GetActiveTags: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		storage := []string{"memory", "disk"}
		c.IndentedJSON(200, gin.H{"tags": tags, "users": users, "storage": storage})
		return
	case "schemaverse":
		users, err := umh.GetActiveUsers()
		if err != nil {
			serv.Errorf("GetFilterDetails: GetActiveUsers: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		tags, err := umh.GetActiveTags()
		if err != nil {
			serv.Errorf("GetFilterDetails: GetActiveTags: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		schemaType := []string{"protobuf", "json", "graphql"}
		usage := []string{"used", "not used"}
		c.IndentedJSON(200, gin.H{"tags": tags, "users": users, "type": schemaType, "usage": usage})
		return
	case "syslogs":
		logType := []string{"info", "warn", "err"}
		c.IndentedJSON(200, gin.H{"type": logType})
		return
	default:
		c.IndentedJSON(200, gin.H{})
		return
	}
}
