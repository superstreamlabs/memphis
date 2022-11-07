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
	"memphis-broker/server"

	"github.com/gin-gonic/gin"
)

func InitializeSchemasRoutes(router *gin.RouterGroup, h *server.Handlers) {
	schemasHandler := h.Schemas
	schemasRoutes := router.Group("/schemas")
	schemasRoutes.POST("/createNewSchema", schemasHandler.CreateNewSchema)
	schemasRoutes.GET("/getAllSchemas", schemasHandler.GetAllSchemas)
	schemasRoutes.GET("/getSchemaDetails", schemasHandler.GetSchemaDetails)
	schemasRoutes.DELETE("/removeSchema", schemasHandler.RemoveSchema)
	schemasRoutes.POST("/createNewVersion", schemasHandler.CreateNewVersion)
	schemasRoutes.PUT("/rollBackVersion", schemasHandler.RollBackVersion)
	schemasRoutes.POST("/validateSchema", schemasHandler.ValidateSchema)
}
