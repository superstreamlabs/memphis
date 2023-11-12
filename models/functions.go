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

import "time"

type Function struct {
	ID               int                      `json:"id"`
	FunctionName     string                   `json:"function_name"`
	Description      string                   `json:"description"`
	Tags             []string                 `json:"tags"`
	Runtime          string                   `json:"runtime"`
	Dependencies     string                   `json:"dependencies"`
	EnvironmentVars  []map[string]interface{} `json:"environment_vars"`
	Memory           int                      `json:"memory"`
	Storage          int                      `json:"storage"`
	Handler          string                   `json:"handler"`
	TenantName       string                   `json:"tenant_name"`
	Scm              string                   `json:"scm"`
	Owner            string                   `json:"owner"`
	Repo             string                   `json:"repo"`
	Branch           string                   `json:"branch"`
	UpdatedAt        time.Time                `json:"installed_updated_at"`
	Version          int                      `json:"installed_version"`
	InProgress       bool                     `json:"installed_in_progress"`
	ComputeEngine    string                   `json:"compute_engine"`
	Installed        bool                     `json:"installed"`
	IsValid          bool                     `json:"is_valid"`
	InvalidReason    string                   `json:"invalid_reason"`
	UpdatesAvailable bool                     `json:"updates_available"`
	ByMemphis        bool                     `json:"by_memphis"`
}

type FunctionResult struct {
	ID                         int                      `json:"id"`
	FunctionName               string                   `json:"function_name"`
	Description                string                   `json:"description"`
	Tags                       []string                 `json:"tags"`
	Runtime                    string                   `json:"runtime"`
	Dependencies               string                   `json:"dependencies"`
	EnvironmentVars            []map[string]interface{} `json:"environment_vars"`
	Memory                     int                      `json:"memory"`
	Storage                    int                      `json:"storage"`
	Handler                    string                   `json:"handler"`
	TenantName                 string                   `json:"tenant_name"`
	Scm                        string                   `json:"scm"`
	Owner                      string                   `json:"owner"`
	Repo                       string                   `json:"repo"`
	Branch                     string                   `json:"branch"`
	UpdatedAt                  time.Time                `json:"installed_updated_at"`
	Version                    int                      `json:"installed_version"`
	InProgress                 bool                     `json:"installed_in_progress"`
	ComputeEngine              string                   `json:"compute_engine"`
	Installed                  bool                     `json:"installed"`
	IsValid                    bool                     `json:"is_valid"`
	InvalidReason              string                   `json:"invalid_reason"`
	UpdatesAvailable           bool                     `json:"updates_available"`
	ByMemphis                  bool                     `json:"by_memphis"`
	Language                   string                   `json:"language"`
	Link                       *string                  `json:"link,omitempty"`
	LastCommit                 *time.Time               `json:"last_commit,omitempty"`
	ClonedUpdatesAvailable     bool                     `json:"cloned_updates_available,omitempty"`
	ClonedUpdatesIsValid       bool                     `json:"cloned_updates_is_valid,omitempty"`
	ClonedUpdatesInvalidReason string                   `json:"cloned_updates_invalid_reason,omitempty"`
}
type FunctionsRes struct {
	InstalledFunctions []FunctionResult         `json:"installed_functions"`
	OtherFunctions     []FunctionResult         `json:"other_functions"`
	ScmIntegrated      bool                     `json:"scm_integrated"`
	ConnectedRepos     []map[string]interface{} `json:"connected_repos"`
}
