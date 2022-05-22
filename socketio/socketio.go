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

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
)

type myResponse struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

func InitializeSocketio(router *gin.Engine) *socketio.Server {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		// authenticate
		a := myResponse{Field1: "test", Field2: 5}
		s.Join("connected_clients")
		server.BroadcastToRoom("", "connected_clients", "data", a)
		return nil
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		logger.Error("An error occured during a socket connection" + e.Error())
	})

	go func() {
		if err := server.Serve(); err != nil {
			logger.Error("socketio listen error " + err.Error())
		}
	}()

	socketIoRouter := router.Group("/api/socket.io")
	socketIoRouter.Use(middlewares.Authenticate)
	socketIoRouter.GET("/*any", gin.WrapH(server))
	return server
}
