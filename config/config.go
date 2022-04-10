package config

import (
	"github.com/gin-gonic/gin"
	"github.com/tkanos/gonfig"
	"os"
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
	BROKER_JWT                     string
}

func GetConfig() Configuration {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "prod"
		gin.SetMode(gin.ReleaseMode)
	} else {
		os.Setenv("GIN_MODE", "debug")
		gin.SetMode(gin.DebugMode)
	}
	configuration := Configuration{}
	gonfig.GetConf("./config/config.json", &configuration)

	return configuration
}
