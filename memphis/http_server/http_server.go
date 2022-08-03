// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http_server

import (
	"memphis-broker/config"
	"memphis-broker/http_server/routes"
	"memphis-broker/server"
	"memphis-broker/socketio"
	"sync"
)

func InitializeHttpServer(wg *sync.WaitGroup, s *server.Server) {
	configuration := config.GetConfig()

	httpServer := routes.InitializeHttpRoutes()
	socketioServer := socketio.InitializeSocketio(httpServer, s)

	defer socketioServer.Close()
	defer wg.Done()

	httpServer.Run("0.0.0.0:" + configuration.HTTP_PORT)
}
