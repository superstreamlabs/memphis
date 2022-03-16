package routes

import (
	"strech-server/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeApplicationsRoutes(router *gin.Engine) {
	applicationsHandler := handlers.ApplicationsHandler{}
	userMgmtRoutes := router.Group("/application")
	userMgmtRoutes.POST("/createApplication", applicationsHandler.CreateApplication)
	userMgmtRoutes.GET("/getAllApplications", applicationsHandler.GetAllApplications)
	userMgmtRoutes.DELETE("/removeApplication", applicationsHandler.RemoveApplication)
	userMgmtRoutes.PUT("/editApplication", applicationsHandler.EditApplication)
}
