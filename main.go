package main

import (
	"memphis-server/broker"
	"memphis-server/config"
	"memphis-server/db"
	"memphis-server/handlers"
	"memphis-server/logger"
	"memphis-server/routes"
	"memphis-server/socketio"
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
