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

package socketio

import (
	"memphis-control-plane/handlers"
	"memphis-control-plane/logger"
	"memphis-control-plane/middlewares"
	"memphis-control-plane/models"

	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var server = socketio.NewServer(nil)

var producersHandler = handlers.ProducersHandler{}
var consumersHandler = handlers.ConsumersHandler{}
var auditLogsHandler = handlers.AuditLogsHandler{}
var stationsHandler = handlers.StationsHandler{}
var factoriesHandler = handlers.FactoriesHandler{}
var monitoringHandler = handlers.MonitoringHandler{}
var sysLogsHandler = handlers.SysLogsHandler{}

type mainOverviewData struct {
	TotalStations    int                      `json:"total_stations"`
	TotalMessages    int                      `json:"total_messages"`
	SystemComponents []models.SystemComponent `json:"system_components"`
	Stations         []models.ExtendedStation `json:"stations"`
}

type stationOverviewData struct {
	Producers     []models.ExtendedProducer `json:"producers"`
	Consumers     []models.ExtendedConsumer `json:"consumers"`
	TotalMessages int                       `json:"total_messages"`
	AvgMsgSize    int64                     `json:"average_message_size"`
	AuditLogs     []models.AuditLog         `json:"audit_logs"`
}

type factoryOverviewData struct {
	ID            primitive.ObjectID `json:"id"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	CreatedByUser string             `json:"created_by_user"`
	CreationDate  time.Time          `json:"creation_date"`
	Stations      []models.Station   `json:"stations"`
	UserAvatarId  int                `json:"user_avatar_id"`
}

func getMainOverviewData() (mainOverviewData, error) {
	stations, err := stationsHandler.GetAllStationsDetails()
	if err != nil {
		return mainOverviewData{}, nil
	}
	totalMessages, err := stationsHandler.GetTotalMessagesAcrossAllStations()
	if err != nil {
		return mainOverviewData{}, err
	}
	systemComponents, err := monitoringHandler.GetSystemComponents()
	if err != nil {
		logger.Error("GetSystemComponents error" + err.Error())
	}

	return mainOverviewData{
		TotalStations:    len(stations),
		TotalMessages:    totalMessages,
		SystemComponents: systemComponents,
		Stations:         stations,
	}, nil
}

func getFactoriesOverviewData() ([]models.ExtendedFactory, error) {
	factories, err := factoriesHandler.GetAllFactoriesDetails()
	if err != nil {
		return factories, err
	}

	return factories, nil
}

func getFactoryOverviewData(factoryName string, s socketio.Conn) (map[string]interface{}, error) {
	factoryName = strings.ToLower(factoryName)
	factory, err := factoriesHandler.GetFactoryDetails(factoryName)
	if err != nil {
		if s != nil && err.Error() == "mongo: no documents in result"  {
			s.Emit("error", "Factory does not exist")
		}
		return factory, err
	}

	return factory, nil
}

func getStationOverviewData(stationName string, s socketio.Conn) (stationOverviewData, error) {
	stationName = strings.ToLower(stationName)
	exist, station, err := handlers.IsStationExist(stationName)
	if err != nil {
		return stationOverviewData{}, err
	}
	if !exist {
		if s != nil {
			s.Emit("error", "Station does not exist")
		}
		return stationOverviewData{}, errors.New("Station does not exist")
	}

	producers, err := producersHandler.GetProducersByStation(station)
	if err != nil {
		return stationOverviewData{}, err
	}
	consumers, err := consumersHandler.GetConsumersByStation(station)
	if err != nil {
		return stationOverviewData{}, err
	}
	auditLogs, err := auditLogsHandler.GetAuditLogsByStation(station)
	if err != nil {
		return stationOverviewData{}, err
	}
	totalMessages, err := stationsHandler.GetTotalMessages(station)
	if err != nil {
		return stationOverviewData{}, err
	}
	avgMsgSize, err := stationsHandler.GetAvgMsgSize(station)
	if err != nil {
		return stationOverviewData{}, err
	}

	// get messages

	return stationOverviewData{
		Producers:     producers,
		Consumers:     consumers,
		TotalMessages: totalMessages,
		AvgMsgSize:    avgMsgSize,
		AuditLogs:     auditLogs,
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

func SendSysLogs(logs []models.SysLog) {
	if server.RoomLen("/api", "system_logs_group") > 0 {
		server.BroadcastToRoom("/api", "system_logs_group", "system_logs_data", logs)
	}
}

func InitializeSocketio(router *gin.Engine) *socketio.Server {

	server.OnConnect("/api", func(s socketio.Conn) error {
		return nil
	})

	server.OnEvent("/api", "register_main_overview_data", func(s socketio.Conn, msg string) string {
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

	server.OnEvent("/api", "register_factories_overview_data", func(s socketio.Conn, msg string) string {
		s.LeaveAll()
		data, err := getFactoriesOverviewData()
		if err != nil {
			logger.Error("Error while trying to get factories overview data " + err.Error())
		} else {
			s.Emit("factories_overview_data", data)
			s.Join("factories_overview_sockets_group")
		}

		return "recv " + msg
	})

	server.OnEvent("/api", "register_factory_overview_data", func(s socketio.Conn, factoryName string) string {
		s.LeaveAll()
		data, err := getFactoryOverviewData(factoryName, s)
		if err != nil {
			logger.Error("Error while trying to get factory overview data " + err.Error())
		} else {
			s.Emit("factory_overview_data", data)
			s.Join("factory_overview_group_" + factoryName)
		}

		return "recv " + factoryName
	})

	server.OnEvent("/api", "register_station_overview_data", func(s socketio.Conn, stationName string) string {
		s.LeaveAll()
		data, err := getStationOverviewData(stationName, s)
		if err != nil {
			logger.Error("Error while trying to get station overview data " + err.Error())
		} else {
			s.Emit("station_overview_data", data)
			s.Join("station_overview_group_" + stationName)
		}

		return "recv " + stationName
	})

	server.OnEvent("/api", "register_system_logs_data", func(s socketio.Conn, msg string) string {
		s.LeaveAll()
		hours := 24
		logs, err := sysLogsHandler.GetSysLogs(hours)
		if err != nil {
			logger.Error("Failed to fetch sys logs")
		} else {
			s.Emit("system_logs_data", logs)
			s.Join("system_logs_group")
		}

		return "recv " + msg
	})

	server.OnEvent("/api", "deregister", func(s socketio.Conn, msg string) string {
		s.LeaveAll()
		return "recv " + msg
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		logger.Error("An error occured during a socket connection " + e.Error())
	})

	go server.Serve()

	go func() {
		for range time.Tick(time.Second * 5) {
			if server.RoomLen("/api", "main_overview_sockets_group") > 0 {
				data, err := getMainOverviewData()
				if err != nil {
					logger.Error("Error while trying to get main overview data - " + err.Error())
				} else {
					server.BroadcastToRoom("/api", "main_overview_sockets_group", "main_overview_data", data)
				}
			}

			if server.RoomLen("/api", "factories_overview_sockets_group") > 0 {
				data, err := getFactoriesOverviewData()
				if err != nil {
					logger.Error("Error while trying to get factories overview data - " + err.Error())
				} else {
					server.BroadcastToRoom("/api", "factories_overview_sockets_group", "factories_overview_data", data)
				}
			}

			rooms := server.Rooms("/api")
			for _, room := range rooms {
				if strings.HasPrefix(room, "station_overview_group_") && server.RoomLen("/api", room) > 0 {
					stationName := strings.Split(room, "station_overview_group_")[1]
					data, err := getStationOverviewData(stationName, nil)
					if err != nil {
						logger.Error("Error while trying to get station overview data - " + err.Error())
					} else {
						server.BroadcastToRoom("/api", room, "station_overview_data", data)
					}
				}

				if strings.HasPrefix(room, "factory_overview_group_") && server.RoomLen("/api", room) > 0 {
					factoryName := strings.Split(room, "factory_overview_group_")[1]
					data, err := getFactoryOverviewData(factoryName, nil)
					if err != nil {
						logger.Error("Error while trying to get factory overview data - " + err.Error())
					} else {
						server.BroadcastToRoom("/api", room, "factory_overview_data", data)
					}
				}
			}
		}
	}()

	socketIoRouter := router.Group("/api/socket.io")
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:9000", "http://0.0.0.0", "https://0.0.0.0", "http://*", "https://*"},
	}))
	socketIoRouter.Use(ginMiddleware())
	socketIoRouter.Use(middlewares.Authenticate)

	socketIoRouter.GET("/*any", gin.WrapH(server))
	socketIoRouter.POST("/*any", gin.WrapH(server))
	return server
}
