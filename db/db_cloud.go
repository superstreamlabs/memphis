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
package db

const (
	testEventsTable                 = ``
	functionsTable                  = ``
	attachedFunctionsTable          = ``
	functionsEngineWorkersTable     = ``
	scheduledFunctionWorkersTable   = ``
	connectorsEngineWorkersTable    = ``
	connectorsConnectionsTable      = ``
	connectorsTable                 = ``
	alterConnectorsTable            = ``
	alterConnectorsConnectionsTable = ``
)

type FunctionSchema struct {
	ID               int    `json:"id"`
	TenantName       string `json:"tenant_name"`
	StationId        int    `json:"station_id"`
	PartitionNumber  int    `json:"partition_number"`
	Name             string `json:"function_name"`
	Version          int    `json:"function_version"`
	NextActiveStepId int    `json:"next_active_step_id"`
	PrevActiveStepId int    `json:"prev_active_step_id"`
	VisibleStep      int    `json:"visible_step"`
	AddedBy          string `json:"added_by"`
	Repo             string `json:"repo"`
	Branch           string `json:"branch"`
	Owner            string `json:"owner"`
	Runtime          string `json:"runtime"`
	OrderingMatter   bool   `json:"ordering_matter"`
	Activated        bool   `json:"activated"`
	ComputeEngine    string `json:"compute_engine"`
	SCM              string `json:"scm"`
	InstalledId      int    `json:"installed_id"`
}

func DeleteAndGetAttachedFunctionsByStation(tenantName string, stationId int, partitions []int) ([]FunctionSchema, error) {
	return []FunctionSchema{}, nil
}

func DeleteAndGetAttachedFunctionsByTenant(tenantName string) ([]FunctionSchema, error) {
	return []FunctionSchema{}, nil
}

func DeleteAllTestEvents(tenantName string) error {
	return nil
}

func DeleteScheduledFunctionWorkersByTenant(tenantName string) error {
	return nil
}

func DeleteScheduledFunctionWorkersByStationId(stationId int, tenantName string) error {
	return nil
}
