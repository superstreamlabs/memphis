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

func InitializeStationsRoutes(router *gin.RouterGroup, h *server.Handlers) {
	stationsHandler := h.Stations
	stationsRoutes := router.Group("/stations")
	stationsRoutes.GET("/getStation", stationsHandler.GetStation)
	stationsRoutes.GET("/getMessageDetails", stationsHandler.GetMessageDetails)
	stationsRoutes.GET("/getAllStations", stationsHandler.GetAllStations)
	stationsRoutes.GET("/getStations", stationsHandler.GetStations)
	stationsRoutes.GET("/getPoisonMessageJourney", stationsHandler.GetPoisonMessageJourney)
	stationsRoutes.POST("/createStation", stationsHandler.CreateStation)
	stationsRoutes.POST("/resendPoisonMessages", stationsHandler.ResendPoisonMessages)
	stationsRoutes.POST("/ackPoisonMessages", stationsHandler.AckPoisonMessages)
	stationsRoutes.DELETE("/removeStation", stationsHandler.RemoveStation)
	stationsRoutes.POST("/useSchema", stationsHandler.UseSchema)
	stationsRoutes.DELETE("/removeSchemaFromStation", stationsHandler.RemoveSchemaFromStation)
	stationsRoutes.GET("/getUpdatesForSchemaByStation", stationsHandler.GetUpdatesForSchemaByStation)
	stationsRoutes.GET("/tierdStorageClicked", stationsHandler.TierdStorageClicked) // TODO to be deleted
	stationsRoutes.PUT("/dlsConfiguration", stationsHandler.DlsConfiguration)
}
