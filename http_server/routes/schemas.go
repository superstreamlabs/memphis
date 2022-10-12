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
}