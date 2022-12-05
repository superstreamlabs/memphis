// Credit for The NATS.IO Authors
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
// limitations under the License.package routes
package routes

import (
	"memphis-broker/middlewares"
	"memphis-broker/server"
	ui "memphis-broker/ui_static_files"
	"memphis-broker/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitializeHttpRoutes(handlers *server.Handlers) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
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
	InitializeStationsRoutes(mainRouter, handlers)
	InitializeProducersRoutes(mainRouter, handlers)
	InitializeConsumersRoutes(mainRouter, handlers)
	InitializeMonitoringRoutes(mainRouter, handlers)
	InitializeTagsRoutes(mainRouter, handlers)
	InitializeSchemasRoutes(mainRouter, handlers)
	InitializeSandboxRoutes(mainRouter)
	InitializeIntegrationsRoutes(mainRouter, handlers)
	InitializeConfigurationsRoutes(mainRouter, handlers)
	ui.InitializeUIRoutes(router)

	mainRouter.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Ok",
		})
	})

	return router
}
