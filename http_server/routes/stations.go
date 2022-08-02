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
	"memphis-broker/handlers"
	"memphis-broker/server"

	"github.com/gin-gonic/gin"
)

func InitializeStationsRoutes(s *server.Server, router *gin.RouterGroup) {
	stationsHandler := handlers.StationsHandler{S: s}
	stationsRoutes := router.Group("/stations")
	stationsRoutes.GET("/getStation", stationsHandler.GetStation)
	stationsRoutes.GET("/getMessageDetails", stationsHandler.GetMessageDetails)
	stationsRoutes.GET("/getAllStations", stationsHandler.GetAllStations)
	stationsRoutes.GET("/getPoisonMessageJourney", stationsHandler.GetPoisonMessageJourney)
	stationsRoutes.POST("/createStation", stationsHandler.CreateStation)
	stationsRoutes.POST("/resendPoisonMessages", stationsHandler.ResendPoisonMessages)
	stationsRoutes.POST("/ackPoisonMessages", stationsHandler.AckPoisonMessages)
	stationsRoutes.DELETE("/removeStation", stationsHandler.RemoveStation)
}
