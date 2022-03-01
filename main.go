package main

import (
	"strech-server/config"
	"strech-server/db"
	"strech-server/logger"
	"strech-server/routes"
	"strech-server/socketio"
)

func main() {
	configuration := config.GetConfig()
	logger.Info("Environment: " + configuration.ENVIRONMENT)

	router := routes.InitializeHttpRoutes()
	socketioServer := socketio.InitializeSocketio(router)

	defer socketioServer.Close()
	defer db.Close(db.Client, db.Ctx, db.Cancel)
	router.Run(":" + configuration.PORT)
}
