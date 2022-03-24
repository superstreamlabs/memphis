package routes

import (
	"strech-server/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeFactoriesRoutes(router *gin.RouterGroup) {
	factoriesHandler := handlers.FactoriesHandler{}
	factoriesRoutes := router.Group("/factories")
	factoriesRoutes.POST("/createFactory", factoriesHandler.CreateFactory)
	factoriesRoutes.GET("/getAllFactories", factoriesHandler.GetAllFactories)
	factoriesRoutes.DELETE("/removeFactory", factoriesHandler.RemoveFactory)
	factoriesRoutes.PUT("/editFactory", factoriesHandler.EditFactory)
}
