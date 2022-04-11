package tcp_server

import (
	"memphis-control-plane/config"
	"memphis-control-plane/logger"
	"net"
	"strings"
	"sync"
)

func handleNewClient(connection net.Conn) {
	logger.Info("A new client connection has been established: " + connection.RemoteAddr().String())
	for {
		buf := make([]byte, 1024)
		_, err := connection.Read(buf)
		if err != nil {
			connection.Write([]byte("Memphis protocol error"))
		} else {
			// on connect
			// get username
			// get nats jwt
			// create connection id
			// create connection in the db if connection id not already exist

			// on create producer
			// get connection id
			// create producer in db

			// when connection lost
			// remove connection + producers + consumers of the same connection id


			message := string(buf)
			message = message[:strings.IndexByte(message, '\n')]
			if message == "STOP" {
				break
			}
			connection.Write([]byte("Ok"))
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
