// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tcp_server

import (
	"encoding/json"
	// "errors"
	"memphis-broker/conf"
	"memphis-broker/handlers"
	"memphis-broker/models"
	"net"
	"strings"
	// "time"
	// "github.com/dgrijalva/jwt-go"
	"memphis-broker/server"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var serv *server.Server


type tcpMessage struct {
	Username          string             `json:"username"`
	BrokerCreds       string             `json:"brokercreds"`
	ConnectionId      primitive.ObjectID `json:"connection_id"`
}

type TcpResponseMessage struct {
	ConnectionId   primitive.ObjectID `json:"connection_id"`
}

var configuration = conf.GetConfig()
var connectionsHandler handlers.ConnectionsHandler
var producersHandler handlers.ProducersHandler
var consumersHandler handlers.ConsumersHandler

// func createAccessToken(user models.User) (string, error) {
// 	username := strings.ToLower(user.Username)
// 	exist, _, err := handlers.IsUserExist(username)
// 	if err != nil {
// 		return "", err
// 	}
// 	if !exist {
// 		return "", errors.New("user does not exist")
// 	}

// 	atClaims := jwt.MapClaims{}
// 	atClaims["user_id"] = user.ID.Hex()
// 	atClaims["username"] = username
// 	atClaims["user_type"] = user.UserType
// 	atClaims["creation_date"] = user.CreationDate
// 	atClaims["already_logged_in"] = user.AlreadyLoggedIn
// 	atClaims["avatar_id"] = user.AvatarId
// 	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(configuration.JWT_EXPIRES_IN_MINUTES)).Unix()
// 	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
// 	token, err := at.SignedString([]byte(configuration.JWT_SECRET))
// 	if err != nil {
// 		return "", err
// 	}

// 	return token, nil
// }

func handleConnectMessage(connection net.Conn) (TcpResponseMessage, models.User) {
	d := json.NewDecoder(connection)
	var message tcpMessage
	err := d.Decode(&message)
	if err != nil {
		connection.Write([]byte("Memphis protocol error"))
		connection.Close()
		return TcpResponseMessage{}, models.User{}
	} else {
		username := strings.ToLower(message.Username)
		exist, user, err := handlers.IsUserExist(username)
		if err != nil {
			// logger.Error("handleConnectMessage: " + err.Error())
			connection.Write([]byte("Server error: " + err.Error()))
			connection.Close()
			return TcpResponseMessage{}, models.User{}
		}
		if !exist {
			connection.Write([]byte("User is not exist"))
			connection.Close()
			return TcpResponseMessage{}, models.User{}
		}
		if user.UserType != "root" && user.UserType != "application" {
			connection.Write([]byte("Please use a user of type Root/Application and not Management"))
			connection.Close()
			return TcpResponseMessage{}, models.User{}
		}

		connectionId := message.ConnectionId
		exist, _, err = handlers.IsConnectionExist(connectionId)
		if err != nil {
			// logger.Error("handleConnectMessage: " + err.Error())
			connection.Write([]byte("Server error: " + err.Error()))
			connection.Close()
			return TcpResponseMessage{}, models.User{}
		}

		// err = broker.ValidateUserCreds(message.BrokerCreds)
		// if err != nil {
		// 	connection.Write([]byte("Server error: " + err.Error()))
		// 	connection.Close()
		// 	return primitive.ObjectID{}, models.User{}
		// }

		clientAddress := connection.RemoteAddr()
		clientAddressString := clientAddress.String()
		clientAddressString = strings.Split(clientAddressString, ":")[0]
		if exist {
			err = connectionsHandler.ReliveConnection(connectionId)
			if err != nil {
				// logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
				return TcpResponseMessage{}, models.User{}
			}
			err = producersHandler.ReliveProducers(connectionId)
			if err != nil {
				// logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
				return TcpResponseMessage{}, models.User{}
			}
			err = consumersHandler.ReliveConsumers(connectionId)
			if err != nil {
				// logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
				return TcpResponseMessage{}, models.User{}
			}
		} else {
			connectionId, err = connectionsHandler.CreateConnection(username, clientAddressString)
			if err != nil {
				// logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
				return TcpResponseMessage{}, models.User{}
			}
		}

		// accessToken, err := createAccessToken(user)
		// if err != nil {
		// 	// logger.Error("handleConnectMessage: " + err.Error())
		// 	connection.Write([]byte("Server error: " + err.Error()))
		// 	connection.Close()
		// 	return primitive.ObjectID{}, models.User{}
		// }

		response := TcpResponseMessage{
			ConnectionId:   connectionId,
		}
		// bytesResponse, _ := json.Marshal(response)
		// connection.Write(bytesResponse)
		return response, user
	}
}

// func killConnectionResources(connectionId primitive.ObjectID) error {
// 	err := connectionsHandler.KillConnection(connectionId)
// 	if err != nil {
// 		return err
// 	}
// 	err = producersHandler.KillProducers(connectionId)
// 	if err != nil {
// 		return err
// 	}
// 	err = consumersHandler.KillConsumers(connectionId)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func StartTcpServer (serv *server.Server) error{
	serv.Start(handleNewClient)
	return nil
}

func handleNewClient(connection net.Conn) (primitive.ObjectID) {
	// logger.Info("A new client connection has been established: " + connection.RemoteAddr().String())
	TcpResponseMessage, _ := handleConnectMessage(connection)
	return TcpResponseMessage.ConnectionId
	// if !connectionId.IsZero() {
	// 	for {
	// 		d := json.NewDecoder(connection)
	// 		var message tcpMessage
	// 		err := d.Decode(&message)
	// 		if err != nil {
	// 			err = killConnectionResources(connectionId)
	// 			if err != nil {
	// 				// logger.Error("handleNewClient error: " + err.Error())
	// 			}
	// 			break
	// 		}

			// if message.ResendAccessToken {
			// 	accessToken, err := createAccessToken(user)
			// 	if err != nil {
			// 		// logger.Error("handleNewClient error: " + err.Error())
			// 		break
			// 	}

				// response := tcpResponseMessage{
				// 	ConnectionId:   connectionId,
				// 	// AccessToken:    accessToken,
				// 	// AccessTokenExp: configuration.JWT_EXPIRES_IN_MINUTES * 60 * 1000,
				// }
				// bytesResponse, _ := json.Marshal(response)
				// connection.Write(bytesResponse)
			// } 
			// else if message.Ping {
			// 	err = connectionsHandler.UpdatePingTime(connectionId)
			// 	if err != nil {
			// 		// logger.Error("handleNewClient error: " + err.Error())
			// 	}
			// 	response := tcpResponseMessage{
			// 		ConnectionId: connectionId,
			// 		PingInterval: configuration.PING_INTERVAL_MS,
			// 	}
			// 	bytesResponse, _ := json.Marshal(response)
			// 	connection.Write(bytesResponse)
			// }
		}
	// }
	// connection.Close()
// }

// func InitializeTcpServer(wg *sync.WaitGroup) {
	// tcpServer, err := net.Listen("tcp4", ":"+configuration.TCP_PORT)
	// if err != nil {
	// 	// logger.Error("Failed initializing the TCP server " + err.Error())
	// 	panic("Failed initializing the TCP server " + err.Error())
	// }
	// defer tcpServer.Close()
	// defer wg.Done()

	// for {
	// 	connection, err := tcpServer.Accept()
	// 	if err != nil {
	// 		// logger.Error("Failed to establish TCP connection: " + err.Error())
	// 	} else {
	// 		go handleNewClient(connection)
	// 	}
	// }
// }