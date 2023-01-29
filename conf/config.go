// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package conf

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tkanos/gonfig"
)

type Configuration struct {
	MEMPHIS_VERSION                string
	DEV_ENV                        string
	HTTP_PORT                      string
	WS_PORT                        int
	WS_TLS                         bool
	WS_TOKEN                       string
	MONGO_URL                      string
	MONGO_USER                     string
	MONGO_PASS                     string
	DB_NAME                        string
	JWT_SECRET                     string
	JWT_EXPIRES_IN_MINUTES         int
	REFRESH_JWT_SECRET             string
	REFRESH_JWT_EXPIRES_IN_MINUTES int
	ROOT_PASSWORD                  string
	CONNECTION_TOKEN               string
	MAX_MESSAGE_SIZE_MB            int
	SHOWABLE_ERROR_STATUS_CODE     int
	DOCKER_ENV                     string
	ANALYTICS                      string
	ANALYTICS_TOKEN                string
	K8S_NAMESPACE                  string
	LOGS_RETENTION_IN_DAYS         string
	GOOGLE_CLIENT_ID               string
	GOOGLE_CLIENT_SECRET           string
	SANDBOX_ENV                    string
	GITHUB_CLIENT_ID               string
	GITHUB_CLIENT_SECRET           string
	SANDBOX_REDIRECT_URI           string
	POISON_MSGS_RETENTION_IN_HOURS int
	MAILCHIMP_KEY                  string
	MAILCHIMP_LIST_ID              string
	SERVER_NAME                    string
	SANDBOX_SLACK_BOT_TOKEN        string
	SANDBOX_SLACK_CHANNEL_ID       string
	SANDBOX_UI_URL                 string
	EXTERNAL_MONGO                 bool
	EXPORTER                       bool
}

func GetConfig() Configuration {
	configuration := Configuration{}
	if os.Getenv("DOCKER_ENV") != "" {
		gonfig.GetConf("./conf/docker-config.json", &configuration)
	} else {
		gonfig.GetConf("./conf/config.json", &configuration)
	}

	if configuration.EXTERNAL_MONGO {
		configuration.DB_NAME = "memphis-db"
	}

	gin.SetMode(gin.ReleaseMode)
	return configuration
}
