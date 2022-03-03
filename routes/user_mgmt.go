package routes

import (
	"strech-server/handlers"
	"github.com/gin-gonic/gin"
)

func InitializeUserMgmtRoutes(router *gin.Engine) {
	userMgmtHandler := handlers.UserMgmtHandler{}

	// router.GET("/xxx", userMgmtHandler.GetAlbums)
	// router.GET("/xxx/:id", userMgmtHandler.GetAlbumByID)
	router.POST("/usermgmt/jwt/v1/accounts", userMgmtHandler.AuthenticateNats)
	router.POST("/usermgmt/addUser", userMgmtHandler.AddUser)
	router.POST("/usermgmt/createRootUser", userMgmtHandler.CreateRootUser)
}
