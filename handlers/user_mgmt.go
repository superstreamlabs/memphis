package handlers

import (
	"context"
	"errors"
	"regexp"
	"strech-server/db"
	"strech-server/logger"
	"strech-server/models"
	"strech-server/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var usersCollection *mongo.Collection = db.GetCollection(db.Client, "users")

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

func validateUserType(userType string) error {
	if userType != "application" && userType != "management" {
		return errors.New("user type has to be application/management")
	}
	return nil
}

func validateUsername(username string) error {
	re := regexp.MustCompile("^[a-z0-9_]*$")
	
	validName := re.MatchString(username)
	if !validName {
		return errors.New("username has to include only letters, numbers and _")
	}
	return nil
}

func (ah UserMgmtHandler) CreateUser(c *gin.Context) {
	var body models.User
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
		c.IndentedJSON(400,  gin.H{"message": "A user with this username is already exist"})
		return
	}

	userType := strings.ToLower(body.UserType)
	userTypeError := validateUserType(userType)
	if userTypeError != nil {
		c.IndentedJSON(400,  gin.H{"message": userTypeError.Error()})
		return
	}

	usernameError := validateUsername(username)
	if usernameError != nil {
		c.IndentedJSON(400,  gin.H{"message": usernameError.Error()})
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
	if err != nil {
		logger.Error("CreateUser error: " + err.Error())
		c.IndentedJSON(500, gin.H{"message": "Server error"})
		return
	}
	hashedPwdString := string(hashedPwd)
	// fff := bcrypt.CompareHashAndPassword(hashedPwd, []byte(body.Password))

	// create a user in NATS

	newUser := models.User{
		ID: primitive.NewObjectID(),
		Username: username,
		Password: hashedPwdString,
		HubUsername: body.HubUsername,
		HubPassword: body.HubPassword,
		UserType: userType,
		CreationDate: time.Now(),
	}

	_, err = usersCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		logger.Error("CreateUser error: " + err.Error())
		c.IndentedJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, gin.H{
		"ID": primitive.NewObjectID(),
		"Username": username,
		"HubUsername": body.HubUsername,
		"HubPassword": body.HubPassword,
		"UserType": userType,
		"CreationDate": time.Now(),
	})
}
