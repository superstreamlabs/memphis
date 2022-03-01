package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strech-server/utils"
	// "strech-server/middlewares"
)

func InitializeHttpRoutes() *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowWebSockets:  true,
		AllowFiles:       true,
	}))
	// router.Use(middlewares.Json)

	utils.InitializeValidations()
	InitializeUserMgmtRoutes(router)
	router.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Ok",
		})
	})

	return router
}
