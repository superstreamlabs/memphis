package routes

import (
	"strech-server/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeUserMgmtRoutes(router *gin.Engine) {
	userMgmtHandler := handlers.UserMgmtHandler{}
	userMgmtRoutes := router.Group("/usermgmt")
	userMgmtRoutes.GET("/jwt/v1/accounts", userMgmtHandler.AuthenticateNatsUser)
	userMgmtRoutes.GET("/jwt/v1/accounts/:publicKey", userMgmtHandler.AuthenticateNatsUser)
	userMgmtRoutes.POST("/createRootUser", userMgmtHandler.CreateRootUser)
	userMgmtRoutes.POST("/login", userMgmtHandler.Login)
	userMgmtRoutes.POST("/refreshToken", userMgmtHandler.RefreshToken)
	userMgmtRoutes.POST("/logout", userMgmtHandler.Logout)
	userMgmtRoutes.POST("/addUser", userMgmtHandler.AddUser)
	userMgmtRoutes.GET("/getAllUsers", userMgmtHandler.GetAllUsers)
	userMgmtRoutes.DELETE("/removeUser", userMgmtHandler.RemoveUser)
	userMgmtRoutes.DELETE("/removeMyUser", userMgmtHandler.RemoveMyUser)
	userMgmtRoutes.PUT("/editHubCreds", userMgmtHandler.EditHubCreds)
	userMgmtRoutes.PUT("/editCompanyLogo", userMgmtHandler.EditCompanyLogo)
	userMgmtRoutes.DELETE("/removeCompanyLogo", userMgmtHandler.RemoveCompanyLogo)
	userMgmtRoutes.GET("/getCompanyLogo", userMgmtHandler.GetCompanyLogo)
}
