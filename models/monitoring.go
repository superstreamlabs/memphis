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

package models

type SystemComponent struct {
	Component   string `json:"component"`
	DesiredPods int    `json:"desired_pods"`
	ActualPods  int    `json:"actual_pods"`
}

type MainOverviewData struct {
	TotalStations    int               `json:"total_stations"`
	TotalMessages    int               `json:"total_messages"`
	SystemComponents []SystemComponent `json:"system_components"`
	Stations         []ExtendedStation `json:"stations"`
}

type StationOverviewData struct {
	ConnectedProducers    []ExtendedProducer         `json:"connected_producers"`
	DisconnectedProducers []ExtendedProducer         `json:"disconnected_producers"`
	DeletedProducers      []ExtendedProducer         `json:"destroyed_producers"`
	ConnectedCgs          []Cg                       `json:"connected_cgs"`
	DisconnectedCgs       []Cg                       `json:"disconnected_cgs"`
	DeletedCgs            []Cg                       `json:"deleted_cgs"`
	TotalMessages         int                        `json:"total_messages"`
	AvgMsgSize            int64                      `json:"average_message_size"`
	AuditLogs             []AuditLog                 `json:"audit_logs"`
	Messages              []Message                  `json:"messages"`
	PoisonMessages        []PoisonMessage `json:"poison_messages"`
}

type GetStationOverviewDataSchema struct {
	StationName string `form:"station_name" json:"station_name"  binding:"required"`
}
