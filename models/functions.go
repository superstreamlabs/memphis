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

type FunctionsResult struct {
	FunctionName    string            `json:"function_name"`
	Description     string            `json:"description"`
	Tags            []string          `json:"tags"`
	RunTime         string            `json:"runtime"`
	Memory          int               `json:"memory"`
	Storgae         int               `json:"storgae"`
	LastCommit      time.Time         `json:"last_commit"`
	Link            string            `json:"link"`
	Repository      string            `json:"repository"`
	Branch          string            `json:"branch"`
	Owner           string            `json:"owner"`
	EnvironmentVars map[string]string `json:"environment_vars"`
	ScmType         string            `json:"scm_type"`
	Language        string            `json:"language"`
}

type FunctionsRes struct {
	Functions     []FunctionsResult `json:"functions"`
	ScmIntegrated bool              `json:"scm_integrated"`
}

type GetFunctionDetails struct {
	Repository string `form:"repo" json:"repo"`
	Branch     string `form:"branch" json:"branch"`
	Owner      string `form:"owner" json:"owner"`
	Scm        string `form:"scm" json:"scm"`
	Type       string `form:"type" json:"type"`
	Path       string `form:"path" json:"path"`
}

type InstallFunction struct {
	ID              int                      `json:"id"`
	FunctionName    string                   `json:"function_name"`
	Description     string                   `json:"description"`
	Tags            []string                 `json:"tags"`
	Runtime         string                   `json:"runtime"`
	Dependencies    string                   `json:"dependencies"`
	EnvironmentVars []map[string]interface{} `json:"environment_vars"`
	Memory          int                      `json:"memory"`
	Storage         int                      `json:"storage"`
	Handler         interface{}              `json:"handler"`
	TenantName      string                   `json:"tenant_name"`
	Scm             string                   `json:"scm"`
	Owner           string                   `json:"owner"`
	Repo            string                   `json:"repo"`
	Branch          string                   `json:"branch"`
	UpdatedAt       time.Time                `json:"updated_at"`
	Version         int                      `json:"version"`
	InProgress      bool                     `json:"in_progress"`
}
