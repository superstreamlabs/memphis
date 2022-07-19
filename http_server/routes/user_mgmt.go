// Copyright 2021-2022 The Memphis Authors
// Licensed under the GNU General Public License v3.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routes

import (
	"memphis-broker/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeUserMgmtRoutes(router *gin.RouterGroup) {
	userMgmtHandler := handlers.UserMgmtHandler{}
	userMgmtRoutes := router.Group("/usermgmt")
	userMgmtRoutes.GET("/nats/authenticate", userMgmtHandler.AuthenticateNatsUser)
	userMgmtRoutes.GET("/nats/authenticate/:publicKey", userMgmtHandler.AuthenticateNatsUser)
	userMgmtRoutes.POST("/login", userMgmtHandler.Login)
	userMgmtRoutes.POST("/doneNextSteps", userMgmtHandler.DoneNextSteps)
	userMgmtRoutes.POST("/refreshToken", userMgmtHandler.RefreshToken)
	userMgmtRoutes.POST("/addUser", userMgmtHandler.AddUser)
	userMgmtRoutes.GET("/getAllUsers", userMgmtHandler.GetAllUsers)
	userMgmtRoutes.DELETE("/removeUser", userMgmtHandler.RemoveUser)
	userMgmtRoutes.DELETE("/removeMyUser", userMgmtHandler.RemoveMyUser)
	userMgmtRoutes.PUT("/editAvatar", userMgmtHandler.EditAvatar)
	userMgmtRoutes.PUT("/editHubCreds", userMgmtHandler.EditHubCreds)
	userMgmtRoutes.PUT("/editCompanyLogo", userMgmtHandler.EditCompanyLogo)
	userMgmtRoutes.DELETE("/removeCompanyLogo", userMgmtHandler.RemoveCompanyLogo)
	userMgmtRoutes.GET("/getCompanyLogo", userMgmtHandler.GetCompanyLogo)
	userMgmtRoutes.PUT("/editAnalytics", userMgmtHandler.EditAnalytics)
}
