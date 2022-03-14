package routes

import (
	"strech-server/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeBoxesRoutes(router *gin.Engine) {
	boxesHandler := handlers.BoxesHandler{}
	userMgmtRoutes := router.Group("/boxes")
	userMgmtRoutes.POST("/createBox", boxesHandler.CreateBox)
	userMgmtRoutes.GET("/getAllBoxes", boxesHandler.GetAllBoxes)
	userMgmtRoutes.DELETE("/removeBox", boxesHandler.RemoveBox)
	userMgmtRoutes.PUT("/editBox", boxesHandler.EditBox)
}
