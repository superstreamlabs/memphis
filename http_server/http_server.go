// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package http_server

import (
	"fmt"

	"github.com/memphisdev/memphis/http_server/routes"
	"github.com/memphisdev/memphis/server"
)

func InitializeHttpServer(s *server.Server) {
	handlers := server.Handlers{
		Producers:      server.ProducersHandler{S: s},
		Consumers:      server.ConsumersHandler{S: s},
		AuditLogs:      server.AuditLogsHandler{},
		Stations:       server.StationsHandler{S: s},
		Monitoring:     server.MonitoringHandler{S: s},
		PoisonMsgs:     server.PoisonMessagesHandler{S: s},
		Schemas:        server.SchemasHandler{S: s},
		Configurations: server.ConfigurationsHandler{S: s},
		Integrations:   server.IntegrationsHandler{S: s},
		Tenants:        server.TenantHandler{S: s},
		Billing:        server.BillingHandler{S: s},
	}

	httpServer := routes.InitializeHttpRoutes(&handlers)
	httpServer.Run(fmt.Sprintf("0.0.0.0:%v", s.Opts().UiPort))
}
