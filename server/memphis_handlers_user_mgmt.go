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
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/memphisdev/memphis/analytics"
	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"
	"github.com/memphisdev/memphis/utils"

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

func isRootUserLoggedIn() (bool, error) {
	exist, user, err := db.GetRootUser(serv.MemphisGlobalAccountString())
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
		return fmt.Errorf("user type has to be application/management and not %v", userType)
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

func removeTenantResources(tenantName string, user models.User) error {
	err := db.RemoveProducersByTenant(tenantName)
	if err != nil {
		return err
	}

	err = db.RemoveConsumersByTenant(tenantName)
	if err != nil {
		return err
	}

	err = db.RemoveSchemaVersionsByTenant(tenantName)
	if err != nil {
		return err
	}

	err = db.RemoveSchemasByTenant(tenantName)
	if err != nil {
		return err
	}

	err = db.RemoveTagsResourcesByTenant(tenantName)
	if err != nil {
		return err
	}

	err = db.RemoveAuditLogsByTenant(tenantName)
	if err != nil {
		return err
	}

	err = db.DeleteDlsMsgsByTenant(tenantName)
	if err != nil {
		return err
	}

	err = db.DeleteAllTestEvents(tenantName)
	if err != nil {
		return err
	}

	_, err = db.DeleteAndGetAttachedFunctionsByTenant(tenantName)
	if err != nil {
		return err
	}
	// TODO: send response of DeleteAndGetAttachedFunctionsByStation to microservice to delete

	err = db.RemoveStationsByTenant(tenantName)
	if err != nil {
		return err
	}

	err = sendDeleteAllFunctionsReqToMS(user, tenantName, "github", "", "", "aws_lambda", "", true)
	if err != nil {
		return err
	}
	err = deleteInstallationForAuthenticatedGithubApp(user.TenantName)
	if err != nil {
		return err
	}

	err = db.DeleteIntegrationsByTenantName(tenantName)
	if err != nil {
		return err
	}

	users_list, err := db.DeleteUsersByTenant(tenantName)
	if err != nil {
		return err
	}

	SendUserDeleteCacheUpdate(users_list, tenantName)

	err = db.DeleteConfByTenantName(tenantName)
	if err != nil {
		return err
	}

	err = db.DeleteAllSharedLocks(tenantName)
	if err != nil {
		return err
	}

	err = db.DeleteAttachedFunctionsByTenant(tenantName)
	if err != nil {
		return err
	}
	if tenantName != MEMPHIS_GLOBAL_ACCOUNT {
		err = db.RemoveTenant(tenantName)
		if err != nil {
			return err
		}
		serv.PurgeIntegrationsAuditLogs(tenantName)
	}

	err = serv.memphisPurgeResourcesAccount(tenantName)
	if err != nil {
		if err != nil && !IsNatsErr(err, JSStreamNotFoundErr) {
			return err
		}
	}

	if configuration.USER_PASS_BASED_AUTH {
		// send signal to reload config
		err = serv.SendReloadSignal()
		if err != nil {
			return err
		}
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
		serv.Errorf("[tenant: %v][user: %v]EditPassword at getUserDetailsFromMiddleware: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if username == ROOT_USERNAME && user.UserType != "root" {
		errMsg := "Change root password: This operation can be done only by the root user"
		serv.Warnf("[tenant: %v][user: %v]EditPassword: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	} else if username != strings.ToLower(user.Username) && strings.ToLower(user.Username) != ROOT_USERNAME {
		errMsg := "Change user password: This operation can be done only by the user or the root user"
		serv.Warnf("[tenant: %v][user: %v]EditPassword: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]EditPassword at GenerateFromPassword: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	hashedPwdString := string(hashedPwd)
	err = db.ChangeUserPassword(username, hashedPwdString, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]EditPassword at ChangeUserPassword: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) GetSignUpFlag(c *gin.Context) {
	showSignup := true
	loggedIn, err := isRootUserLoggedIn()
	if err != nil {
		serv.Errorf("GetSignUpFlag: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if loggedIn {
		showSignup = false
	} else {
		count, err := db.CountAllUsers()
		if err != nil {
			serv.Errorf("GetSignUpFlag: %v", err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if count > 1 { // more than 1 user exists
			showSignup = false
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent("", "", analyticsParams, "user-open-ui")
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
		serv.Errorf("CreateUserSignUp at GenerateFromPassword: User %v: %v", body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	hashedPwdString := string(hashedPwd)
	subscription := body.Subscribtion

	newUser, err := db.CreateUser(username, "management", hashedPwdString, fullName, subscription, 1, serv.MemphisGlobalAccountString(), false, "", "", "", "")
	if err != nil {
		if strings.Contains(err.Error(), "already exist") {
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "User already exists"})
			return
		}
		serv.Errorf("CreateUserSignUp error at db.CreateUser: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	serv.Noticef("User %v has been signed up", username)
	token, refreshToken, err := CreateTokens(newUser)
	if err != nil {
		serv.Errorf("CreateUserSignUp error at CreateTokens: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	env := "K8S"
	if configuration.DOCKER_ENV != "" || configuration.LOCAL_CLUSTER_ENV {
		env = "docker"
	}

	exist, tenant, err := db.GetTenantByName(newUser.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateUserSignUp at GetTenantByName: User %v: %v", newUser.TenantName, newUser.Username, body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("[tenant: %v][user: %v]CreateUserSignUp: User %v: tenant %v does not exist", newUser.TenantName, newUser.Username, body.Username, newUser.TenantName)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	decriptionKey := getAESKey()
	decryptedUserPassword, err := DecryptAES(decriptionKey, tenant.InternalWSPass)
	if err != nil {
		serv.Errorf("CreateUserSignUp: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := map[string]interface{}{
			"email":        username,
			"newsletter":   strconv.FormatBool(subscription),
			"organization": body.Organization,
			"full_name":    body.FullName,
		}
		analytics.SendEvent(newUser.TenantName, username, analyticsParams, "user-signup")
	}

	domain := ""
	secure := false
	c.SetCookie("memphis-jwt-refresh-token", refreshToken, REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
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
		"account_id":              tenant.ID,
		"internal_ws_pass":        decryptedUserPassword,
		"dls_retention":           serv.opts.DlsRetentionHours[newUser.TenantName],
		"logs_retention":          serv.opts.LogsRetentionDays,
		"max_msg_size_mb":         serv.opts.MaxPayload / 1024 / 1024,
	})
}

func (umh UserMgmtHandler) GetAllUsers(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetAllUsers: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	users, err := db.GetAllUsers(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetAllUsers: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-enter-users-page")
	}

	applicationUsers := []models.FilteredGenericUser{}
	managementUsers := []models.FilteredGenericUser{}

	for _, user := range users {
		if user.UserType == "application" {
			applicationUsers = append(applicationUsers, user)
		} else if user.UserType == "management" || user.UserType == "root" {
			managementUsers = append(managementUsers, user)
		}
	}

	if len(users) == 0 {
		c.IndentedJSON(200, []models.User{})
	} else {
		c.IndentedJSON(200, gin.H{"application_users": applicationUsers, "management_users": managementUsers})
	}
}

func (umh UserMgmtHandler) GetApplicationUsers(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetApplicationUsers: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	users, err := db.GetAllUsersByTypeAndTenantName([]string{"application"}, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetApplicationUsers at GetAllUsersByTypeAndTenantName: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(users) == 0 {
		c.IndentedJSON(200, []models.User{})
	} else {
		c.IndentedJSON(200, users)
	}
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
		serv.Errorf("EditAvatar: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	err = db.EditAvatar(user.Username, avatarId, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]EditAvatar: User %v: %v", user.TenantName, user.Username, user.Username, err.Error())
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
		serv.Errorf("EditCompanyLogo at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	fileName := "company_logo" + filepath.Ext(file.Filename)
	if err := c.SaveUploadedFile(&file, fileName); err != nil {
		serv.Errorf("[tenant: %v][user: %v]EditCompanyLogo at SaveUploadedFile: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	base64Encoding, err := imageToBase64(fileName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]EditCompanyLogo at imageToBase64: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	_ = os.Remove(fileName)

	err = db.InsertImage("company_logo", base64Encoding, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]EditCompanyLogo error insertin image to db: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{"image": base64Encoding})
}

func (umh UserMgmtHandler) RemoveCompanyLogo(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveCompanyLogo at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	err = db.DeleteImage("company_logo", user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RemoveCompanyLogo at deleting from the db: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) GetCompanyLogo(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetCompanyLogo at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	exist, image, err := db.GetImage("company_logo", user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetCompanyLogo at db.GetImage: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.IndentedJSON(200, gin.H{"image": ""})
		return
	}

	c.IndentedJSON(200, gin.H{"image": image.Image})
}

func (umh UserMgmtHandler) DoneNextSteps(c *gin.Context) {
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-done-next-steps")
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) SkipGetStarted(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("SkipGetStarted at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	username := strings.ToLower(user.Username)
	err = db.UpdateSkipGetStarted(username, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]SkipGetStarted at UpdateSkipGetStarted: User %v: %v", user.TenantName, user.Username, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-skip-get-started")
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) GetActiveUsers(tenantName, page string) ([]string, error) {
	var users []string
	var userList []models.FilteredUser
	var err error

	switch page {
	case "stations":
		userList, err = db.GetAllActiveUsersStations(tenantName)
		if err != nil {
			return []string{}, err
		}
	case "schemaverse":
		userList, err = db.GetAllActiveUsersSchemaVersions(tenantName)
		if err != nil {
			return []string{}, err
		}
	}

	for _, user := range userList {
		if user.Username != "" {
			users = append(users, user.Username)
		}
	}

	return users, nil
}

func (umh UserMgmtHandler) GetActiveTags(tenantName, page string) ([]models.CreateTag, error) {
	tagsRes := []models.CreateTag{}
	var tags []models.Tag
	var err error

	switch page {
	case "stations":
		tags, err = db.GetAllUsedStationsTags(tenantName)
		if err != nil {
			return []models.CreateTag{}, err
		}
	case "schemaverse":
		tags, err = db.GetAllUsedSchemasTags(tenantName)
		if err != nil {
			return []models.CreateTag{}, err
		}
	}

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
		serv.Errorf("GetFilterDetails at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	tenantName := user.TenantName
	route := strings.ToLower(body.Route)
	switch body.Route {
	case "stations":
		users, err := umh.GetActiveUsers(tenantName, route)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]GetFilterDetails: GetActiveUsers: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		tags, err := umh.GetActiveTags(tenantName, route)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]GetFilterDetails: GetActiveTags: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		storage := []string{"memory", "disk"}
		c.IndentedJSON(200, gin.H{"tags": tags, "users": users, "storage": storage})
		return
	case "schemaverse":
		users, err := umh.GetActiveUsers(tenantName, route)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]GetFilterDetails: GetActiveUsers: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		tags, err := umh.GetActiveTags(tenantName, route)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]GetFilterDetails: GetActiveTags: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		schemaType := []string{"protobuf", "json", "graphql", "avro"}
		usage := []string{"used", "not used"}
		c.IndentedJSON(200, gin.H{"tags": tags, "users": users, "type": schemaType, "usage": usage})
		return
	case "syslogs":
		logType := []string{"info", "warn", "err"}
		v, err := serv.Varz(nil)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]GetFilterDetails: %v", user.TenantName, user.Username, err.Error())
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

func SendUserDeleteCacheUpdate(usernames []string, tenantName string) {
	deleteRequest := models.CacheUpdateRequest{
		CacheType:  "user",
		Operation:  "delete",
		Usernames:  usernames,
		TenantName: tenantName,
	}

	msg, err := json.Marshal(deleteRequest)
	if err != nil {
		serv.Errorf("[tenant: %v] user cache at SendUserCacheUpdates json.Marshal: %v", tenantName, err.Error())
		return
	}

	err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), CACHE_UDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		serv.Errorf("[tenant: %v]user cache at SendUserCacheUpdates: error sending internal msg : %v", tenantName, err.Error())
		return
	}
}

func validateUsername(username string) error {
	if len(username) > 60 {
		return errors.New("username exceeds the maximum allowed length of 60 characters")
	}
	re := regexp.MustCompile("^[a-z0-9_.-]*$")
	validName := re.MatchString(username)
	if !validName || len(username) == 0 {
		return errors.New("username has to include only letters/numbers/./_/- ")
	}
	return nil
}

func validateUserDescription(description string) error {
	if len(description) > 100 {
		return errors.New("description exceeds the maximum allowed length of 100 characters")
	}
	return nil
}

func validateUserTeam(team string) error {
	if len(team) > 20 {
		return errors.New("team exceeds the maximum allowed length of 20 characters")
	}
	return nil
}

func validateUserPosition(position string) error {
	if len(position) > 30 {
		return errors.New("position exceeds the maximum allowed length of 30 characters")
	}
	return nil
}

func validateUserFullName(fullName string) error {
	if len(fullName) > 30 {
		return errors.New("full name exceeds the maximum allowed length of 30 characters")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) > 20 {
		return errors.New("password exceeds the maximum allowed length of 20 characters")
	}
	pattern := `^[A-Za-z0-9!?\-@#$%]+$`
	match, _ := regexp.MatchString(pattern, password)
	if !match {
		return errors.New("Password must be at least 8 characters long, contain both uppercase and lowercase, and at least one number and one special character")
	}
	if len(password) < 8 {
		return errors.New("Password must be at least 8 characters long, contain both uppercase and lowercase, and at least one number and one special character")
	}
	var (
		hasUppercase   bool
		hasLowercase   bool
		hasDigit       bool
		hasSpecialChar bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUppercase = true
		case unicode.IsLower(char):
			hasLowercase = true
		case unicode.IsDigit(char):
			hasDigit = true
		case char == '!' || char == '?' || char == '-' || char == '@' || char == '#' || char == '$' || char == '%':
			hasSpecialChar = true
		}
	}

	if hasUppercase && hasLowercase && hasDigit && hasSpecialChar {
		return nil
	}

	return errors.New("Password must be at least 8 characters long, contain both uppercase and lowercase, and at least one number and one special character")
}

func CreateInternalApplicationUserForExistTenants() error {
	tenants, err := db.GetAllTenantsWithoutGlobal()
	if err != nil {
		return err
	}
	password, err := EncryptAES([]byte(configuration.CONNECTION_TOKEN + "_" + configuration.ROOT_PASSWORD))
	if err != nil {
		return err
	}
	for _, tenant := range tenants {
		_, err := db.CreateUserIfNotExist("$"+tenant.Name, "application", password, "", false, 1, tenant.Name, false, "", "", "", "")
		if err != nil {
			return err
		}
	}

	return nil
}

func (umh UserMgmtHandler) SendTrace(c *gin.Context) {
	var body models.SendTraceSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	traceName := strings.ToLower(body.TraceName)
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]SendTrace at getUserDetailsFromMiddleware: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.TenantName, user.Username, body.TraceParams, traceName)
	}

	c.IndentedJSON(200, gin.H{})
}
