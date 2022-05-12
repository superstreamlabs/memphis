// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
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
	PING_INTERVAL_MS               int
	ANALYTICS                      bool
	ANALYTICS_TOKEN                string
}

func GetConfig() Configuration {
	configuration := Configuration{}
	if os.Getenv("DOCKER_ENV") != "" {
		gonfig.GetConf("./config/docker-config.json", &configuration)
	} else {
		gonfig.GetConf("./config/config.json", &configuration)
	}

	gin.SetMode(gin.ReleaseMode)
	return configuration
}
