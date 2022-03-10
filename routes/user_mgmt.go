package routes

import (
	"strech-server/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeUserMgmtRoutes(router *gin.Engine) {
	userMgmtHandler := handlers.UserMgmtHandler{}

	router.POST("/usermgmt/createRootUser", userMgmtHandler.CreateRootUser)
	router.POST("/usermgmt/login", userMgmtHandler.Login)
	router.POST("/usermgmt/refreshToken", userMgmtHandler.RefreshToken)
	router.POST("/usermgmt/logout", userMgmtHandler.Logout)
	router.POST("/usermgmt/jwt/v1/accounts", userMgmtHandler.AuthenticateNats)
	router.POST("/usermgmt/addUser", userMgmtHandler.AddUser)
	router.GET("/usermgmt/getAllUsers", userMgmtHandler.GetAllUsers)
	router.DELETE("/usermgmt/removeUser", userMgmtHandler.RemoveUser)
	router.DELETE("/usermgmt/removeMyUser", userMgmtHandler.RemoveMyUser)
	router.PUT("/usermgmt/editHubCreds", userMgmtHandler.EditHubCreds)
	// router.POST("/usermgmt/uploadCompanyLogo", userMgmtHandler.UploadCompanyLogo)

}
