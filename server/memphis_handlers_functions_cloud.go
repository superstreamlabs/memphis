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
	"fmt"
	"regexp"
	"time"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/memphisdev/memphis/models"
)

type FunctionsHandler struct{}

func (fh FunctionsHandler) GetAllFunctions(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetAllFunctions at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	functionsResult, err := fh.GetFunctions(user.TenantName)
	if err != nil {
		if strings.Contains(err.Error(), "403 API rate limit exceeded") {
			serv.Warnf("[tenant: %v][user: %v]GetAllFunctions at GetFunctions: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Github's rate limit has been reached, please try again in 1 hour"})
			return
		} else {
			serv.Errorf("[tenant: %v][user: %v]GetAllFunctions at GetFunctions: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	c.IndentedJSON(200, gin.H{"scm_integrated": functionsResult.ScmIntegrated, "other": functionsResult.OtherFunctions, "installed": functionsResult.InstalledFunctions, "connected_repos": functionsResult.ConnectedRepos})
}

func (fh FunctionsHandler) GetFunctions(tenantName string) (models.FunctionsRes, error) {
	contentDetailsOfSelectedRepos, scmIntegrated, err := GetContentOfSelectedRepos(tenantName)
	if err != nil {
		return models.FunctionsRes{}, err
	}
	functions, err := GetFunctionsDetails(contentDetailsOfSelectedRepos)
	if err != nil {
		return models.FunctionsRes{}, err
	}

	installedFunctions := functions["installed"]
	OtherFunctions := functions["other"]
	if len(installedFunctions) == 0 {
		installedFunctions = []models.FunctionResult{}
	}

	if len(OtherFunctions) == 0 {
		OtherFunctions = []models.FunctionResult{}
	}

	var lastModified *time.Time
	OtherFunctions = []models.FunctionResult{}
	for _, function := range functions["other"] {
		if function.Owner == memphisDevFunctionsOwnerName && function.Repo == memphisDevFunctionsRepoName {
			otherFunctionResult := models.FunctionResult{
				FunctionName:               function.FunctionName,
				Description:                function.Description,
				Tags:                       function.Tags,
				Runtime:                    function.Runtime,
				Dependencies:               function.Dependencies,
				Inputs:                     function.Inputs,
				Memory:                     function.Memory,
				Storage:                    function.Storage,
				Handler:                    function.Handler,
				Scm:                        "github",
				Repo:                       function.Repo,
				Branch:                     function.Branch,
				Owner:                      function.Owner,
				Language:                   function.Language,
				Version:                    -1,
				IsValid:                    function.IsValid,
				InvalidReason:              function.InvalidReason,
				InProgress:                 false,
				UpdatesAvailable:           false,
				ByMemphis:                  function.ByMemphis,
				TenantName:                 function.TenantName,
				ClonedUpdatesAvailable:     function.ClonedUpdatesAvailable,
				ClonedUpdatesIsValid:       function.ClonedUpdatesIsValid,
				ClonedUpdatesInvalidReason: function.ClonedUpdatesInvalidReason,
			}
			OtherFunctions = append(OtherFunctions, otherFunctionResult)
			lastModified = function.LastCommit
		}
	}

	memphisDevFucntions := []map[string]interface{}{}
	memphisFunc := map[string]interface{}{
		"repo_name":     memphisFunctions["repo_name"].(string),
		"branch":        memphisFunctions["branch"].(string),
		"owner":         memphisFunctions["repo_owner"].(string),
		"last_modified": lastModified,
		"in_progress":   false,
	}
	memphisDevFucntions = append(memphisDevFucntions, memphisFunc)

	allFunctions := models.FunctionsRes{
		InstalledFunctions: installedFunctions,
		OtherFunctions:     OtherFunctions,
		ScmIntegrated:      scmIntegrated,
		ConnectedRepos:     memphisDevFucntions,
	}

	return allFunctions, nil
}

func (mh MonitoringHandler) GetFunctionsOverview(sName, tenantName string, partition int) (FunctionsOverviewResponse, error) {
	return FunctionsOverviewResponse{}, nil
}

func validateYamlContent(yamlMap map[string]interface{}) error {
	requiredFields := []string{"function_name", "runtime", "dependencies"}
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

func GetFunctionsDetails(functionsDetails map[string][]functionDetails) (map[string][]models.FunctionResult, error) {
	functions := map[string][]models.FunctionResult{}
	for key, functionDetails := range functionsDetails {
		for _, funcDetailsPerInstalled := range functionDetails {
			fucntionContentMap := funcDetailsPerInstalled.ContentMap
			commit := funcDetailsPerInstalled.Commit
			link := funcDetailsPerInstalled.DirectoryUrl
			repo := funcDetailsPerInstalled.RepoName
			branch := funcDetailsPerInstalled.Branch
			owner := funcDetailsPerInstalled.Owner
			tenantName := funcDetailsPerInstalled.TenantName
			isValid := funcDetailsPerInstalled.IsValid
			invalidReason := funcDetailsPerInstalled.InvalidReason
			tagsInterfaceSlice, ok := fucntionContentMap["tags"].([]interface{})
			tagsStrings := []string{}
			if ok {
				tagsStrings = make([]string, len(fucntionContentMap["tags"].([]interface{})))
				for i, tag := range tagsInterfaceSlice {
					tagMap := tag.(map[interface{}]interface{})
					for _, v := range tagMap {
						if str, ok := v.(string); ok {
							tagsStrings[i] = str
						}
					}
				}
			}

			inputs := []map[string]string{}
			inputsInterfaceSlice, ok := fucntionContentMap["inputs"].([]interface{})
			if ok {
				for _, environmentVarInterface := range inputsInterfaceSlice {
					environmentVarMap, _ := environmentVarInterface.(map[interface{}]interface{})
					environmentVar := make(map[string]string)
					for k, v := range environmentVarMap {
						if key, ok := k.(string); ok {
							if val, ok := v.(string); ok {
								environmentVar[key] = val
							}
						}
					}
					inputs = append(inputs, environmentVar)
				}
			}

			description := ""
			descriptionInterface, ok := fucntionContentMap["description"]
			if ok {
				description = descriptionInterface.(string)
			}

			functionName := ""
			if functionNameInterface, ok := fucntionContentMap["function_name"]; !ok || functionNameInterface == nil || functionNameInterface.(string) == "" {
				errMsg := fmt.Errorf("function in %s repository is invalid since its memphis.yaml file is missing the function_name field", repo)
				return functions, errMsg
			} else {
				functionName = functionNameInterface.(string)
			}

			runtime := ""
			if runtimeInterface, ok := fucntionContentMap["runtime"]; !ok || runtimeInterface == nil || runtimeInterface.(string) == "" {
				errMsg := fmt.Errorf("function %s placed in %s repository is invalid since its memphis.yaml file is missing the runtime field", repo, functionName)
				return functions, errMsg
			} else {
				runtime = runtimeInterface.(string)
			}
			regex := regexp.MustCompile(`[0-9]+|\\.$`)
			language := regex.ReplaceAllString(runtime, "")
			language = strings.TrimRight(language, ".")
			if strings.Contains(language, "-edge") {
				language = strings.Trim(language, ".-edge")
			}

			dependencies := ""
			dependenciesMissing := false
			if dependenciesInterface, ok := fucntionContentMap["dependencies"]; !ok || dependenciesInterface == nil || dependenciesInterface.(string) == "" {
				dependenciesMissing = true
			} else {
				dependencies = dependenciesInterface.(string)
			}

			if dependenciesMissing {
				switch language {
				case "go":
					dependencies = "go.mod"
				case "nodejs":
					dependencies = "package.json"
				case "python":
					dependencies = "requirements.txt"
				}
			}

			byMemphis := false
			if repo == memphisDevFunctionsRepoName && owner == memphisDevFunctionsOwnerName {
				byMemphis = true
			}

			handler := ""
			if _, ok := fucntionContentMap["handler"].(string); ok {
				handler = fucntionContentMap["handler"].(string)
			}
			var lastCommit *time.Time
			if commit != nil {
				lastCommit = commit.Commit.Committer.Date
			}
			memory := 128 * 1024 * 1024
			if memoryInterface, ok := fucntionContentMap["memory"]; ok && memoryInterface != nil {
				if memoryVal, ok := memoryInterface.(int); ok {
					memory = memoryVal
				}
			}

			storage := 512 * 1024 * 1024
			if storageInterface, ok := fucntionContentMap["storage"]; ok && storageInterface != nil {
				if storageVal, ok := storageInterface.(int); ok {
					storage = storageVal
				}
			}

			functionDetails := models.FunctionResult{
				FunctionName:               functionName,
				Description:                description,
				Tags:                       tagsStrings,
				Runtime:                    runtime,
				Dependencies:               dependencies,
				Inputs:                     inputs,
				Memory:                     memory,
				Storage:                    storage,
				Handler:                    handler,
				Scm:                        "github",
				Repo:                       repo,
				Branch:                     branch,
				Owner:                      owner,
				LastCommit:                 lastCommit,
				Link:                       link,
				Language:                   language,
				InProgress:                 false,
				UpdatesAvailable:           false,
				ByMemphis:                  byMemphis,
				TenantName:                 tenantName,
				IsValid:                    isValid,
				InvalidReason:              invalidReason,
				ClonedUpdatesAvailable:     false,
				ClonedUpdatesIsValid:       isValid,
				ClonedUpdatesInvalidReason: invalidReason,
			}

			functions[key] = append(functions[key], functionDetails)
		}
	}
	return functions, nil
}
