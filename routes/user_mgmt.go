package routes

import (
	"memphis-control-plane/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeUserMgmtRoutes(router *gin.RouterGroup) {
	userMgmtHandler := handlers.UserMgmtHandler{}
	userMgmtRoutes := router.Group("/usermgmt")
	userMgmtRoutes.GET("/nats/authenticate", userMgmtHandler.AuthenticateNatsUser)
	userMgmtRoutes.GET("/nats/authenticate/:publicKey", userMgmtHandler.AuthenticateNatsUser)
	userMgmtRoutes.POST("/login", userMgmtHandler.Login)
	userMgmtRoutes.POST("/refreshToken", userMgmtHandler.RefreshToken)
	userMgmtRoutes.POST("/logout", userMgmtHandler.Logout)
	userMgmtRoutes.POST("/addUser", userMgmtHandler.AddUser)
	userMgmtRoutes.GET("/getAllUsers", userMgmtHandler.GetAllUsers)
	userMgmtRoutes.DELETE("/removeUser", userMgmtHandler.RemoveUser)
	userMgmtRoutes.DELETE("/removeMyUser", userMgmtHandler.RemoveMyUser)
	userMgmtRoutes.PUT("/editAvatar", userMgmtHandler.EditAvatar)
	userMgmtRoutes.PUT("/editHubCreds", userMgmtHandler.EditHubCreds)
	userMgmtRoutes.PUT("/editCompanyLogo", userMgmtHandler.EditCompanyLogo)
	userMgmtRoutes.DELETE("/removeCompanyLogo", userMgmtHandler.RemoveCompanyLogo)
	userMgmtRoutes.GET("/getCompanyLogo", userMgmtHandler.GetCompanyLogo)
}
