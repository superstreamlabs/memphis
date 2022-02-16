package config

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tkanos/gonfig"
	"os"
)

type Configuration struct {
	ENVIRONMENT string
	PORT        string
	LOGGER      string
	MONGO_URL   string
	MONGO_USER  string
	MONGO_PASS  string
	DB_NAME     string
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
	fileName := fmt.Sprintf("./config/%s_config.json", env)
	gonfig.GetConf(fileName, &configuration)

	return configuration
}
