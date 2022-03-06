package handlers

import (
	"context"
	"errors"
	"regexp"
	"strech-server/config"
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
	"golang.org/x/crypto/bcrypt"
)

var usersCollection *mongo.Collection = db.GetCollection(db.Client, "users")
var refreshTokensCollection *mongo.Collection = db.GetCollection(db.Client, "refresh_tokens")
var configuration = config.GetConfig()

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
	// atClaims["authorized"] = true
	atClaims["user_id"] = user.ID
	atClaims["username"] = user.Username
	atClaims["hub_username"] = user.HubUsername
	atClaims["hub_password"] = user.HubPassword
	atClaims["user_type"] = user.UserType
	atClaims["creation_date"] = user.CreationDate
	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(configuration.JWT_EXPIRES_IN_MINUTES)).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(configuration.JWT_SECRET))
	if err != nil {
		return "", "", err
	}

	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(configuration.REFRESH_JWT_EXPIRES_IN_MINUTES)).Unix()
	at = jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	refreshToken, err := at.SignedString([]byte(configuration.REFRESH_JWT_SECRET))

	return token, refreshToken, nil
}

// TODO
func (umh UserMgmtHandler) CreateRootUser(c *gin.Context) {
	var body models.CreateRootUserSchema
	ok := utils.Validate(c, &body)
	if !ok {
		return
	}

	exist, err := isRootUserExist()
	if err != nil {
		logger.Error("CreateRootUser error: " + err.Error())
		c.IndentedJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		c.IndentedJSON(400, gin.H{"message": "This account already has root user"})
		return
	}

	username := strings.ToLower(body.Username)
	usernameError := validateUsername(username)
	if usernameError != nil {
		c.IndentedJSON(400, gin.H{"message": usernameError.Error()})
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
	if err != nil {
		logger.Error("CreateUser error: " + err.Error())
		c.IndentedJSON(500, gin.H{"message": "Server error"})
		return
	}
	hashedPwdString := string(hashedPwd)

	// trying to login with the given hub creds if needed and see if they are valid

	newUser := models.User{
		ID:           primitive.NewObjectID(),
		Username:     username,
		Password:     hashedPwdString,
		HubUsername:  body.HubUsername,
		HubPassword:  body.HubPassword,
		UserType:     "root",
		CreationDate: time.Now(),
	}

	_, err = usersCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		logger.Error("CreateUser error: " + err.Error())
		c.IndentedJSON(500, gin.H{"message": "Server error"})
		return
	}

	// login the user

	c.IndentedJSON(200, newUser.GetUserWithoutPassword())
}

func (umh UserMgmtHandler) Login(c *gin.Context) {
	var body models.LoginSchema
	ok := utils.Validate(c, &body)
	if !ok {
		return
	}

	username := strings.ToLower(body.Username)
	authenticated, user, err := authenticateUser(username, body.Password)
	if err != nil {
		logger.Error("Login error: " + err.Error())
		c.IndentedJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !authenticated || user.UserType == "application" {
		c.IndentedJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	token, refreshToken, err := createTokens(user)
	if err != nil {
		logger.Error("Login error: " + err.Error())
		c.IndentedJSON(500, gin.H{"message": "Server error"})
		return
	}

	newRefreshToken := models.RefreshToken{
		ID:           primitive.NewObjectID(),
		UserId:       user.ID,
		RefreshToken: refreshToken,
	}

	_, err = refreshTokensCollection.InsertOne(context.TODO(), newRefreshToken)
	if err != nil {
		logger.Error("Login error: " + err.Error())
		c.IndentedJSON(500, gin.H{"message": "Server error"})
		return
	}

	type Response struct {
		Jwt          string             `json:"jwt"`
		ExpiresIn    int                `json:"expires_in"`
		UserId       primitive.ObjectID `json:"user_id"`
		Username     string             `json:"username"`
		UserType     string             `json:"user_type"`
		CreationDate time.Time          `json:"creation_date"`
	}
	response := Response{
		Jwt:          token,
		ExpiresIn:    configuration.JWT_EXPIRES_IN_MINUTES * 60 * 1000,
		UserId:       user.ID,
		Username:     user.Username,
		UserType:     user.UserType,
		CreationDate: user.CreationDate,
	}

	domain := ".strech.io"
	secure := true
	if configuration.ENVIRONMENT == "dev" {
		domain = "localhost"
		secure = false
	}
	c.SetCookie("jwt-refresh-token", refreshToken, configuration.REFRESH_JWT_EXPIRES_IN_MINUTES * 60 * 1000, "/", domain, secure, true)
	c.IndentedJSON(200, response)
}

func (umh UserMgmtHandler) AddUser(c *gin.Context) {
	var body models.AddUserSchema
	ok := utils.Validate(c, &body)
	if !ok {
		return
	}

	username := strings.ToLower(body.Username)
	exist, err := isUserExist(username)
	if err != nil {
		logger.Error("CreateUser error: " + err.Error())
		c.IndentedJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		c.IndentedJSON(400, gin.H{"message": "A user with this username is already exist"})
		return
	}

	userType := strings.ToLower(body.UserType)
	userTypeError := validateUserType(userType)
	if userTypeError != nil {
		c.IndentedJSON(400, gin.H{"message": userTypeError.Error()})
		return
	}

	usernameError := validateUsername(username)
	if usernameError != nil {
		c.IndentedJSON(400, gin.H{"message": usernameError.Error()})
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
	if err != nil {
		logger.Error("CreateUser error: " + err.Error())
		c.IndentedJSON(500, gin.H{"message": "Server error"})
		return
	}
	hashedPwdString := string(hashedPwd)

	newUser := models.User{
		ID:           primitive.NewObjectID(),
		Username:     username,
		Password:     hashedPwdString,
		HubUsername:  body.HubUsername,
		HubPassword:  body.HubPassword,
		UserType:     userType,
		CreationDate: time.Now(),
	}

	_, err = usersCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		logger.Error("CreateUser error: " + err.Error())
		c.IndentedJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, newUser.GetUserWithoutPassword())
}

func (umh UserMgmtHandler) AuthenticateNats(c *gin.Context) {
	var body models.AuthenticateNatsSchema
	ok := utils.Validate(c, &body)
	if !ok {
		return
	}

	authenticated, user, err := authenticateUser(body.Username, body.Password)
	if err != nil {
		logger.Error("AuthenticateNats error: " + err.Error())
		c.IndentedJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !authenticated || user.UserType == "management" {
		c.IndentedJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	c.IndentedJSON(200, "")
}
