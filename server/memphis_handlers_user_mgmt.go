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
package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"memphis-broker/analytics"
	"memphis-broker/models"
	"memphis-broker/utils"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/hanzoai/gochimp3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserMgmtHandler struct{}
type MailChimpErr struct {
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

func isRootUserExist() (bool, error) {
	filter := bson.M{"user_type": "root"}
	var user models.User
	err := usersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func isOnlyRootUserExist() (bool, error) {
	if configuration.SANDBOX_ENV == "true" {
		return false, nil
	}
	cursor, err := usersCollection.Find(context.TODO(), bson.M{})
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	var users []models.User
	if err = cursor.All(context.TODO(), &users); err != nil {
		serv.Errorf("isOnlyRootUserExist error: " + err.Error())
		return false, err
	}
	if len(users) == 1 && users[0].UserType == "root" {
		return true, nil
	} else {
		return false, nil
	}
}

func authenticateUser(username string, password string) (bool, models.User, error) {
	filter := bson.M{"username": username}
	var user models.User
	err := usersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
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

// TODO check against hub api
func validateHubCreds(hubUsername string, hubPassword string) error {
	if hubUsername != "" && hubPassword != "" {
		// TODO
	}
	return nil
}

func updateUserResources(user models.User) error {
	if user.UserType == "application" {
		err := RemoveUser(user.Username)
		if err != nil {
			return err
		}
	}

	_, err := factoriesCollection.UpdateMany(context.TODO(),
		bson.M{"created_by_user": user.Username},
		bson.M{"$set": bson.M{"created_by_user": user.Username + "(deleted)"}},
	)
	if err != nil {
		return err
	}

	_, err = stationsCollection.UpdateMany(context.TODO(),
		bson.M{"created_by_user": user.Username},
		bson.M{"$set": bson.M{"created_by_user": user.Username + "(deleted)"}},
	)
	if err != nil {
		return err
	}

	_, err = connectionsCollection.UpdateMany(context.TODO(),
		bson.M{"created_by_user": user.Username},
		bson.M{"$set": bson.M{"created_by_user": user.Username + "(deleted)", "is_active": false}},
	)
	if err != nil {
		return err
	}

	_, err = producersCollection.UpdateMany(context.TODO(),
		bson.M{"created_by_user": user.Username},
		bson.M{"$set": bson.M{"created_by_user": user.Username + "(deleted)", "is_active": false}},
	)
	if err != nil {
		return err
	}

	_, err = consumersCollection.UpdateMany(context.TODO(),
		bson.M{"created_by_user": user.Username},
		bson.M{"$set": bson.M{"created_by_user": user.Username + "(deleted)", "is_active": false}},
	)
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
		atClaims["user_id"] = u.ID.Hex()
		atClaims["username"] = u.Username
		atClaims["user_type"] = u.UserType
		atClaims["creation_date"] = u.CreationDate
		atClaims["already_logged_in"] = u.AlreadyLoggedIn
		atClaims["avatar_id"] = u.AvatarId
		atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(configuration.JWT_EXPIRES_IN_MINUTES)).Unix()
		at = jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	case models.SandboxUser:
		atClaims["user_id"] = u.ID.Hex()
		atClaims["username"] = u.Username
		atClaims["user_type"] = u.UserType
		atClaims["creation_date"] = u.CreationDate
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
	bytes, err := ioutil.ReadFile(imagePath)
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
		if configuration.ANALYTICS == "true" {
			analytics.IncrementInstallationsCounter()
		}

		newUser := models.User{
			ID:              primitive.NewObjectID(),
			Username:        "root",
			Password:        hashedPwdString,
			HubUsername:     "",
			HubPassword:     "",
			UserType:        "root",
			CreationDate:    time.Now(),
			AlreadyLoggedIn: false,
			AvatarId:        1,
		}

		_, err = usersCollection.InsertOne(context.TODO(), newUser)
		if err != nil {
			return err
		}
	} else {
		_, err = usersCollection.UpdateOne(context.TODO(),
			bson.M{"username": "root"},
			bson.M{"$set": bson.M{"password": hashedPwdString}},
		)
		if err != nil {
			return err
		}
	}

	return nil
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
		serv.Errorf("Login error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !authenticated || user.UserType == "application" {
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	var systemKey models.SystemKey
	err = systemKeysCollection.FindOne(context.TODO(), bson.M{"key": "analytics"}).Decode(&systemKey)
	if err != nil {
		serv.Errorf("Login error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	sendAnalytics, _ := strconv.ParseBool(systemKey.Value)

	token, refreshToken, err := CreateTokens(user)
	if err != nil {
		serv.Errorf("Login error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !user.AlreadyLoggedIn {
		usersCollection.UpdateOne(context.TODO(),
			bson.M{"_id": user.ID},
			bson.M{"$set": bson.M{"already_logged_in": true}},
		)
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.IncrementLoginsCounter()
	}

	var env string
	if configuration.DOCKER_ENV != "" {
		env = "docker"
	} else {
		env = "K8S"
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
		"send_analytics":    sendAnalytics,
		"env":               env,
		"namespace":         configuration.K8S_NAMESPACE,
		"full_name":         user.FullName,
	})
}

func (umh UserMgmtHandler) RefreshToken(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("refreshToken error: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}
	_, user, err = IsUserExist(user.Username)
	if err != nil {
		serv.Errorf("RefreshToken error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	var systemKey models.SystemKey
	err = systemKeysCollection.FindOne(context.TODO(), bson.M{"key": "analytics"}).Decode(&systemKey)
	if err != nil {
		serv.Errorf("RefreshToken error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	sendAnalytics, _ := strconv.ParseBool(systemKey.Value)

	token, refreshToken, err := CreateTokens(user)
	if err != nil {
		serv.Errorf("RefreshToken error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	var env string
	if configuration.DOCKER_ENV != "" {
		env = "docker"
	} else {
		env = "K8S"
	}

	domain := ""
	secure := true
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
		"send_analytics":    sendAnalytics,
		"env":               env,
		"namespace":         configuration.K8S_NAMESPACE,
	})
}

// TODO
func (umh UserMgmtHandler) AuthenticateNatsUser(c *gin.Context) {
	publicKey := c.Param("publicKey")
	if publicKey != "" {
		fmt.Println(publicKey)
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) GetSignUpFlag(c *gin.Context) {
	exist, err := isOnlyRootUserExist()
	if err != nil {
		serv.Errorf("GetSignUpFlag error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, gin.H{"exist": exist})
}

func (umh UserMgmtHandler) AddUserSignUp(c *gin.Context) {
	var body models.AddUserSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	exist, err := isOnlyRootUserExist()
	if err != nil {
		serv.Errorf("CreateUserSignUp error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		c.IndentedJSON(401, gin.H{"message": "Unauthorized"})
		return
	} else {
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
			serv.Errorf("CreateUserSignUp error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		hashedPwdString := string(hashedPwd)
		subscription := body.Subscribtion

		var tag []string
		if subscription {
			tag = []string{"installation", "newsletter"}
		} else {
			tag = []string{"installation"}
		}

		mailchimpClient := gochimp3.New(configuration.MAILCHIMP_KEY)
		mailchimpListID := configuration.MAILCHIMP_LIST_ID
		mailchimpList, err := mailchimpClient.GetList(mailchimpListID, nil)
		if err != nil {
			serv.Debugf("getList in mailchimp error: " + err.Error())
		}

		mailchimpReq := &gochimp3.MemberRequest{
			EmailAddress: username,
			Status:       "subscribed",
			Tags:         tag,
		}
		_, err = mailchimpList.CreateMember(mailchimpReq)
		if err != nil {
			data, err := json.Marshal(err)
			if err != nil {
				serv.Debugf("Error: " + err.Error())
			}
			var mailChimpErr MailChimpErr
			if err = json.Unmarshal([]byte(data), &mailChimpErr); err != nil {
				serv.Debugf("Error: " + err.Error())
			}
			mailChimpReqSearch := &gochimp3.SearchMembersQueryParams{
				Query: body.Username,
			}
			if data != nil {
				if mailChimpErr.Title == "Member Exists" && mailChimpErr.Status == 400 {
					res, err := mailchimpList.SearchMembers(mailChimpReqSearch)
					if err != nil {
						serv.Debugf("Failed to search member in mailChimp: " + err.Error())
					}
					_, err = mailchimpList.UpdateMember(res.ExactMatches.Members[0].ID, mailchimpReq)
					if err != nil {
						serv.Debugf("Failed to update member in mailChimp: " + err.Error())
					}
				} else {
					serv.Debugf("Failed to subscribe in mailChimp: " + err.Error())
				}
			}
		}

		newUser := models.User{
			ID:              primitive.NewObjectID(),
			Username:        username,
			Password:        hashedPwdString,
			FullName:        fullName,
			Subscribtion:    subscription,
			UserType:        "management",
			CreationDate:    time.Now(),
			AlreadyLoggedIn: false,
			AvatarId:        1,
		}

		_, err = usersCollection.InsertOne(context.TODO(), newUser)
		if err != nil {
			serv.Errorf("CreateUserSignUp error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		serv.Noticef("User " + username + " has been created")
		token, refreshToken, err := CreateTokens(newUser)
		if err != nil {
			serv.Errorf("CreateUserSignUp error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		var env string
		if configuration.DOCKER_ENV != "" {
			env = "docker"
		} else {
			env = "K8S"
		}

		var systemKey models.SystemKey
		err = systemKeysCollection.FindOne(context.TODO(), bson.M{"key": "analytics"}).Decode(&systemKey)
		if err != nil {
			serv.Errorf("Login error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		sendAnalytics, _ := strconv.ParseBool(systemKey.Value)

		domain := ""
		secure := false
		c.SetCookie("jwt-refresh-token", refreshToken, configuration.REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
		c.IndentedJSON(200, gin.H{
			"jwt":               token,
			"expires_in":        configuration.JWT_EXPIRES_IN_MINUTES * 60 * 1000,
			"user_id":           newUser.ID,
			"username":          newUser.Username,
			"user_type":         newUser.UserType,
			"creation_date":     newUser.CreationDate,
			"already_logged_in": newUser.AlreadyLoggedIn,
			"avatar_id":         newUser.AvatarId,
			"send_analytics":    sendAnalytics,
			"env":               env,
			"namespace":         configuration.K8S_NAMESPACE,
			"full_name":         newUser.FullName,
		})

	}
}

func (umh UserMgmtHandler) AddUser(c *gin.Context) {
	var body models.AddUserSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	username := strings.ToLower(body.Username)
	exist, _, err := IsUserExist(username)
	if err != nil {
		serv.Errorf("CreateUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		serv.Warnf("A user with this username is already exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "A user with this username is already exist"})
		return
	}

	userType := strings.ToLower(body.UserType)
	userTypeError := validateUserType(userType)
	if userTypeError != nil {
		serv.Warnf(userTypeError.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": userTypeError.Error()})
		return
	}

	usernameError := validateUsername(username)
	if usernameError != nil {
		serv.Warnf(usernameError.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": usernameError.Error()})
		return
	}

	var hashedPwdString string
	var avatarId int
	if userType == "management" {
		if body.Password == "" {
			serv.Warnf("Password was not provided")
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Password was not provided"})
			return
		}

		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
		if err != nil {
			serv.Errorf("CreateUser error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		hashedPwdString = string(hashedPwd)

		avatarId = 1
		if body.AvatarId > 0 {
			avatarId = body.AvatarId
		}
	}

	err = validateHubCreds(body.HubUsername, body.HubPassword)
	if err != nil {
		serv.Errorf("CreateUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	var brokerConnectionCreds string
	if userType == "application" {
		brokerConnectionCreds, err = AddUser(username)
		if err != nil || len(username) == 0 {
			serv.Errorf("CreateUser error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
			return
		}
	}

	newUser := models.User{
		ID:              primitive.NewObjectID(),
		Username:        username,
		Password:        hashedPwdString,
		HubUsername:     body.HubUsername,
		HubPassword:     body.HubPassword,
		UserType:        userType,
		CreationDate:    time.Now(),
		AlreadyLoggedIn: false,
		AvatarId:        avatarId,
	}

	_, err = usersCollection.InsertOne(context.TODO(), newUser)
	if err != nil || len(username) == 0 {
		serv.Errorf("CreateUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	serv.Noticef("User " + username + " has been created")
	c.IndentedJSON(200, gin.H{
		"id":                      newUser.ID,
		"username":                username,
		"hub_username":            body.HubUsername,
		"hub_password":            body.HubPassword,
		"user_type":               userType,
		"creation_date":           newUser.CreationDate,
		"already_logged_in":       false,
		"avatar_id":               body.AvatarId,
		"broker_connection_creds": brokerConnectionCreds,
	})
}

func (umh UserMgmtHandler) GetAllUsers(c *gin.Context) {
	type filteredUser struct {
		ID              primitive.ObjectID `json:"id" bson:"_id"`
		Username        string             `json:"username" bson:"username"`
		UserType        string             `json:"user_type" bson:"user_type"`
		CreationDate    time.Time          `json:"creation_date" bson:"creation_date"`
		AlreadyLoggedIn bool               `json:"already_logged_in" bson:"already_logged_in"`
		AvatarId        int                `json:"avatar_id" bson:"avatar_id"`
	}
	var users []filteredUser

	cursor, err := usersCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		serv.Errorf("GetAllUsers error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &users); err != nil {
		serv.Errorf("GetAllUsers error: " + err.Error())
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
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}
	var body models.RemoveUserSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	username := strings.ToLower(body.Username)
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveUser error: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}
	if user.Username == username {
		serv.Warnf("You can't remove your own user")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can't remove your own user"})
		return
	}

	exist, userToRemove, err := IsUserExist(username)
	if err != nil {
		serv.Errorf("RemoveUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("User does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "User does not exist"})
		return
	}
	if userToRemove.UserType == "root" {
		serv.Warnf("You can not remove the root user")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can not remove the root user"})
		return
	}

	err = updateUserResources(userToRemove)
	if err != nil {
		serv.Errorf("RemoveUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	_, err = usersCollection.DeleteOne(context.TODO(), bson.M{"username": username})
	if err != nil {
		serv.Errorf("RemoveUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	serv.Noticef("User " + username + " has been deleted")
	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) RemoveMyUser(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveUser error: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}

	if user.UserType == "root" {
		c.AbortWithStatusJSON(500, gin.H{"message": "Root user can not be deleted"})
		return
	}

	err = updateUserResources(user)
	if err != nil {
		serv.Errorf("RemoveMyUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	_, err = usersCollection.DeleteOne(context.TODO(), bson.M{"username": user.Username})
	if err != nil {
		serv.Errorf("RemoveMyUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	serv.Noticef("User " + user.Username + " has been deleted")
	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) EditHubCreds(c *gin.Context) {
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}
	var body models.EditHubCredsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	err := validateHubCreds(body.HubUsername, body.HubPassword)
	if err != nil {
		serv.Errorf("EditHubCreds error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("EditHubCreds error: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}

	_, err = usersCollection.UpdateOne(context.TODO(),
		bson.M{"username": user.Username},
		bson.M{"$set": bson.M{"hub_username": body.HubUsername, "hub_password": body.HubPassword}},
	)
	if err != nil {
		serv.Errorf("EditHubCreds error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{
		"id":                user.ID,
		"username":          user.Username,
		"hub_username":      body.HubUsername,
		"hub_password":      body.HubPassword,
		"user_type":         user.UserType,
		"creation_date":     user.CreationDate,
		"already_logged_in": user.AlreadyLoggedIn,
		"avatar_id":         user.AvatarId,
	})
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
		serv.Errorf("EditAvatar error: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}

	_, err = usersCollection.UpdateOne(context.TODO(),
		bson.M{"username": user.Username},
		bson.M{"$set": bson.M{"avatar_id": avatarId}},
	)
	if err != nil {
		serv.Errorf("EditAvatar error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{
		"id":                user.ID,
		"username":          user.Username,
		"hub_username":      user.HubUsername,
		"hub_password":      user.HubPassword,
		"user_type":         user.UserType,
		"creation_date":     user.CreationDate,
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
		serv.Errorf("EditCompanyLogo error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	base64Encoding, err := imageToBase64(fileName)
	if err != nil {
		serv.Errorf("EditCompanyLogo error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	_ = os.Remove(fileName)

	newImage := models.Image{
		ID:    primitive.NewObjectID(),
		Name:  "company_logo",
		Image: base64Encoding,
	}

	_, err = imagesCollection.InsertOne(context.TODO(), newImage)
	if err != nil {
		serv.Errorf("EditCompanyLogo error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{"image": base64Encoding})
}

func (umh UserMgmtHandler) RemoveCompanyLogo(c *gin.Context) {
	_, err := imagesCollection.DeleteOne(context.TODO(), bson.M{"name": "company_logo"})
	if err != nil {
		serv.Errorf("RemoveCompanyLogo error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) GetCompanyLogo(c *gin.Context) {
	var image models.Image
	err := imagesCollection.FindOne(context.TODO(), bson.M{"name": "company_logo"}).Decode(&image)
	if err == mongo.ErrNoDocuments {
		c.IndentedJSON(200, gin.H{"image": ""})
		return
	} else if err != nil {
		serv.Errorf("GetCompanyLogo error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{"image": image.Image})
}

func (umh UserMgmtHandler) EditAnalytics(c *gin.Context) {
	if err := DenyForSandboxEnv(c); err != nil {
		return
	}
	var body models.EditAnalyticsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	flag := "false"
	if body.SendAnalytics {
		flag = "true"
	}

	_, err := systemKeysCollection.UpdateOne(context.TODO(),
		bson.M{"key": "analytics"},
		bson.M{"$set": bson.M{"value": flag}},
	)
	if err != nil {
		serv.Errorf("EditAnalytics error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !body.SendAnalytics {
		analytics.IncrementDisableAnalyticsCounter()
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) DoneNextSteps(c *gin.Context) {
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.IncrementNextStepsCounter()
	}

	c.IndentedJSON(200, gin.H{})
}
