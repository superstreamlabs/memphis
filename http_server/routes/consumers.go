package routes

import (
	"memphis-control-plane/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeConsumersRoutes(router *gin.RouterGroup) {
	consumersHandler := handlers.ConsumersHandler{}
	consumersRoutes := router.Group("/consumers")
	consumersRoutes.GET("/getAllConsumers", consumersHandler.GetAllConsumers)
	consumersRoutes.GET("/getAllConsumersByStation", consumersHandler.GetAllConsumersByStation)
	consumersRoutes.POST("/createConsumer", consumersHandler.CreateConsumer)
}
