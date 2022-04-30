package main

import (
	"memphis-control-plane/broker"
	"memphis-control-plane/db"
	"memphis-control-plane/handlers"
	"memphis-control-plane/http_server"
	"memphis-control-plane/logger"
	"memphis-control-plane/tcp_server"
	"memphis-control-plane/utils"
	"os"
	"sync"
)

func main() {
	err := handlers.CreateRootUserOnFirstSystemLoad()
	if err != nil {
		logger.Error("Failed to create root user: " + err.Error())
		panic("Failed to create root user: " + err.Error())
	}

	defer db.Close()
	defer broker.Close()

	wg := new(sync.WaitGroup)
	wg.Add(3)

	go tcp_server.InitializeTcpServer(wg)
	go http_server.InitializeHttpServer(wg)
	go utils.KillZombieConnections(wg)

	env := os.Getenv("ENVIRONMENT")
	if env == "" && os.Getenv("DOCKER_ENV") != "" {
		env = "Docker"
	} else if env == "" && os.Getenv("DOCKER_ENV") == "" {
		env = "K8S"
	}
	
	logger.Info("Memphis control plane is up and running, ENV: " + env)
	wg.Wait()
}
