package routes

import (
	"strech-server/handlers"
	"github.com/gin-gonic/gin"
)

func InitializeUserMgmtRoutes(router *gin.Engine) {
	userMgmtHandler := handlers.UserMgmtHandler{}

	// router.GET("/xxx", userMgmtHandler.GetAlbums)
	// router.GET("/xxx/:id", userMgmtHandler.GetAlbumByID)
	router.POST("/usermgmt/createUser", userMgmtHandler.CreateUser)
}
