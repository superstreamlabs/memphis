package tcp_server

import (
	"context"
	"encoding/json"
	"memphis-control-plane/broker"
	"memphis-control-plane/config"
	"memphis-control-plane/db"
	"memphis-control-plane/handlers"
	"memphis-control-plane/logger"
	"memphis-control-plane/models"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type tcpMessage struct {
	Username          string             `json:"username"`
	BrokerCreds       string             `json:"broker_creds"`
	ConnectionId      primitive.ObjectID `json:"connection_id"`
	ResendAccessToken bool               `json:"resend_access_token"`
}

type tcpResponseMessage struct {
	ConnectionId   primitive.ObjectID `json:"connection_id"`
	AccessToken    string             `json:"access_token"`
	AccessTokenExp int                `json:"access_token_exp"`
}

var configuration = config.GetConfig()
var connectionsHandler handlers.ConnectionsHandler
var producersHandler handlers.ProducersHandler
var consumersHandler handlers.ConsumersHandler

func createAccessToken(user models.User) (string, error) {
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
		return "", err
	}

	var tokensCollection *mongo.Collection = db.GetCollection("tokens")
	opts := options.Update().SetUpsert(true)
	_, err = tokensCollection.UpdateOne(context.TODO(),
		bson.M{"username": user.Username},
		bson.M{"$set": bson.M{"jwt_token": token}},
		opts,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

func handleConnectMessage(connection net.Conn) (primitive.ObjectID, models.User) {
	d := json.NewDecoder(connection)
	var message tcpMessage
	err := d.Decode(&message)
	if err != nil {
		connection.Write([]byte("Memphis protocol error"))
		connection.Close()
		return primitive.ObjectID{}, models.User{}
	} else {
		username := strings.ToLower(message.Username)
		exist, user, err := handlers.IsUserExist(username)
		if err != nil {
			logger.Error("handleConnectMessage: " + err.Error())
			connection.Write([]byte("Server error: " + err.Error()))
			connection.Close()
			return primitive.ObjectID{}, models.User{}
		}
		if !exist {
			connection.Write([]byte("User is not exist"))
			connection.Close()
			return primitive.ObjectID{}, models.User{}
		}
		if user.UserType != "application" {
			connection.Write([]byte("You have to connect with application type user"))
			connection.Close()
			return primitive.ObjectID{}, models.User{}
		}

		connectionId := message.ConnectionId
		exist, _, err = handlers.IsConnectionExist(connectionId)
		if err != nil {
			logger.Error("handleConnectMessage: " + err.Error())
			connection.Write([]byte("Server error: " + err.Error()))
			connection.Close()
			return primitive.ObjectID{}, models.User{}
		}

		err = broker.ValidateUserCreds(message.BrokerCreds)
		if err != nil {
			connection.Write([]byte("Server error: " + err.Error()))
			connection.Close()
			return primitive.ObjectID{}, models.User{}
		}

		if exist {
			err = connectionsHandler.ReliveConnection(connectionId)
			if err != nil {
				logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
				return primitive.ObjectID{}, models.User{}
			}
			err = producersHandler.ReliveProducers(connectionId)
			if err != nil {
				logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
				return primitive.ObjectID{}, models.User{}
			}
			err = consumersHandler.ReliveConsumers(connectionId)
			if err != nil {
				logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
				return primitive.ObjectID{}, models.User{}
			}
		} else {
			connectionId, err = connectionsHandler.CreateConnection(username)
			if err != nil {
				logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
				return primitive.ObjectID{}, models.User{}
			}
		}

		accessToken, err := createAccessToken(user)
		if err != nil {
			logger.Error("handleConnectMessage: " + err.Error())
			connection.Write([]byte("Server error: " + err.Error()))
			connection.Close()
			return primitive.ObjectID{}, models.User{}
		}

		response := tcpResponseMessage{
			ConnectionId:   connectionId,
			AccessToken:    accessToken,
			AccessTokenExp: configuration.JWT_EXPIRES_IN_MINUTES * 60 * 1000,
		}
		bytesResponse, _ := json.Marshal(response)
		connection.Write(bytesResponse)
		return connectionId, user
	}
}

func removeConnectionResources(connectionId primitive.ObjectID) error {
	err := connectionsHandler.RemoveConnection(connectionId)
	if err != nil {
		return err
	}
	err = producersHandler.RemoveProducers(connectionId)
	if err != nil {
		return err
	}
	err = consumersHandler.RemoveConsumers(connectionId)
	if err != nil {
		return err
	}

	return nil
}

func handleNewClient(connection net.Conn) {
	logger.Info("A new client connection has been established: " + connection.RemoteAddr().String())
	connectionId, user := handleConnectMessage(connection)
	if !connectionId.IsZero() {
		for {
			d := json.NewDecoder(connection)
			var message tcpMessage
			err := d.Decode(&message)
			if err != nil {
				err = removeConnectionResources(connectionId)
				if err != nil {
					logger.Error("handleNewClient error: " + err.Error())
				}
				break
			}

			if message.ResendAccessToken {
				accessToken, err := createAccessToken(user)
				if err != nil {
					logger.Error("handleNewClient error: " + err.Error())
					break
				}

				response := tcpResponseMessage{
					ConnectionId:   connectionId,
					AccessToken:    accessToken,
					AccessTokenExp: configuration.JWT_EXPIRES_IN_MINUTES * 60 * 1000,
				}
				bytesResponse, _ := json.Marshal(response)
				connection.Write(bytesResponse)
			}
		}
	}
	connection.Close()
}

func InitializeTcpServer(wg *sync.WaitGroup) {
	tcpServer, err := net.Listen("tcp4", ":"+configuration.TCP_PORT)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	defer tcpServer.Close()
	defer wg.Done()

	for {
		connection, err := tcpServer.Accept()
		if err != nil {
			logger.Error("InitializeTcpServer: " + err.Error())
		} else {
			go handleNewClient(connection)
		}
	}
}
