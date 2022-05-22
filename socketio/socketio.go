// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package socketio

import (
	"memphis-control-plane/logger"
	"memphis-control-plane/middlewares"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
)

type sysyemComponent struct {
	PodName     string `json:"pod_name"`
	DesiredPods int    `json:"desired_pods"`
	ActualPods  int    `json:"actual_pods"`
}

type stations struct {
	StationName string `json:"station_name"`
	FactoryName string `json:"factory_name"`
}

type mainOverviewData struct {
	TotalStations    int               `json:"total_stations"`
	TotalMessages    int               `json:"total_messages"`
	SystemComponents []sysyemComponent `json:"system_components"`
	Stations         []stations        `json:"stations"`
}

type stationOverviewData struct {
	Field1 string `json:"field1"`
}

func getMainOverviewData() (mainOverviewData, error) {
	// getTotalMessages
	// getTotalStations
	// getStationsInfo
	systemComponents := []sysyemComponent{
		{PodName: "MongoDB", DesiredPods: 2, ActualPods: 2},
		{PodName: "Memphis Broker", DesiredPods: 9, ActualPods: 3},
		{PodName: "Memphis UI", DesiredPods: 2, ActualPods: 1},
	}

	stations := []stations{
		{StationName: "station_1", FactoryName: "factory_1"},
		{StationName: "station_2", FactoryName: "factory_2"},
		{StationName: "station_3", FactoryName: "factory_3"},
	}

	return mainOverviewData{
		TotalStations:    13,
		TotalMessages:    12000,
		SystemComponents: systemComponents,
		Stations:         stations,
	}, nil
}

func getStationOverviewData(stationName string) stationOverviewData {
	return stationOverviewData{
		Field1: "station_overview data " + stationName,
	}
}

func InitializeSocketio(router *gin.Engine) *socketio.Server {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		return nil
	})

	server.OnEvent("/", "register_main_overview_data", func(s socketio.Conn, msg string) {
		data, err := getMainOverviewData()
		if err != nil {
			logger.Error("Error while trying to get main overview data " + err.Error())
		} else {
			s.Emit("main_overview_data", data)
			s.Join("main_overview_sockets_group")
		}
	})

	server.OnEvent("/", "deregister_main_overview_data", func(s socketio.Conn, msg string) {
		s.Leave("main_overview_sockets_group")
	})

	server.OnEvent("/", "register_station_overview_data", func(s socketio.Conn, stationName string) {
		stationName = strings.ToLower(stationName)
		data := getStationOverviewData(stationName)
		s.Emit("station_overview_data_"+stationName, data)
		s.Join("station_sockets_group_" + stationName)
	})

	server.OnEvent("/", "deregister_station_overview_data", func(s socketio.Conn, stationName string) {
		stationName = strings.ToLower(stationName)
		s.Leave("station_sockets_group_" + stationName)
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		logger.Error("An error occured during a socket connection" + e.Error())
	})

	go func() {
		if err := server.Serve(); err != nil {
			logger.Error("socketio listen error " + err.Error())
		}
	}()

	go func() {
		for range time.Tick(time.Second * 10) {
			data, err := getMainOverviewData()
			if err != nil {
				logger.Error("Error while trying to get main overview data " + err.Error())
			} else {
				server.BroadcastToRoom("", "main_overview_sockets_group", "main_overview_data", data)
			}
		}
	}()

	socketIoRouter := router.Group("/api/socket.io")
	socketIoRouter.Use(middlewares.Authenticate)
	socketIoRouter.GET("/*any", gin.WrapH(server))
	return server
}
