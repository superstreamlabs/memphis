package routes

import (
	"strech-server/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeApplicationsRoutes(router *gin.RouterGroup) {
	applicationsHandler := handlers.ApplicationsHandler{}
	userMgmtRoutes := router.Group("/applications")
	userMgmtRoutes.POST("/createApplication", applicationsHandler.CreateApplication)
	userMgmtRoutes.GET("/getAllApplications", applicationsHandler.GetAllApplications)
	userMgmtRoutes.DELETE("/removeApplication", applicationsHandler.RemoveApplication)
	userMgmtRoutes.PUT("/editApplication", applicationsHandler.EditApplication)
}
