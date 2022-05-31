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

package handlers

import (
	"io/ioutil"
	"memphis-control-plane/models"

	"github.com/gin-gonic/gin"

	"memphis-control-plane/logger"
)

type MonitoringHandler struct{}

func (mh MonitoringHandler) GetSystemComponents() ([]models.SystemComponent, error) {
	if configuration.DOCKER_ENV != "" {
		// localhost:9000 UI
		// localhost: 5555 control plane
		// localhost: 7766 broker
		// localhost:27017 mongo
	} else {

	}

	return []models.SystemComponent{}, nil
}

func (mh MonitoringHandler) GetClusterInfo(c *gin.Context) {
	body, err := ioutil.ReadFile("version.conf")
	if err != nil {
		logger.Error("GetClusterInfo error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, string(body))
}
