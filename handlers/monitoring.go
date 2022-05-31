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
	"context"
	"fmt"
	"memphis-control-plane/broker"
	"memphis-control-plane/db"
	"memphis-control-plane/models"
	"net/http"
)

type MonitoringHandler struct{}

func (mh MonitoringHandler) GetSystemComponents() ([]models.SystemComponent, error) {
	var components []models.SystemComponent
	if configuration.DOCKER_ENV != "" {
		resp, err := http.Get("http://localhost:9000")
		fmt.Print((resp))
		if err != nil {
			components = append(components, models.SystemComponent{
				Component:   "UI",
				DesiredPods: 1,
				ActualPods:  0,
			})
		} else {
			components = append(components, models.SystemComponent{
				Component:   "UI",
				DesiredPods: 1,
				ActualPods:  1,
			})
		}

		if broker.IsConnectionAlive() {
			components = append(components, models.SystemComponent{
				Component:   "broker",
				DesiredPods: 1,
				ActualPods:  1,
			})
		} else {
			components = append(components, models.SystemComponent{
				Component:   "broker",
				DesiredPods: 1,
				ActualPods:  0,
			})
		}

		err = db.Client.Ping(context.TODO(), nil)
		if err != nil {
			components = append(components, models.SystemComponent{
				Component:   "application-db",
				DesiredPods: 1,
				ActualPods:  0,
			})
		} else {
			components = append(components, models.SystemComponent{
				Component:   "application-db",
				DesiredPods: 1,
				ActualPods:  1,
			})
		}

		components = append(components, models.SystemComponent{
			Component:   "control-plane",
			DesiredPods: 1,
			ActualPods:  1,
		})
	} else {
		// k8s implementation

	}

	return components, nil
}
