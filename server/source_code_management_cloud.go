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
	"github.com/google/go-github/github"
	"github.com/memphisdev/memphis/models"
)

type GetSourceCodeBranchesSchema struct {
	RepoName  string `form:"repo_name" json:"repo_name" binding:"required"`
	RepoOwner string `form:"repo_owner" json:"repo_owner" binding:"required"`
}

type functionDetails struct {
	Content    *github.RepositoryContent `json:"content"`
	Commit     *github.RepositoryCommit  `json:"commit"`
	ContentMap map[string]interface{}    `json:"content_map"`
	RepoName   string                    `json:"repo_name"`
	Branch     string                    `json:"branch"`
	Scm        string                    `json:"scm"`
	Owner      string                    `json:"owner"`
}

func getSourceCodeDetails(tenantName string, getAllReposSchema interface{}, actionType string) (models.Integration, interface{}, error) {
	return models.Integration{}, map[string]string{}, nil
}

func GetContentOfSelectedRepo(connectedRepo map[string]interface{}, contentDetails []functionDetails) ([]functionDetails, error) {
	var err error
	contentDetails, err = GetGithubContentFromConnectedRepo(connectedRepo, contentDetails)
	if err != nil {
		return contentDetails, err
	}

	return contentDetails, nil
}

func getConnectedSourceCodeRepos(tenantName string) (map[string][]interface{}, bool) {
	selectedReposPerSourceCodeIntegration := map[string][]interface{}{}
	scmIntegrated := false
	selectedRepos := []interface{}{}
	selectedRepos = append(selectedRepos, memphisFunctions)
	selectedReposPerSourceCodeIntegration["memphis_functions"] = selectedRepos

	return selectedReposPerSourceCodeIntegration, scmIntegrated
}

func GetContentOfSelectedRepos(tenantName string) ([]functionDetails, bool, error) {
	contentDetails := []functionDetails{}
	connectedRepos, scmIntegrated := getConnectedSourceCodeRepos(tenantName)
	var err error
	for _, connectedRepoPerIntegration := range connectedRepos {
		for _, connectedRepo := range connectedRepoPerIntegration {
			connectedRepoRes := connectedRepo.(map[string]interface{})
			contentDetails, err = GetContentOfSelectedRepo(connectedRepoRes, contentDetails)
			if err != nil {
				return contentDetails, scmIntegrated, err
			}
		}
	}
	return contentDetails, scmIntegrated, nil
}
