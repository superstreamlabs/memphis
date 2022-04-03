package main

import (
	"strech-server/broker"
	"strech-server/config"
	"strech-server/db"
	"strech-server/handlers"
	"strech-server/logger"
	"strech-server/routes"
	"strech-server/socketio"
)

func main() {
	configuration := config.GetConfig()
	err := handlers.CreateRootUserOnFirstSystemLoad()
	if err != nil {
		logger.Error("Failed to create root user: " + err.Error())
		panic("Failed to create root user: " + err.Error())
	}
	router := routes.InitializeHttpRoutes()
	socketioServer := socketio.InitializeSocketio(router)

	defer socketioServer.Close()
	defer db.Close()
	defer broker.Close()
	router.Run(":" + configuration.PORT)
}
