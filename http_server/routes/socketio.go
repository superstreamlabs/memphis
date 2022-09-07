// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package routes

import (
	"memphis-broker/middlewares"
	"memphis-broker/models"
	"memphis-broker/server"

	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
)

var socketServer = socketio.NewServer(nil)

func getMainOverviewData(h *server.Handlers) (models.MainOverviewData, error) {
	stations, err := h.Stations.GetAllStationsDetails()
	if err != nil {
		return models.MainOverviewData{}, nil
	}
	totalMessages, err := h.Stations.GetTotalMessagesAcrossAllStations()
	if err != nil {
		return models.MainOverviewData{}, err
	}
	systemComponents, err := h.Monitoring.GetSystemComponents()
	if err != nil {
		return models.MainOverviewData{}, err
	}

	return models.MainOverviewData{
		TotalStations:    len(stations),
		TotalMessages:    totalMessages,
		SystemComponents: systemComponents,
		Stations:         stations,
	}, nil
}

func getFactoriesOverviewData(h *server.Handlers) ([]models.ExtendedFactory, error) {
	factories, err := h.Factories.GetAllFactoriesDetails()
	if err != nil {
		return factories, err
	}

	return factories, nil
}

func getFactoryOverviewData(factoryName string, s socketio.Conn, h *server.Handlers) (map[string]interface{}, error) {
	factoryName = strings.ToLower(factoryName)
	factory, err := h.Factories.GetFactoryDetails(factoryName)
	if err != nil {
		if s != nil && err.Error() == "mongo: no documents in result" {
			s.Emit("error", "Factory does not exist")
		}
		return factory, err
	}

	return factory, nil
}

func getStationOverviewData(stationName string, s socketio.Conn, h *server.Handlers) (models.StationOverviewData, error) {
	stationName = strings.ToLower(stationName)
	exist, station, err := server.IsStationExist(stationName)
	if err != nil {
		return models.StationOverviewData{}, err
	}
	if !exist {
		if s != nil {
			s.Emit("error", "Station does not exist")
		}
		return models.StationOverviewData{}, errors.New("Station does not exist")
	}

	connectedProducers, disconnectedProducers, deletedProducers, err := h.Producers.GetProducersByStation(station)
	if err != nil {
		return models.StationOverviewData{}, err
	}
	connectedCgs, disconnectedCgs, deletedCgs, err := h.Consumers.GetCgsByStation(station)
	if err != nil {
		return models.StationOverviewData{}, err
	}
	auditLogs, err := h.AuditLogs.GetAuditLogsByStation(station)
	if err != nil {
		return models.StationOverviewData{}, err
	}
	totalMessages, err := h.Stations.GetTotalMessages(station)
	if err != nil {
		return models.StationOverviewData{}, err
	}
	avgMsgSize, err := h.Stations.GetAvgMsgSize(station)
	if err != nil {
		return models.StationOverviewData{}, err
	}

	messagesToFetch := 1000
	messages, err := h.Stations.GetMessages(station, messagesToFetch)
	if err != nil {
		return models.StationOverviewData{}, err
	}

	poisonMessages, err := h.PoisonMsgs.GetPoisonMsgsByStation(station)
	if err != nil {
		return models.StationOverviewData{}, err
	}

	return models.StationOverviewData{
		ConnectedProducers:    connectedProducers,
		DisconnectedProducers: disconnectedProducers,
		DeletedProducers:      deletedProducers,
		ConnectedCgs:          connectedCgs,
		DisconnectedCgs:       disconnectedCgs,
		DeletedCgs:            deletedCgs,
		TotalMessages:         totalMessages,
		AvgMsgSize:            avgMsgSize,
		AuditLogs:             auditLogs,
		Messages:              messages,
		PoisonMessages:        poisonMessages,
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

func InitializeSocketio(router *gin.Engine, h *server.Handlers) *socketio.Server {
	serv := h.Stations.S
	socketServer.OnConnect("/api", func(s socketio.Conn) error {
		return nil
	})

	socketServer.OnEvent("/api", "register_main_overview_data", func(s socketio.Conn, msg string) string {
		s.LeaveAll()
		s.Join("main_overview_sockets_group")

		return "recv " + msg
	})

	socketServer.OnEvent("/api", "register_factories_overview_data", func(s socketio.Conn, msg string) string {
		s.LeaveAll()
		s.Join("factories_overview_sockets_group")

		return "recv " + msg
	})

	socketServer.OnEvent("/api", "register_factory_overview_data", func(s socketio.Conn, factoryName string) string {
		s.LeaveAll()
		s.Join("factory_overview_group_" + factoryName)

		return "recv " + factoryName
	})

	socketServer.OnEvent("/api", "register_station_overview_data", func(s socketio.Conn, stationName string) string {
		s.LeaveAll()
		s.Join("station_overview_group_" + stationName)

		return "recv " + stationName
	})

	socketServer.OnEvent("/api", "register_poison_message_journey_data", func(s socketio.Conn, poisonMsgId string) string {
		s.LeaveAll()
		s.Join("poison_message_journey_group_" + poisonMsgId)

		return "recv " + poisonMsgId
	})

	socketServer.OnEvent("/api", "deregister", func(s socketio.Conn, msg string) string {
		s.LeaveAll()
		return "recv " + msg
	})

	socketServer.OnEvent("/api", "get_all_stations_overview_data", func(s socketio.Conn, msg string) string {
		s.LeaveAll()
		s.Join("stations_overview_group_")
		return "recv " + msg
	})

	socketServer.OnError("/", func(s socketio.Conn, e error) {
		serv.Errorf("An error occured during a socket connection " + e.Error())
	})

	go socketServer.Serve()

	go func() {
		for range time.Tick(time.Second * 5) {
			if socketServer.RoomLen("/api", "main_overview_sockets_group") > 0 {
				data, err := getMainOverviewData(h)
				if err != nil {
					serv.Errorf("Error while trying to get main overview data - " + err.Error())
				} else {
					socketServer.BroadcastToRoom("/api", "main_overview_sockets_group", "main_overview_data", data)
				}
			}

			if socketServer.RoomLen("/api", "factories_overview_sockets_group") > 0 {
				data, err := getFactoriesOverviewData(h)
				if err != nil {
					serv.Errorf("Error while trying to get factories overview data - " + err.Error())
				} else {
					socketServer.BroadcastToRoom("/api", "factories_overview_sockets_group", "factories_overview_data", data)
				}
			}

			rooms := socketServer.Rooms("/api")
			for _, room := range rooms {
				if strings.HasPrefix(room, "station_overview_group_") && socketServer.RoomLen("/api", room) > 0 {
					stationName := strings.Split(room, "station_overview_group_")[1]
					data, err := getStationOverviewData(stationName, nil, h)
					if err != nil {
						serv.Errorf("Error while trying to get station overview data - " + err.Error())
					} else {
						socketServer.BroadcastToRoom("/api", room, "station_overview_data_"+stationName, data)
					}
				}

				if strings.HasPrefix(room, "factory_overview_group_") && socketServer.RoomLen("/api", room) > 0 {
					factoryName := strings.Split(room, "factory_overview_group_")[1]
					data, err := getFactoryOverviewData(factoryName, nil, h)
					if err != nil {
						serv.Errorf("Error while trying to get factory overview data - " + err.Error())
					} else {
						socketServer.BroadcastToRoom("/api", room, "factory_overview_data_"+factoryName, data)
					}
				}

				if strings.HasPrefix(room, "poison_message_journey_group_") && socketServer.RoomLen("/api", room) > 0 {
					poisonMsgId := strings.Split(room, "poison_message_journey_group_")[1]
					data, err := h.Stations.GetPoisonMessageJourneyDetails(poisonMsgId)
					if err != nil {
						serv.Errorf("Error while trying to get poison message journey - " + err.Error())
					} else {
						socketServer.BroadcastToRoom("/api", room, "poison_message_journey_data_"+poisonMsgId, data)
					}
				}

				if strings.HasPrefix(room, "stations_overview_group_") && socketServer.RoomLen("/api", room) > 0 {
					poisonMsgId := strings.Split(room, "stations_overview_group_")[1]
					data, err := h.Stations.GetAllStationsExtendedDetails()
					if err != nil {
						serv.Errorf("Error while trying to get all stations details - " + err.Error())
					} else {
						socketServer.BroadcastToRoom("/api", room, "stations_overview_group_"+poisonMsgId, data)
					}
				}
			}
		}
	}()

	socketIoRouter := router.Group("/api/socket.io")
	socketIoRouter.Use(ginMiddleware())
	socketIoRouter.Use(middlewares.Authenticate)

	socketIoRouter.GET("/*any", gin.WrapH(socketServer))
	socketIoRouter.POST("/*any", gin.WrapH(socketServer))
	return socketServer
}
