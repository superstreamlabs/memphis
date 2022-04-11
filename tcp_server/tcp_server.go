package tcp_server

import (
	"encoding/json"
	"memphis-control-plane/broker"
	"memphis-control-plane/config"
	"memphis-control-plane/handlers"
	"memphis-control-plane/logger"
	"net"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type tcpMessage struct {
	Operation    string             `json:"operation"`
	Username     string             `json:"username"`
	Jwt          string             `json:"jwt"`
	ConnectionId primitive.ObjectID `json:"connection_id"`
}

var connectionsHandler handlers.ConnectionsHandler
var producersHandler handlers.ProducersHandler
var consumersHandler handlers.ConsumersHandler

func handleConnectMessage(connection net.Conn) primitive.ObjectID {
	var connectionId primitive.ObjectID
	d := json.NewDecoder(connection)
	var message tcpMessage
	err := d.Decode(&message)
	if err != nil || message.Operation != "connect" {
		connection.Write([]byte("Memphis protocol error"))
		connection.Close()
		return connectionId
	} else {
		username := strings.ToLower(message.Username)
		exist, user, err := handlers.IsUserExist(username)
		if err != nil {
			logger.Error("handleConnectMessage: " + err.Error())
			connection.Write([]byte("Server error: " + err.Error()))
			connection.Close()
			return connectionId
		}
		if !exist {
			connection.Write([]byte("User is not exist"))
			connection.Close()
			return connectionId
		}
		if user.UserType != "application" {
			connection.Write([]byte("You have to connect with application type user"))
			connection.Close()
			return connectionId
		}

		connectionId := message.ConnectionId
		exist, _, err = handlers.IsConnectionExist(connectionId)
		if err != nil {
			logger.Error("handleConnectMessage: " + err.Error())
			connection.Write([]byte("Server error: " + err.Error()))
			connection.Close()
			return connectionId
		}

		err = broker.ValidateUserCreds(message.Jwt)
		if err != nil {
			logger.Error("handleConnectMessage: " + err.Error())
			connection.Write([]byte("Server error: " + err.Error()))
			connection.Close()
			return connectionId
		}

		if exist {
			err = connectionsHandler.ReliveConnection(connectionId)
			if err != nil {
				logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
				return connectionId
			}
			err = producersHandler.ReliveProducers(connectionId)
			if err != nil {
				logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
				return connectionId
			}
			err = consumersHandler.ReliveConsumers(connectionId)
			if err != nil {
				logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
				return connectionId
			}
		} else {
			connectionId, err = connectionsHandler.CreateConnection(username)
			if err != nil {
				logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
				return connectionId
			}
		}

		connection.Write([]byte(connectionId.String()))
		return connectionId
	}
}

func handleCreateProducerMessage() {

}

func handleCreateConsumerMessage() {

}

func handleNewClient(connection net.Conn) {
	logger.Info("A new client connection has been established: " + connection.RemoteAddr().String())
	connectionId := handleConnectMessage(connection)
	if !connectionId.IsZero() {
	acceptMessagesLoop:
		for {
			d := json.NewDecoder(connection)
			var message tcpMessage
			err := d.Decode(&message)
			if err != nil {
				connection.Write([]byte("Memphis protocol error"))

				// only on connection lost
				err = connectionsHandler.RemoveConnection(connectionId)
				if err != nil {
					connection.Write([]byte("Server error: " + err.Error()))
					connection.Close()
					break
				}
				err = producersHandler.RemoveProducers(connectionId)
				if err != nil {
					connection.Write([]byte("Server error: " + err.Error()))
					connection.Close()
					break
				}
				err = consumersHandler.RemoveConsumers(connectionId)
				if err != nil {
					connection.Write([]byte("Server error: " + err.Error()))
					connection.Close()
					break
				}

				break
			}
			switch message.Operation {
			case "CreateProducer":
				// producersHandler.CreateProducer()
				handleCreateProducerMessage()
				// create producer in db
				break
			case "CreateConsumer":
				handleCreateConsumerMessage()
				break
			default:
				connection.Write([]byte("Memphis protocol error"))
				break acceptMessagesLoop
			}
		}
	}
	connection.Close()
}

func InitializeTcpServer(wg *sync.WaitGroup) {
	configuration := config.GetConfig()
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
			logger.Error(err.Error())
		} else {
			go handleNewClient(connection)
		}
	}
}
