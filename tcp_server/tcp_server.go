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

func handleConnectMessage(connection net.Conn) {
	d := json.NewDecoder(connection)
	var message tcpMessage
	err := d.Decode(&message)
	if err != nil || message.Operation != "connect" {
		connection.Write([]byte("Memphis protocol error"))
		connection.Close()
	} else {
		username := strings.ToLower(message.Username)
		exist, user, err := handlers.IsUserExist(username)
		if err != nil {
			logger.Error("handleConnectMessage: " + err.Error())
			connection.Write([]byte("Server error: " + err.Error()))
			connection.Close()
		}
		if !exist {
			connection.Write([]byte("User is not exist"))
			connection.Close()
		}
		if user.UserType != "application" {
			connection.Write([]byte("You have to connect with application type user"))
			connection.Close()
		}

		connectionId := message.ConnectionId
		exist, _, err = handlers.IsConnectionExist(connectionId)
		if err != nil {
			logger.Error("handleConnectMessage: " + err.Error())
			connection.Write([]byte("Server error: " + err.Error()))
			connection.Close()
		}

		err = broker.ValidateUserCreds(message.Jwt)
		if err != nil {
			logger.Error("handleConnectMessage: " + err.Error())
			connection.Write([]byte("Server error: " + err.Error()))
			connection.Close()
		}

		if exist {
			err = connectionsHandler.ReliveConnection(connectionId)
			if err != nil {
				logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
			}
			err = producersHandler.ReliveProducers(connectionId)
			if err != nil {
				logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
			}
			err = consumersHandler.ReliveConsumers(connectionId)
			if err != nil {
				logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
			}
		} else {
			connectionId, err = connectionsHandler.CreateConnection(username)
			if err != nil {
				logger.Error("handleConnectMessage: " + err.Error())
				connection.Write([]byte("Server error: " + err.Error()))
				connection.Close()
			}
		}

		connection.Write([]byte(connectionId.String()))
	}
}

func handleCreateProducerMessage() {

}

func handleCreateConsumerMessage() {

}

func handleNewClient(connection net.Conn) {
	logger.Info("A new client connection has been established: " + connection.RemoteAddr().String())
	handleConnectMessage(connection)

acceptMessagesLoop:
	for {
		d := json.NewDecoder(connection)
		var message tcpMessage
		err := d.Decode(&message)
		if err != nil {
			connection.Write([]byte("Memphis protocol error"))
			// set connection + producers + consumers is active to false
			break
		}
		switch message.Operation {
		case "CreateProducer":
			logger.Info("CreateProducer")
			handleCreateProducerMessage(connection)
			// create producer in db
			break
		case "CreateConsumer":
			handleCreateConsumerMessage(connection)
			break
		default:
			connection.Write([]byte("Memphis protocol error"))
			break acceptMessagesLoop
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
