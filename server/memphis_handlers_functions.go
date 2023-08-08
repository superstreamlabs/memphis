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
	"time"

	"fmt"
	"memphis/models"

	"github.com/gin-gonic/gin"
)

type ContentYamlFile struct {
	FunctionName string   `json:"function_name" yaml:"function_name"`
	Description  string   `json:"description" yaml:"description"`
	Tags         []string `json:"tags" yaml:"tags"`
	Language     string   `json:"language" yaml:"language"`
}

type FunctionsResult struct {
	FunctionName string    `json:"function_name"`
	Description  string    `json:"description"`
	Tags         []string  `json:"tags"`
	Language     string    `json:"language"`
	LastCommit   time.Time `json:"last_commit"`
	Link         string    `json:"link"`
	Repository   string    `json:"repository"`
	Branch       string    `json:"branch"`
}

type FunctionsHandler struct{}

func (fh FunctionsHandler) GetAllFunctions(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetAllFunctions at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	functions := []FunctionsResult{}

	if tenantIntegrations, ok := IntegrationsConcurrentCache.Load(user.TenantName); ok {
		if integration, ok := tenantIntegrations["github"].(models.Integration); ok {
			connectedRepos, err := fh.GetConnectedSourceCodeRepos(integration)
			if err != nil {
				serv.Errorf("[tenant: %v][user: %v]GetAllFunctions at GetConnectedSourceCodeRepos: %v", user.TenantName, user.Username, err.Error())
				return
			}
			for _, connectedRepo := range connectedRepos {
				connectedRepoRes := connectedRepo.(map[string]interface{})
				contentDetails, err := GetContentOfSelectedRepos(integration.Name, integration, connectedRepoRes)
				if err != nil {
					serv.Errorf("[tenant: %v][user: %v]GetAllFunctions at GetContentOfSelectedRepos: %v", user.TenantName, user.Username, err.Error())
					continue
				}
				functions, err = GetFunctionsDetails(contentDetails, integration.Name, functions, connectedRepoRes["repo_name"].(string), connectedRepoRes["branch"].(string))
				if err != nil {
					serv.Errorf("[tenant: %v][user: %v]GetAllFunctions at GetFunctionsDetails: %v", user.TenantName, user.Username, err.Error())
					continue
				}
			}
		}
	}
	c.JSON(200, functions)
}

func validateYamlContent(yamlMap map[string]interface{}) error {
	requiredFields := []string{"function_name", "description", "tags", "language"}
	missingFields := make([]string, 0)
	for _, field := range requiredFields {
		if _, exists := yamlMap[field]; !exists {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("Missing fields: %v\n", missingFields)
	}
	return nil
}

func (fh FunctionsHandler) GetConnectedSourceCodeRepos(integration models.Integration) ([]interface{}, error) {
	connectedRepos := integration.Keys["connected_repos"].([]interface{})
	selectedRepos := []interface{}{}
	for _, repo := range connectedRepos {
		repository := repo.(map[string]interface{})
		repoType := repository["type"].(string)
		if repoType == "functions" {
			selectedRepos = append(selectedRepos, repo)
		}
	}
	return selectedRepos, nil
}

func GetFunctionsDetails(contentDetails []fileContentDetails, sourceCodeTypeIntegration string, functions []FunctionsResult, repo, branch string) ([]FunctionsResult, error) {
	switch sourceCodeTypeIntegration {
	case "github":
		for _, fileDetails := range contentDetails {
			contentMapContent := fileDetails.ContentMap
			commit := fileDetails.Commit
			fileContent := fileDetails.Content
			tagsInterfaceSlice := contentMapContent["tags"].([]interface{})
			tagsStrings := make([]string, len(contentMapContent["tags"].([]interface{})))

			for i, v := range tagsInterfaceSlice {
				if str, ok := v.(string); ok {
					tagsStrings[i] = str
				}
			}

			fileYaml := ContentYamlFile{
				FunctionName: contentMapContent["function_name"].(string),
				Description:  contentMapContent["description"].(string),
				Tags:         tagsStrings,
				Language:     contentMapContent["language"].(string),
			}

			functionDetails := FunctionsResult{
				FunctionName: fileYaml.FunctionName,
				Description:  fileYaml.Description,
				Tags:         fileYaml.Tags,
				Language:     fileYaml.Language,
				LastCommit:   *commit.Commit.Committer.Date,
				Link:         *fileContent.HTMLURL,
				Repository:   repo,
				Branch:       branch,
			}

			functions = append(functions, functionDetails)
		}
	}
	return functions, nil
}
