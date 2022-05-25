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
	"memphis-control-plane/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
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
	Producers     []models.Producer `json:"producers"`
	Consumers     []models.Consumer `json:"consumers"`
	TotalMessages int               `json:"total_messages"`
	AvgMsgSize    int               `json:"average_message_size"`
	// AuditLogs     []stations        `json:"audit_logs"`
}

func getMainOverviewData() (mainOverviewData, error) {
	// getTotalMessages - 
	// getTotalStations - 
	// getStationsInfo - 
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

func getStationOverviewData(stationName string) (stationOverviewData, error) {
	// get producers -
	// get consumers -
	// get total messages -
	// get avg msg size -
	// get audit logs
	// get messages

	var producers []models.Producer
	var consumers []models.Consumer
	// var auditLogs []models.AuditLog
	return stationOverviewData{
		Producers:     producers,
		Consumers:     consumers,
		TotalMessages: 400,
		AvgMsgSize:    100,
		// AuditLogs: auditLogs,
	}, nil
}

func ginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Request.Header.Del("Origin")
		c.Next()
	}
}

func InitializeSocketio(router *gin.Engine) *socketio.Server {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		return nil
	})

	server.OnEvent("/", "register_main_overview_data", func(s socketio.Conn, msg string) string {
		s.LeaveAll()
		data, err := getMainOverviewData()
		if err != nil {
			logger.Error("Error while trying to get main overview data " + err.Error())
		} else {
			s.Emit("main_overview_data", data)
			s.Join("main_overview_sockets_group")
		}

		return "recv " + msg
	})

	server.OnEvent("/", "deregister", func(s socketio.Conn, msg string) string {
		s.LeaveAll()
		return "recv " + msg
	})

	server.OnEvent("/", "register_station_overview_data", func(s socketio.Conn, stationName string) string {
		stationName = strings.ToLower(stationName)
		s.LeaveAll()
		data, err := getStationOverviewData(stationName)
		if err != nil {
			logger.Error("Error while trying to get station overview data " + err.Error())
		} else {
			s.Emit("station_overview_data", data)
			s.Join("station_overview_group_" + stationName)
		}

		return "recv " + stationName
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		logger.Error("An error occured during a socket connection " + e.Error())
	})

	go server.Serve()

	go func() {
		for range time.Tick(time.Second * 10) {
			if server.RoomLen("/", "main_overview_sockets_group") > 0 {
				data, err := getMainOverviewData()
				if err != nil {
					logger.Error("Error while trying to get main overview data " + err.Error())
				} else {
					server.BroadcastToRoom("/", "main_overview_sockets_group", "main_overview_data", data)
				}
			}

			rooms := server.Rooms("/")
			for _, room := range rooms {
				if strings.HasPrefix(room, "station_overview_group_") && server.RoomLen("", room) > 0 {
					stationName := strings.Split(room, "station_overview_group_")[1]
					data, err := getStationOverviewData(stationName)
					if err != nil {
						logger.Error("Error while trying to get station overview data " + err.Error())
					} else {
						server.BroadcastToRoom("/", room, "station_overview_data", data)
					}
				}
			}
		}
	}()

	socketIoRouter := router.Group("/api/socket.io")
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000", "http://*", "https://*"},
	}))
	socketIoRouter.Use(ginMiddleware())
	socketIoRouter.Use(middlewares.Authenticate)

	socketIoRouter.GET("/*any", gin.WrapH(server))
	socketIoRouter.POST("/*any", gin.WrapH(server))
	return server
}
