package main

import (
	"memphis-control-plane/broker"
	"memphis-control-plane/config"
	"memphis-control-plane/db"
	"memphis-control-plane/handlers"
	"memphis-control-plane/logger"
	"memphis-control-plane/routes"
	"memphis-control-plane/socketio"
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
