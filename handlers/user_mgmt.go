package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"mime/multipart"

	"os"
	"path/filepath"
	"regexp"
	"strech-server/db"
	"strech-server/logger"
	"strech-server/models"
	"strech-server/utils"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var tokensCollection *mongo.Collection = db.GetCollection(db.Client, "tokens")

type UserMgmtHandler struct{}

func isUserExist(username string) (bool, error) {
	filter := bson.M{"username": username}
	var user models.User
	err := usersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
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

func validateUsername(username string) error {
	re := regexp.MustCompile("^[a-z0-9_.]*$")

	validName := re.MatchString(username)
	if !validName {
		return errors.New("username has to include only letters, numbers and _")
	}
	return nil
}

func createTokens(user models.User) (string, string, error) {
	atClaims := jwt.MapClaims{}
	atClaims["user_id"] = user.ID.Hex()
	atClaims["username"] = user.Username
	atClaims["user_type"] = user.UserType
	atClaims["creation_date"] = user.CreationDate
	atClaims["already_logged_in"] = user.AlreadyLoggedIn
	atClaims["avatar_id"] = user.AvatarId
	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(configuration.JWT_EXPIRES_IN_MINUTES)).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
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

func getCompanyLogoPath() (string, error) {
	files, err := ioutil.ReadDir("/tmp/strech")
	if err != nil {
		return "", err
	}

	for _, file := range files {
		fileName := file.Name()
		if strings.HasPrefix(fileName, "company_logo") {
			return "/tmp/strech/" + fileName, nil
		}
	}

	return "", nil
}

// TODO
func (umh UserMgmtHandler) CreateRootUser(c *gin.Context) {
	var body models.CreateRootUserSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	exist, err := isRootUserExist()
	if err != nil {
		logger.Error("CreateRootUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		c.AbortWithStatusJSON(400, gin.H{"message": "This account already has root user"})
		return
	}

	username := strings.ToLower(body.Username)
	usernameError := validateUsername(username)
	if usernameError != nil {
		c.AbortWithStatusJSON(400, gin.H{"message": usernameError.Error()})
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
	if err != nil {
		logger.Error("CreateUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	hashedPwdString := string(hashedPwd)

	// trying to login with the given hub creds if needed and see if they are valid

	newUser := models.User{
		ID:              primitive.NewObjectID(),
		Username:        username,
		Password:        hashedPwdString,
		HubUsername:     body.HubUsername,
		HubPassword:     body.HubPassword,
		UserType:        "root",
		CreationDate:    time.Now(),
		AlreadyLoggedIn: false,
		AvatarId:        body.AvatarId,
	}

	_, err = usersCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		logger.Error("CreateUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	// login the user

	c.IndentedJSON(200, gin.H{
		"id":                newUser.ID,
		"username":          username,
		"hub_username":      body.HubUsername,
		"hub_password":      body.HubPassword,
		"user_type":         "root",
		"creation_date":     newUser.CreationDate,
		"already_logged_in": false,
		"avatar_id":         body.AvatarId,
	})
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
		logger.Error("Login error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !authenticated || user.UserType == "application" {
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	token, refreshToken, err := createTokens(user)
	if err != nil {
		logger.Error("Login error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

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

	domain := ".strech.io"
	secure := true
	if configuration.ENVIRONMENT == "dev" {
		domain = "localhost"
		secure = false
	}
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

func (umh UserMgmtHandler) RefreshToken(c *gin.Context) {
	user := getUserDetailsFromMiddleware(c)
	token, refreshToken, err := createTokens(user)
	if err != nil {
		logger.Error("RefreshToken error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	opts := options.Update().SetUpsert(true)
	_, err = tokensCollection.UpdateOne(context.TODO(),
		bson.M{"username": user.Username},
		bson.M{"$set": bson.M{"jwt_token": token, "refresh_token": refreshToken}},
		opts,
	)
	if err != nil {
		logger.Error("RefreshToken error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	domain := ".strech.io"
	secure := true
	if configuration.ENVIRONMENT == "dev" {
		domain = "localhost"
		secure = false
	}
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

func (umh UserMgmtHandler) Logout(c *gin.Context) {
	user := getUserDetailsFromMiddleware(c)
	_, err := tokensCollection.DeleteOne(context.TODO(), bson.M{"username": user.Username})
	if err != nil {
		logger.Error("Logout error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) AuthenticateNats(c *gin.Context) {
	var body models.AuthenticateNatsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	authenticated, user, err := authenticateUser(body.Username, body.Password)
	if err != nil {
		logger.Error("AuthenticateNats error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !authenticated || user.UserType == "management" {
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	c.IndentedJSON(200, gin.H{})
}

// TODO
func (umh UserMgmtHandler) AddUser(c *gin.Context) {
	var body models.AddUserSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	username := strings.ToLower(body.Username)
	exist, err := isUserExist(username)
	if err != nil {
		logger.Error("CreateUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		c.AbortWithStatusJSON(400, gin.H{"message": "A user with this username is already exist"})
		return
	}

	userType := strings.ToLower(body.UserType)
	userTypeError := validateUserType(userType)
	if userTypeError != nil {
		c.AbortWithStatusJSON(400, gin.H{"message": userTypeError.Error()})
		return
	}

	usernameError := validateUsername(username)
	if usernameError != nil {
		c.AbortWithStatusJSON(400, gin.H{"message": usernameError.Error()})
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
	if err != nil {
		logger.Error("CreateUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	hashedPwdString := string(hashedPwd)

	// trying to login with the given hub creds if needed and see if they are valid

	newUser := models.User{
		ID:              primitive.NewObjectID(),
		Username:        username,
		Password:        hashedPwdString,
		HubUsername:     body.HubUsername,
		HubPassword:     body.HubPassword,
		UserType:        userType,
		CreationDate:    time.Now(),
		AlreadyLoggedIn: false,
		AvatarId:        body.AvatarId,
	}

	_, err = usersCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		logger.Error("CreateUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{
		"id":                newUser.ID,
		"username":          username,
		"hub_username":      body.HubUsername,
		"hub_password":      body.HubPassword,
		"user_type":         userType,
		"creation_date":     newUser.CreationDate,
		"already_logged_in": false,
		"avatar_id":         body.AvatarId,
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
		logger.Error("GetAllUsers error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if err = cursor.All(context.TODO(), &users); err != nil {
		logger.Error("GetAllUsers error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(users) == 0 {
		c.IndentedJSON(200, []string{})
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

	user := getUserDetailsFromMiddleware(c)
	if user.Username == body.Username {
		c.AbortWithStatusJSON(400, gin.H{"message": "You can't remove yourself"})
		return
	}

	_, err := usersCollection.DeleteOne(context.TODO(), bson.M{"username": body.Username})
	if err != nil {
		logger.Error("RemoveUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	_, err = tokensCollection.DeleteOne(context.TODO(), bson.M{"username": body.Username})
	if err != nil {
		logger.Error("RemoveUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) RemoveMyUser(c *gin.Context) {
	user := getUserDetailsFromMiddleware(c)

	_, err := usersCollection.DeleteOne(context.TODO(), bson.M{"username": user.Username})
	if err != nil {
		logger.Error("RemoveMyUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	_, err = tokensCollection.DeleteOne(context.TODO(), bson.M{"username": user.Username})
	if err != nil {
		logger.Error("RemoveMyUser error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) EditHubCreds(c *gin.Context) {
	var body models.EditHubCredsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user := getUserDetailsFromMiddleware(c)
	_, err := usersCollection.UpdateOne(context.TODO(),
		bson.M{"username": user.Username},
		bson.M{"$set": bson.M{"hub_username": body.HubUsername, "hub_password": body.HubPassword}},
	)
	if err != nil {
		logger.Error("EditHubCreds error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	// try to login with the given hub creds

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

func (umh UserMgmtHandler) EditCompanyLogo(c *gin.Context) {
	var file multipart.FileHeader
	ok := utils.Validate(c, nil, true, &file)
	if !ok {
		return
	}

	directoryPath := "/tmp/strech"
	if _, err := os.Stat(directoryPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(directoryPath, os.ModePerm)
		if err != nil {
			logger.Error("EditCompanyLogo error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	fileName := "company_logo" + filepath.Ext(file.Filename)
	saveAtPath := "/tmp/strech/" + fileName
	if err := c.SaveUploadedFile(&file, saveAtPath); err != nil {
		logger.Error("EditCompanyLogo error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	base64Encoding, err := imageToBase64(saveAtPath)
	if err != nil {
		logger.Error("EditCompanyLogo error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{"image": base64Encoding})
}

func (umh UserMgmtHandler) RemoveCompanyLogo(c *gin.Context) {
	path, err := getCompanyLogoPath()
	if err != nil {
		logger.Error("RemoveCompanyLogo error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	err = os.Remove(path)
	if err != nil {
		logger.Error("RemoveCompanyLogo error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) GetCompanyLogo(c *gin.Context) {
	path, err := getCompanyLogoPath()
	if err != nil {
		logger.Error("GetCompanyLogo error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	base64Encoding, err := imageToBase64(path)
	if err != nil {
		logger.Error("GetCompanyLogo error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{"image": base64Encoding})
}
