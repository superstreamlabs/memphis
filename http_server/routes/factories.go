package routes

import (
	"memphis-control-plane/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeFactoriesRoutes(router *gin.RouterGroup) {
	factoriesHandler := handlers.FactoriesHandler{}
	factoriesRoutes := router.Group("/factories")
	factoriesRoutes.POST("/createFactory", factoriesHandler.CreateFactory)
	factoriesRoutes.GET("/getAllFactories", factoriesHandler.GetAllFactories)
	factoriesRoutes.GET("/getFactory", factoriesHandler.GetFactory)
	factoriesRoutes.DELETE("/removeFactory", factoriesHandler.RemoveFactory)
	factoriesRoutes.PUT("/editFactory", factoriesHandler.EditFactory)
}
