package routes

import (
	"strech-server/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeStationsRoutes(router *gin.RouterGroup) {
	stationsHandler := handlers.StationsHandler{}
	stationsRoutes := router.Group("/stations")
	stationsRoutes.GET("/getStation", stationsHandler.GetStation)
	stationsRoutes.POST("/createStation", stationsHandler.CreateStation)
	stationsRoutes.DELETE("/removeStation", stationsHandler.RemoveStation)
}
