// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package conf

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tkanos/gonfig"
)

type Configuration struct {
	DEV_ENV                        string
	HTTP_PORT                      string
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
	BROKER_NAME                    string
}

func GetConfig() Configuration {
	configuration := Configuration{}
	if os.Getenv("DOCKER_ENV") != "" {
		gonfig.GetConf("./conf/docker-config.json", &configuration)
	} else {
		gonfig.GetConf("./conf/config.json", &configuration)
	}

	gin.SetMode(gin.ReleaseMode)
	return configuration
}
