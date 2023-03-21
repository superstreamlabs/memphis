// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package routes

import (
	"memphis/middlewares"
	"memphis/server"
	ui "memphis/ui_static_files"
	"memphis/utils"

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
	mainRouter.Use(func(c *gin.Context) {
		c.Set("server", handlers.Consumers.S)
		c.Next()
	})
	mainRouter.Use(middlewares.Authenticate)

	utils.InitializeValidations()
	InitializeUserMgmtRoutes(mainRouter)
	InitializeStationsRoutes(mainRouter, handlers)
	InitializeProducersRoutes(mainRouter, handlers)
	InitializeConsumersRoutes(mainRouter, handlers)
	InitializeMonitoringRoutes(mainRouter, handlers)
	InitializeTagsRoutes(mainRouter, handlers)
	InitializeSchemasRoutes(mainRouter, handlers)
	// InitializeSandboxRoutes(mainRouter)
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
