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

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	REFRESH_JWT_EXPIRES_IN_MINUTES = 2880
	JWT_EXPIRES_IN_MINUTES         = 15
	ROOT_USERNAME                  = "root"
	MEMPHIS_USERNAME               = "$memphis_user"
)

type UserMgmtHandler struct{}

func isRootUserExist() (bool, error) {
	exist, _, err := db.GetRootUser(globalAccountName)
	if err != nil {
		return false, err
	} else if !exist {
		return false, nil
	}
	return true, nil
}

func isRootUserLoggedIn() (bool, error) {
	exist, user, err := db.GetRootUser(globalAccountName)
	if err != nil {
		return false, err
	} else if !exist {
		return false, errors.New("root user does not exist")
	}

	if user.AlreadyLoggedIn {
		return true, nil
	} else {
		return false, nil
	}
}

func authenticateUser(username string, password string) (bool, models.User, error) {
	exist, user, err := db.GetUserForLogin(username)
	if err != nil {
		return false, models.User{}, err
	} else if !exist {
		return false, models.User{}, nil
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
	tenantName := user.TenantName
	if user.UserType == "application" {
		err := RemoveUser(user.Username)
		if err != nil {
			return err
		}
	}

	err := db.UpdateStationsOfDeletedUser(user.ID, tenantName)
	if err != nil {
		return err
	}

	err = db.UpdateConnectionsOfDeletedUser(user.ID, tenantName)
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

	err = db.UpdateSchemasOfDeletedUser(user.ID, tenantName)
	if err != nil {
		return err
	}

	err = db.UpdateSchemaVersionsOfDeletedUser(user.ID, tenantName)
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
	models.User
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
		atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(JWT_EXPIRES_IN_MINUTES)).Unix()
		atClaims["tenant_name"] = u.TenantName
		at = jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	}
	token, err := at.SignedString([]byte(configuration.JWT_SECRET))
	if err != nil {
		return "", "", err
	}

	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(REFRESH_JWT_EXPIRES_IN_MINUTES)).Unix()

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
	if username == ROOT_USERNAME && user.UserType != "root" {
		errMsg := "Change root password: This operation can be done only by the root user"
		serv.Warnf("EditPassword: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	} else if username != strings.ToLower(user.Username) && strings.ToLower(user.Username) != ROOT_USERNAME {
		errMsg := "Change user password: This operation can be done only by the user or the root user"
		serv.Warnf("EditPassword: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
	if err != nil {
		serv.Errorf("EditPassword: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	hashedPwdString := string(hashedPwd)
	err = db.ChangeUserPassword(username, hashedPwdString, user.TenantName)
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
		err = db.UpdateUserAlreadyLoggedIn(user.ID)
		if err != nil {
			serv.Errorf("Login: User " + body.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-login")
	}

	env := "K8S"
	if configuration.DOCKER_ENV != "" || configuration.LOCAL_CLUSTER_ENV {
		env = "docker"
	}
	exist, tenant, err := db.GetTenantByName(user.TenantName)
	if err != nil {
		serv.Errorf("Login: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("Login: User " + body.Username + ": tenant " + user.TenantName + " does not exist")
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	domain := ""
	secure := false
	c.SetCookie("jwt-refresh-token", refreshToken, REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
	c.IndentedJSON(200, gin.H{
		"jwt":                     token,
		"expires_in":              JWT_EXPIRES_IN_MINUTES * 60 * 1000,
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
		"broker_host":             serv.opts.BrokerHost,
		"rest_gw_host":            serv.opts.RestGwHost,
		"ui_host":                 serv.opts.UiHost,
		"tiered_storage_time_sec": serv.opts.TieredStorageUploadIntervalSec,
		"ws_port":                 serv.opts.Websocket.Port,
		"http_port":               serv.opts.UiPort,
		"clients_port":            serv.opts.Port,
		"rest_gw_port":            serv.opts.RestGwPort,
		"user_pass_based_auth":    configuration.USER_PASS_BASED_AUTH,
		"connection_token":        configuration.CONNECTION_TOKEN,
		"account_id":              tenant.ID,
	})
}

func (umh UserMgmtHandler) RefreshToken(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("refreshToken: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}
	username := user.Username
	_, systemKey, err := db.GetSystemKey("analytics", globalAccountName)
	if err != nil {
		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	sendAnalytics, _ := strconv.ParseBool(systemKey.Value)
	exist, user, err := db.GetUserByUsername(username, user.TenantName)
	if err != nil {
		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("refreshToken: user " + username + " does not exist")
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	token, refreshToken, err := CreateTokens(user)
	if err != nil {
		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	env := "K8S"
	if configuration.DOCKER_ENV != "" || configuration.LOCAL_CLUSTER_ENV {
		env = "docker"
	}

	exist, tenant, err := db.GetTenantByName(user.TenantName)
	if err != nil {
		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("Login: User " + username + ": tenant " + user.TenantName + " does not exist")
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	domain := ""
	secure := true
	c.SetCookie("jwt-refresh-token", refreshToken, REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
	c.IndentedJSON(200, gin.H{
		"jwt":                     token,
		"expires_in":              JWT_EXPIRES_IN_MINUTES * 60 * 1000,
		"user_id":                 user.ID,
		"username":                user.Username,
		"user_type":               user.UserType,
		"created_at":              user.CreatedAt,
		"already_logged_in":       user.AlreadyLoggedIn,
		"avatar_id":               user.AvatarId,
		"send_analytics":          sendAnalytics,
		"env":                     env,
		"namespace":               serv.opts.K8sNamespace,
		"full_name":               user.FullName,
		"skip_get_started":        user.SkipGetStarted,
		"broker_host":             serv.opts.BrokerHost,
		"rest_gw_host":            serv.opts.RestGwHost,
		"ui_host":                 serv.opts.UiHost,
		"tiered_storage_time_sec": serv.opts.TieredStorageUploadIntervalSec,
		"ws_port":                 serv.opts.Websocket.Port,
		"http_port":               serv.opts.UiPort,
		"clients_port":            serv.opts.Port,
		"rest_gw_port":            serv.opts.RestGwPort,
		"user_pass_based_auth":    configuration.USER_PASS_BASED_AUTH,
		"connection_token":        configuration.CONNECTION_TOKEN,
		"account_id":              tenant.ID,
	})
}

func (umh UserMgmtHandler) GetSignUpFlag(c *gin.Context) {
	showSignup := true
	loggedIn, err := isRootUserLoggedIn()
	if err != nil {
		serv.Errorf("GetSignUpFlag: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if loggedIn {
		showSignup = false
	} else {
		count, err := db.CountAllUsers()
		if err != nil {
			serv.Errorf("GetSignUpFlag: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if count > 1 { // more than 1 user exists
			showSignup = false
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent("", "user-open-ui")
	}
	c.IndentedJSON(200, gin.H{"show_signup": showSignup})
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
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": usernameError.Error()})
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

	newUser, err := db.CreateUser(username, "management", hashedPwdString, fullName, subscription, 1, globalAccountName)
	if err != nil {
		if strings.Contains(err.Error(), "already exist") {
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "User already exist"})
			return
		}
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

	env := "K8S"
	if configuration.DOCKER_ENV != "" || configuration.LOCAL_CLUSTER_ENV {
		env = "docker"
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
	c.SetCookie("jwt-refresh-token", refreshToken, REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
	c.IndentedJSON(200, gin.H{
		"jwt":                     token,
		"expires_in":              JWT_EXPIRES_IN_MINUTES * 60 * 1000,
		"user_id":                 newUser.ID,
		"username":                newUser.Username,
		"user_type":               newUser.UserType,
		"created_at":              newUser.CreatedAt,
		"already_logged_in":       newUser.AlreadyLoggedIn,
		"avatar_id":               newUser.AvatarId,
		"send_analytics":          shouldSendAnalytics,
		"env":                     env,
		"namespace":               serv.opts.K8sNamespace,
		"full_name":               newUser.FullName,
		"skip_get_started":        newUser.SkipGetStarted,
		"broker_host":             serv.opts.BrokerHost,
		"rest_gw_host":            serv.opts.RestGwHost,
		"ui_host":                 serv.opts.UiHost,
		"tiered_storage_time_sec": serv.opts.TieredStorageUploadIntervalSec,
		"ws_port":                 serv.opts.Websocket.Port,
		"http_port":               serv.opts.UiPort,
		"clients_port":            serv.opts.Port,
		"rest_gw_port":            serv.opts.RestGwPort,
		"user_pass_based_auth":    configuration.USER_PASS_BASED_AUTH,
		"connection_token":        configuration.CONNECTION_TOKEN,
	})
}

func (umh UserMgmtHandler) AddUser(c *gin.Context) {
	var body models.AddUserSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("AddUser: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	username := strings.ToLower(body.Username)
	exist, _, err := db.GetUserByUsername(username, user.TenantName)
	if err != nil {
		serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		errMsg := "A user with the name " + body.Username + " already exists"
		serv.Warnf("CreateUser: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	userType := strings.ToLower(body.UserType)
	userTypeError := validateUserType(userType)
	if userTypeError != nil {
		serv.Warnf("AddUser: " + userTypeError.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": userTypeError.Error()})
		return
	}

	usernameError := validateUsername(username)
	if usernameError != nil {
		serv.Warnf("AddUser: " + usernameError.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": usernameError.Error()})
		return
	}

	var password string
	var avatarId int
	if userType == "management" {
		if body.Password == "" {
			serv.Warnf("AddUser: Password was not provided for user " + username)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Password was not provided"})
			return
		}

		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
		if err != nil {
			serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		password = string(hashedPwd)

		avatarId = 1
		if body.AvatarId > 0 {
			avatarId = body.AvatarId
		}
	}

	var brokerConnectionCreds string
	if userType == "application" {
		if configuration.USER_PASS_BASED_AUTH {
			if body.Password == "" {
				serv.Warnf("AddUser: Password was not provided for user " + username)
				c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Password was not provided"})
				return
			}
			password, err = EncryptAES([]byte(body.Password))
			if err != nil {
				serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			avatarId = 1
			if body.AvatarId > 0 {
				avatarId = body.AvatarId
			}
		} else {
			brokerConnectionCreds = configuration.CONNECTION_TOKEN
		}
	}
	newUser, err := db.CreateUser(username, userType, password, "", false, avatarId, user.TenantName)
	if err != nil {
		serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-add-user")
	}

	if userType == "application" && configuration.USER_PASS_BASED_AUTH {
		// send signal to reload config
		err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), CONFIGURATIONS_RELOAD_SIGNAL_SUBJ, _EMPTY_, nil, _EMPTY_, true)
		if err != nil {
			serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
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
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetAllUsers: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	users, err := db.GetAllUsers(user.TenantName)
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
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetApplicationUsers: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}
	users, err := db.GetAllUsersByTypeAndTenantName([]string{"application"}, user.TenantName)
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
		return
	}
	if user.Username == username {
		serv.Warnf("RemoveUser: You can not remove your own user")
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can not remove your own user"})
		return
	}

	exist, userToRemove, err := db.GetUserByUsername(username, user.TenantName)
	if err != nil {
		serv.Errorf("RemoveUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("RemoveUser: User does not exist")
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "User does not exist"})
		return
	}
	if userToRemove.UserType == "root" {
		serv.Warnf("RemoveUser: You can not remove the root user")
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can not remove the root user"})
		return
	}

	err = updateDeletedUserResources(userToRemove)
	if err != nil {
		serv.Errorf("RemoveUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	err = db.DeleteUser(username, userToRemove.TenantName)
	if err != nil {
		serv.Errorf("RemoveUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if userToRemove.UserType == "application" && configuration.USER_PASS_BASED_AUTH {
		// send signal to reload config
		err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), CONFIGURATIONS_RELOAD_SIGNAL_SUBJ, _EMPTY_, nil, _EMPTY_, true)
		if err != nil {
			serv.Errorf("RemoveUser: User " + body.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
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
		return
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

	err = db.DeleteUser(user.Username, user.TenantName)
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
		return
	}

	err = db.EditAvatar(user.Username, avatarId, user.TenantName)
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

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("EditCompanyLogo: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
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

	err = db.InsertImage("company_logo", base64Encoding, user.TenantName)
	if err != nil {
		serv.Errorf("EditCompanyLogo: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{"image": base64Encoding})
}

func (umh UserMgmtHandler) RemoveCompanyLogo(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveCompanyLogo: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}
	err = db.DeleteImage("company_logo", user.TenantName)
	if err != nil {
		serv.Errorf("RemoveCompanyLogo: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) GetCompanyLogo(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetCompanyLogo: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}
	exist, image, err := db.GetImage("company_logo", user.TenantName)
	if err != nil {
		serv.Errorf("GetCompanyLogo: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.IndentedJSON(200, gin.H{"image": ""})
		return
	}

	c.IndentedJSON(200, gin.H{"image": image.Image})
}

func (umh UserMgmtHandler) EditAnalytics(c *gin.Context) {
	var body models.EditAnalyticsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	flag := "false"
	if body.SendAnalytics {
		flag = "true"
	}

	err := db.EditConfigurationValue("analytics", flag, globalAccountName)
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
		return
	}

	username := strings.ToLower(user.Username)
	err = db.UpdateSkipGetStarted(username, user.TenantName)
	if err != nil {
		serv.Errorf("SkipGetStarted: User " + user.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-skip-get-started")
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) GetActiveUsers(tenantName string) ([]string, error) {
	userList, err := db.GetAllActiveUsers(tenantName)
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

func (umh UserMgmtHandler) GetActiveTags(tenantName string) ([]models.CreateTag, error) {
	tags, err := db.GetAllUsedTags(tenantName)
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

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetFilterDetails: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}
	tenantName := user.TenantName
	switch body.Route {
	case "stations":
		users, err := umh.GetActiveUsers(tenantName)
		if err != nil {
			serv.Errorf("GetFilterDetails: GetActiveUsers: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		tags, err := umh.GetActiveTags(tenantName)
		if err != nil {
			serv.Errorf("GetFilterDetails: GetActiveTags: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		storage := []string{"memory", "disk"}
		c.IndentedJSON(200, gin.H{"tags": tags, "users": users, "storage": storage})
		return
	case "schemaverse":
		users, err := umh.GetActiveUsers(tenantName)
		if err != nil {
			serv.Errorf("GetFilterDetails: GetActiveUsers: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		tags, err := umh.GetActiveTags(tenantName)
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
		v, err := serv.Varz(nil)
		if err != nil {
			serv.Errorf("GetFilterDetails: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		var logSource []string
		if len(v.Cluster.URLs) == 0 {
			logSource = append(logSource, "memphis-0")
		}
		for i := range v.Cluster.URLs {
			logSource = append(logSource, "memphis-"+strconv.Itoa(i))
		}
		logSource = append(logSource, "rest-gateway")

		c.IndentedJSON(200, gin.H{"type": logType, "source": logSource})
		return
	default:
		c.IndentedJSON(200, gin.H{})
		return
	}
}
