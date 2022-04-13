package routes

import (
	"memphis-control-plane/middlewares"
	"memphis-control-plane/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitializeHttpRoutes() *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://*", "https://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowWebSockets:  true,
		AllowFiles:       true,
	}))
	mainRouter := router.Group("/api")
	mainRouter.Use(middlewares.Authenticate)

	utils.InitializeValidations()
	InitializeUserMgmtRoutes(mainRouter)
	InitializeFactoriesRoutes(mainRouter)
	InitializeStationsRoutes(mainRouter)
	InitializeProducersRoutes(mainRouter)
	InitializeConsumersRoutes(mainRouter)
	mainRouter.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Ok",
		})
	})

	return router
}
