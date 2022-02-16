package main

import (
	"strech-server/logger"
	"strech-server/config"
	"strech-server/routes"
	"strech-server/socketio"
	// "strech-server/db"
)

func main() {
	configuration := config.GetConfig()
	logger.Info("Environment: " + configuration.ENVIRONMENT)

	router := routes.InitializeHttpRoutes()
	socketioServer := socketio.InitializeSocketio(router)
	
	defer socketioServer.Close()
	// defer db.Close(db.Client, db.Ctx, db.Cancel)
	router.Run(":" + configuration.PORT)
}
