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
	"memphis-control-plane/middlewares"
	"memphis-control-plane/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitializeHttpRoutes() *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://*", "https://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowWebSockets:  true,
		AllowFiles:       true,
	}))
	mainRouter := router.Group("/api")
	mainRouter.Use(middlewares.Authenticate)

	utils.InitializeValidations()
	InitializeUserMgmtRoutes(mainRouter)
	InitializeFactoriesRoutes(mainRouter)
	InitializeStationsRoutes(mainRouter)
	InitializeProducersRoutes(mainRouter)
	InitializeConsumersRoutes(mainRouter)
	mainRouter.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Ok",
		})
	})

	return router
}
