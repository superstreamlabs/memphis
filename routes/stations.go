package routes

import (
	"strech-server/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeStationsRoutes(router *gin.RouterGroup) {
	stationsHandler := handlers.StationsHandler{}
	stationsRoutes := router.Group("/stations")
	stationsRoutes.GET("/getStationById", stationsHandler.GetStationById)
	stationsRoutes.GET("/getFactoryStations", stationsHandler.GetFactoryStations)
	stationsRoutes.POST("/createStation", stationsHandler.CreateStation)
	stationsRoutes.DELETE("/removeStation", stationsHandler.RemoveStation)
}
