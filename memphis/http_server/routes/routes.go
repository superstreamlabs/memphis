// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routes

import (
	"memphis-broker/logger"
	"memphis-broker/middlewares"
	"memphis-broker/utils"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type loggerWriter struct {
}

func (lw loggerWriter) Write(p []byte) (int, error) {
	log := string(p)
	splitted := strings.Split(log, "| ")
	statusCode := strings.Trim(splitted[1], " ")
	if statusCode != "200" && statusCode != "204" {
		logger.Error(log)
	}
	return len(p), nil
}

func InitializeHttpRoutes() *gin.Engine {
	gin.DefaultWriter = loggerWriter{}
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:9000", "https://sandbox.memphis.dev", "http://*", "https://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowWebSockets:  true,
		AllowFiles:       true,
	}))
	mainRouter := router.Group("/api")
	mainRouter.Use(middlewares.Authenticate)

	utils.InitializeValidations()
	InitializeUserMgmtRoutes(mainRouter)
	InitializeFactoriesRoutes(mainRouter)
	InitializeStationsRoutes(mainRouter)
	InitializeProducersRoutes(mainRouter)
	InitializeConsumersRoutes(mainRouter)
	InitializeMonitoringRoutes(mainRouter)
	InitializeSandboxRoutes(mainRouter)
	mainRouter.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Ok",
		})
	})

	return router
}
