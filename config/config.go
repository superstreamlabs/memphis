package config

import (
	"encoding/base64"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tkanos/gonfig"
)

type Configuration struct {
	ENVIRONMENT                    string
	HTTP_PORT                      string
	TCP_PORT                       string
	MONGO_URL                      string
	MONGO_USER                     string
	MONGO_PASS                     string
	DB_NAME                        string
	JWT_SECRET                     string
	JWT_EXPIRES_IN_MINUTES         int
	REFRESH_JWT_SECRET             string
	REFRESH_JWT_EXPIRES_IN_MINUTES int
	ROOT_PASSWORD                  string
	BROKER_URL                     string
	CONNECTION_TOKEN               string
	MAX_MESSAGE_SIZE_MB            int
	SHOWABLE_ERROR_STATUS_CODE     int
	DOCKER_ENV                     string
}

func GetConfig() Configuration {
	configuration := Configuration{}
	if os.Getenv("DOCKER_ENV") != "" {
		gonfig.GetConf("./config/docker-config.json", &configuration)
	} else {
		gonfig.GetConf("./config/config.json", &configuration)
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "prod"
		gin.SetMode(gin.ReleaseMode)
	} else {
		os.Setenv("GIN_MODE", "debug")
		gin.SetMode(gin.DebugMode)
		token, _ := base64.StdEncoding.DecodeString(configuration.CONNECTION_TOKEN)
		configuration.CONNECTION_TOKEN = string(token)
	}

	return configuration
}
