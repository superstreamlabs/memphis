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

package server

import (
	"memphis-broker/models"

	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConnectionsHandler struct{}

var connectionsHandler ConnectionsHandler
var producersHandler ProducersHandler
var consumersHandler ConsumersHandler

func  handleConnectMessage(client *client) (error) {
	// d := json.NewDecoder(connection)
	// var message tcpMessage
	// err := d.Decode(&message)
	

	// if err != nil {
	// 	connection.Write([]byte("Memphis protocol error"))
	// 	connection.Close()
	// 	return  nil
	
		username := strings.ToLower(client.opts.Username)
		exist, user, err := IsUserExist(username)
		if err != nil {
			// logger.Error("handleConnectMessage: " + err.Error())
			// connection.Write([]byte("Server error: " + err.Error()))
			// connection.Close()
			return  err
		}
		if !exist {
			// connection.Write([]byte("User is not exist"))
			// connection.Close()
			return errors.New("User is not exist")
		}
		if user.UserType != "root" && user.UserType != "application" {
			// connection.Write([]byte("Please use a user of type Root/Application and not Management"))
			// connection.Close()
			return errors.New("Please use a user of type Root/Application and not Management")
		}
		// connectionId := []byte(connectionId)
		// connectionId = connectionId[:12]
		connectionId := primitive.NewObjectID()
		exist, _, err = IsConnectionExist(primitive.ObjectID(connectionId))
		if err != nil {
			// logger.Error("handleConnectMessage: " + err.Error())
			// connection.Write([]byte("Server error: " + err.Error()))
			// connection.Close()
			return err
		}

		// err = broker.ValidateUserCreds(message.BrokerCreds)
		// if err != nil {
		// 	connection.Write([]byte("Server error: " + err.Error()))
		// 	connection.Close()
		// 	return primitive.ObjectID{}, models.User{}
		// }

		// clientAddress := connection.RemoteAddr()
		clientAddress:= client.host
		// clientAddressString := clientAddress.String()
		clientAddress = strings.Split(clientAddress, ":")[0]
		if exist {
			err = connectionsHandler.ReliveConnection(connectionId)
			if err != nil {
				// logger.Error("handleConnectMessage: " + err.Error())
				// connection.Write([]byte("Server error: " + err.Error()))
				// connection.Close()
				return err
			}
			err = producersHandler.ReliveProducers(connectionId)
			if err != nil {
				// logger.Error("handleConnectMessage: " + err.Error())
				// connection.Write([]byte("Server error: " + err.Error()))
				// connection.Close()
				return err
			}
			err = consumersHandler.ReliveConsumers(connectionId)
			if err != nil {
				// logger.Error("handleConnectMessage: " + err.Error())
				// connection.Write([]byte("Server error: " + err.Error()))
				// connection.Close()
				return err
			}
		} else {
			connectionId, err = connectionsHandler.CreateConnection(username, clientAddress)
			if err != nil {
				// logger.Error("handleConnectMessage: " + err.Error())
				// connection.Write([]byte("Server error: " + err.Error()))
				// connection.Close()
				return err
			}
		}

		// accessToken, err := createAccessToken(user)
		// if err != nil {
		// 	// logger.Error("handleConnectMessage: " + err.Error())
		// 	connection.Write([]byte("Server error: " + err.Error()))
		// 	connection.Close()
		// 	return primitive.ObjectID{}, models.User{}
		// }

		// response := TcpResponseMessage{
		// 	ConnectionId:   connectionId,
		// }
		// bytesResponse, _ := json.Marshal(response)
		// client.nc.Write(bytesResponse)
		return nil
	}


func (ch ConnectionsHandler) CreateConnection(username string, clientAddress string) (primitive.ObjectID, error) {
	connectionId := primitive.NewObjectID()

	username = strings.ToLower(username)
	exist, _, err := IsUserExist(username)
	if err != nil {
		serv.Errorf("CreateConnection error: " + err.Error())
		return connectionId, err
	}
	if !exist {
		return connectionId, errors.New("User was not found")
	}

	newConnection := models.Connection{
		ID:            connectionId,
		CreatedByUser: username,
		IsActive:      true,
		CreationDate:  time.Now(),
		LastPing:      time.Now(),
		ClientAddress: clientAddress,
	}

	_, err = connectionsCollection.InsertOne(context.TODO(), newConnection)
	if err != nil {
		serv.Errorf("CreateConnection error: " + err.Error())
		return connectionId, err
	}
	return connectionId, nil
}

func (ch ConnectionsHandler) KillConnection(connectionId primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateOne(context.TODO(),
		bson.M{"_id": connectionId},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		serv.Errorf("KillConnection error: " + err.Error())
		return err
	}

	return nil
}

func (ch ConnectionsHandler) ReliveConnection(connectionId primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateOne(context.TODO(),
		bson.M{"_id": connectionId},
		bson.M{"$set": bson.M{"is_active": true}},
	)
	if err != nil {
		serv.Errorf("ReliveConnection error: " + err.Error())
		return err
	}

	return nil
}

func (ch ConnectionsHandler) UpdatePingTime(connectionId primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateOne(context.TODO(),
		bson.M{"_id": connectionId},
		bson.M{"$set": bson.M{"last_ping": time.Now()}},
	)
	if err != nil {
		serv.Errorf("UpdatePingTime error: " + err.Error())
		return err
	}

	return nil
}
