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
package server

import (
	"errors"
	"fmt"
	"memphis/models"
)

type GetSourceCodeBranchesSchema struct {
	RepoName  string `form:"repo_name" json:"repo_name" binding:"required"`
	RepoOwner string `form:"repo_owner" json:"repo_owner" binding:"required"`
}

func getSourceCodeDetails(tenantName string, getAllReposSchema interface{}, actionType string) (models.Integration, interface{}, error) {
	integrationRes := models.Integration{}
	var allRepos interface{}
	for k, sourceCodeActions := range SourceCodeManagementFunctionsMap {
		switch k {
		case "github":
			if tenantIntegrations, ok := IntegrationsConcurrentCache.Load(tenantName); !ok {
				return models.Integration{}, map[string]string{}, fmt.Errorf("failed get source code %s branches: github integration does not exist", k)
			} else {
				if githubIntegration, ok := tenantIntegrations["github"].(models.Integration); ok {
					for a, f := range sourceCodeActions {
						switch a {
						case actionType:
							var schema interface{}
							if actionType == "get_all_repos" {
								schema = getAllReposSchema.(models.GetIntegrationDetailsSchema)
							} else if actionType == "get_all_branches" {
								schema = getAllReposSchema.(GetSourceCodeBranchesSchema)
							}
							integrationRes, allRepos, err := f.(func(models.Integration, interface{}) (models.Integration, interface{}, error))(githubIntegration, schema)
							if err != nil {
								return models.Integration{}, map[string]string{}, err
							}
							return integrationRes, allRepos, nil
						}

					}
				} else if !ok {
					return models.Integration{}, map[string]string{}, fmt.Errorf("failed get source code %s branches: github integration does not exist", k)
				}
			}
		default:
			return models.Integration{}, map[string]string{}, errors.New("failed get source branches : unsupported integration")
		}
	}
	return integrationRes, allRepos, nil
}

func orderBranchesPerConnectedRepos(connectedRepos []interface{}) map[string][]string {
	branchesPerRepo := map[string][]string{}
	for _, connectRepo := range connectedRepos {
		var connectedBranchList []string
		repo := connectRepo.(map[string]interface{})["repository"].(string)
		branch := connectRepo.(map[string]interface{})["branch"].(string)
		if _, ok := branchesPerRepo[repo]; !ok {
			connectedBranchList = append(connectedBranchList, branch)
			branchesPerRepo[repo] = connectedBranchList
		} else {
			connectedBranchList = append(branchesPerRepo[repo], branch)
			branchesPerRepo[repo] = connectedBranchList
		}
	}
	return branchesPerRepo
}
