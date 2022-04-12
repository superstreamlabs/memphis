package http_server

import (
	"memphis-control-plane/config"
	"memphis-control-plane/http_server/routes"
	"memphis-control-plane/socketio"
	"sync"
)

func InitializeHttpServer(wg *sync.WaitGroup) {
	configuration := config.GetConfig()

	httpServer := routes.InitializeHttpRoutes()
	socketioServer := socketio.InitializeSocketio(httpServer)
	
	defer socketioServer.Close()
	defer wg.Done()

	httpServer.Run(":" + configuration.HTTP_PORT)
}