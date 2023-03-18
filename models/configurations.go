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
package models

type EditClusterConfigSchema struct {
	PMRetention   int    `json:"pm_retention" binding:"required"`
	LogsRetention int    `json:"logs_retention" binding:"required"`
	BrokerHost    string `json:"broker_host"`
	UiHost        string `json:"ui_host"`
	RestGWHost    string `json:"rest_gw_host"`
	TSTimeSec     int    `json:"tiered_storage_time_sec" binding:"min=5,max=3600"`
}

type GlobalConfigurationsUpdate struct {
	Notifications bool `json:"notifications"`
}

type SdkClientsUpdates struct {
	StationName string `json:"station_name"`
	Type        string `json:"type"`
	Update      any    `json:"update"`
}

type ConfigurationsIntValue struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Value int    `json:"value"`
}

type ConfigurationsStringValue struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}
