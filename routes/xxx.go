package routes

import (
	"github.com/gin-gonic/gin"
	"strech-server/handlers"
)

func InitializeXXXRoutes(router *gin.Engine) {
	xxxHandler := handlers.XxxHandler{}

	router.GET("/xxx", xxxHandler.GetAlbums)
	router.GET("/xxx/:id", xxxHandler.GetAlbumByID)
	router.POST("/xxx", xxxHandler.PostAlbums)
}