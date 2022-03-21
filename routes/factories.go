package routes

import (
	"strech-server/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeFactoriesRoutes(router *gin.RouterGroup) {
	factoriesHandler := handlers.FactoriesHandler{}
	factoriesRoutes := router.Group("/factories")
	factoriesRoutes.GET("/getFactoryById", factoriesHandler.GetFactoryById)
	factoriesRoutes.GET("/getApplicationFactories", factoriesHandler.GetApplicationFactories)
	factoriesRoutes.POST("/createFactory", factoriesHandler.CreateFactory)
	factoriesRoutes.DELETE("/removeFactory", factoriesHandler.RemoveFactory)
}
