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
	"context"
	"encoding/base64"
	"time"

	"fmt"
	"memphis/models"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"gopkg.in/yaml.v2"
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
		serv.Errorf("GetAllFunctions at GetFunctionsDetails: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	functions := []FunctionsResult{}

	for k := range SourceCodeManagementFunctionsMap {
		if tenantIntegrations, ok := IntegrationsConcurrentCache.Load(user.TenantName); !ok {
			serv.Errorf("GetAllFunctions at Load: %v", err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		} else {
			if integration, ok := tenantIntegrations[k].(models.Integration); ok {
				selectedRepos, _ := fh.GetSelectedSourceCodeRepos(integration)
				var repoContent interface{}
				for _, connectedRepo := range selectedRepos {
					connectedRepoRes := connectedRepo.(map[string]interface{})
					repo := connectedRepoRes["repo_name"].(string)
					owner := connectedRepoRes["repo_owner"].(string)
					branch := connectedRepoRes["branch"].(string)

					repoContent, _ = fh.GetContentOfSelectedRepos(repo, owner, k, integration)

					switch k {
					case "github":
						client, _ := getGithubClient(integration.Keys["token"].(string))
						for _, directoryContent := range repoContent.([]*github.RepositoryContent) {
							if directoryContent.GetType() == "dir" {
								_, filesContent, _, err := client.Repositories.GetContents(context.Background(), owner, repo, *directoryContent.Path, nil)
								if err != nil {
									serv.Errorf("[tenant: %v][user: %v]GetAllFunctions at GetContents: %v", user.TenantName, user.Username, err.Error())
									c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
									return
								}

								isValidFileYaml := false
								for _, fileContent := range filesContent {
									if *fileContent.Type == "file" && strings.HasSuffix(*fileContent.Name, ".yaml") {
										content, _, _, err := client.Repositories.GetContents(context.Background(), owner, repo, *fileContent.Path, nil)
										if err != nil {
											serv.Errorf("[tenant: %v][user: %v]GetAllFunctions at GetContents: %v", user.TenantName, user.Username, err.Error())
											c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
											return
										}

										decodedContent, err := base64.StdEncoding.DecodeString(*content.Content)
										if err != nil {
											serv.Errorf("[tenant: %v][user: %v]GetAllFunctions at DecodeString: %v", user.TenantName, user.Username, err.Error())
											c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
											return
										}

										var contentMap map[string]interface{}
										err = yaml.Unmarshal(decodedContent, &contentMap)
										if err != nil {
											serv.Errorf("[tenant: %v][user: %v]GetAllFunctions at yaml.Unmarshal: %v", user.TenantName, user.Username, err.Error())
											c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
											return
										}

										err = validateYamlContent(contentMap)
										if err != nil {
											isValidFileYaml = false
											serv.Warnf("[tenant: %v][user: %v]GetAllFunctions at validateYamlContent: %v", user.TenantName, user.Username, err.Error())
											continue
										}
										isValidFileYaml = true
										tagsInterfaceSlice := contentMap["tags"].([]interface{})
										tagsStrings := make([]string, len(contentMap["tags"].([]interface{})))

										for i, v := range tagsInterfaceSlice {
											if str, ok := v.(string); ok {
												tagsStrings[i] = str
											}
										}

										fileYaml := ContentYamlFile{
											FunctionName: contentMap["function_name"].(string),
											Description:  contentMap["description"].(string),
											Tags:         tagsStrings,
											Language:     contentMap["language"].(string),
										}

										commit, _, err := client.Repositories.GetCommit(context.Background(), owner, repo, branch)
										if err != nil {
											serv.Errorf("[tenant: %v][user: %v]GetAllFunctions at GetCommit: %v", user.TenantName, user.Username, err.Error())
											c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
											return
										}

										res := FunctionsResult{
											FunctionName: fileYaml.FunctionName,
											Description:  fileYaml.Description,
											Tags:         fileYaml.Tags,
											Language:     fileYaml.Language,
											LastCommit:   *commit.Commit.Committer.Date,
											Link:         *fileContent.HTMLURL,
											Repository:   repo,
											Branch:       branch,
										}

										functions = append(functions, res)
										if isValidFileYaml {
											break
										}
									}
								}
								if !isValidFileYaml {
									serv.Warnf("[tenant: %v][user: %v]GetAllFunctions: %v", user.TenantName, user.Username, "You must include in your repo directory that includes yaml file")
									continue
								}
							}

						}
					}
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

func (fh FunctionsHandler) GetSelectedSourceCodeRepos(integration models.Integration) ([]interface{}, error) {
	connectedRepos := integration.Keys["connected_repos"].([]interface{})
	fmt.Println(connectedRepos)
	selectedRepos := []interface{}{}
	for _, repo := range connectedRepos {
		fmt.Println(repo)
		repository := repo.(map[string]interface{})
		repoType := repository["type"].(string)
		fmt.Println(repoType)
		if repoType == "functions" {
			selectedRepos = append(selectedRepos, repo)
		}
	}
	return selectedRepos, nil
}

func (fh FunctionsHandler) GetContentOfSelectedRepos(repo, owner string, sourceCodeIntegrationType string, integration models.Integration) ([]*github.RepositoryContent, error) {
	repos, _ := GetContentDetails(repo, owner, sourceCodeIntegrationType, integration)
	return repos, nil
}
