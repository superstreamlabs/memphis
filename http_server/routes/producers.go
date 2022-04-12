package routes

import (
	"memphis-control-plane/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeProducersRoutes(router *gin.RouterGroup) {
	producersHandler := handlers.ProducersHandler{}
	producersRoutes := router.Group("/producers")
	producersRoutes.GET("/getAllProducers", producersHandler.GetAllProducers)
	producersRoutes.GET("/getAllProducersByStation", producersHandler.GetAllProducersByStation)
	producersRoutes.POST("/createProducer", producersHandler.CreateProducer)
	producersRoutes.DELETE("/destroyProducer", producersHandler.DestroyProducer)
}
