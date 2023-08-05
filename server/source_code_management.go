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
	"encoding/json"
	"errors"
	"fmt"
	"memphis/db"
	"memphis/models"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GetSourceCodeBranchesSchema struct {
	RepoName string `form:"repo_name" json:"repo_name" binding:"required"`
	Owner    string `form:"owner" json:"owner" binding:"required"`
}

func cacheDetailsGithub(keys map[string]interface{}, properties map[string]bool, tenantName string) {
	githubIntegration := models.Integration{}
	githubIntegration.Keys = make(map[string]interface{})
	githubIntegration.Properties = make(map[string]bool)
	if keys == nil {
		deleteIntegrationFromTenant(tenantName, "github", IntegrationsConcurrentCache)
		return
	}

	githubIntegration.Keys = keys
	githubIntegration.Name = "github"
	githubIntegration.TenantName = tenantName

	if _, ok := IntegrationsConcurrentCache.Load(tenantName); !ok {
		IntegrationsConcurrentCache.Add(tenantName, map[string]interface{}{"github": githubIntegration})
	} else {
		err := addIntegrationToTenant(tenantName, "github", IntegrationsConcurrentCache, githubIntegration)
		if err != nil {
			serv.Errorf("cacheDetailsGithub: %s ", err.Error())
			return
		}
	}
}

func createGithubIntegration(tenantName string, keys map[string]interface{}, properties map[string]bool) (models.Integration, error) {
	exist, githubIntegration, err := db.GetIntegration("github", tenantName)
	if err != nil {
		return models.Integration{}, err
	} else if !exist {
		integrationRes, insertErr := db.InsertNewIntegration(tenantName, "github", keys, properties)
		if insertErr != nil {
			if strings.Contains(insertErr.Error(), "already exists") {
				return models.Integration{}, errors.New("github integration already exists")
			} else {
				return models.Integration{}, insertErr
			}
		}
		githubIntegration = integrationRes
		integrationToUpdate := models.CreateIntegration{
			Name:       "github",
			Keys:       keys,
			Properties: properties,
			TenantName: tenantName,
		}
		msg, err := json.Marshal(integrationToUpdate)
		if err != nil {
			return models.Integration{}, err
		}
		err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
		if err != nil {
			return models.Integration{}, err
		}
		githubIntegration.Keys["token"] = hideIntegrationSecretKey(keys["token"].(string))
		return githubIntegration, nil
	}
	return models.Integration{}, errors.New("github integration already exists")
}

func (it IntegrationsHandler) handleCreateGithubIntegration(tenantName string, keys map[string]interface{}) (models.Integration, int, error) {
	statusCode, keys, err := it.handleGithubIntegration(tenantName, keys)
	if err != nil {
		return models.Integration{}, statusCode, err
	}

	keys, properties := createIntegrationsKeysAndProperties("github", "", "", false, false, false, "", "", "", "", "", "", keys["token"].(string), "", "", "", "")
	githubIntegration, err := createGithubIntegration(tenantName, keys, properties)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return models.Integration{}, SHOWABLE_ERROR_STATUS_CODE, err
		}
		return models.Integration{}, 500, err
	}
	return githubIntegration, statusCode, nil
}

func (it IntegrationsHandler) handleGithubIntegration(tenantName string, keys map[string]interface{}) (int, map[string]interface{}, error) {
	statusCode := 500
	if keys["token"] == "" {
		exist, integrationFromDb, err := db.GetIntegration("github", tenantName)
		if err != nil {
			return 500, map[string]interface{}{}, err
		}
		if !exist {
			statusCode = SHOWABLE_ERROR_STATUS_CODE
			return SHOWABLE_ERROR_STATUS_CODE, map[string]interface{}{}, errors.New("github integration does not exist")
		}
		keys["token"] = integrationFromDb.Keys["token"]
	} else {
		encryptedValue, err := EncryptAES([]byte(keys["token"].(string)))
		if err != nil {
			return 500, map[string]interface{}{}, err
		}
		keys["token"] = encryptedValue
	}
	return statusCode, keys, nil
}

func (it IntegrationsHandler) handleUpdateGithubIntegration(user models.User, body models.CreateIntegrationSchema) (models.Integration, int, error) {
	statusCode, keys, err := it.handleGithubIntegration(user.TenantName, body.Keys)
	if err != nil {
		return models.Integration{}, statusCode, err
	}
	githubIntegration, err := updateGithubIntegration(user, keys, map[string]bool{})
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return githubIntegration, SHOWABLE_ERROR_STATUS_CODE, err
		}
		return githubIntegration, 500, err
	}
	return githubIntegration, statusCode, nil
}

func updateGithubIntegration(user models.User, keys map[string]interface{}, properties map[string]bool) (models.Integration, error) {
	var githubIntegrationFromCache models.Integration
	if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(user.TenantName); ok {
		if githubIntegrationFromCache, ok = tenantInetgrations["github"].(models.Integration); !ok {
			return models.Integration{}, fmt.Errorf("github integration does not exist")
		}
	} else if !ok {
		return models.Integration{}, fmt.Errorf("github integration does not exist")
	}

	client, err := getGithubClient(githubIntegrationFromCache.Keys["token"].(string), user)
	if err != nil {
		return models.Integration{}, fmt.Errorf("updateGithubIntegration at getGithubClient: Integration %v: %v", "github", err.Error())
	}

	owner, ok := keys["owner"].(string)
	if !ok {
		userr, _, err := client.Users.Get(context.Background(), "")
		if err != nil {
			return models.Integration{}, fmt.Errorf("updateGithubIntegration at client.Users.Get : failed getting authenticated user: %v", err.Error())
		}
		owner = userr.GetLogin()
	}

	_, _, err = client.Repositories.Get(context.Background(), owner, keys["repo"].(string))
	if err != nil {
		return models.Integration{}, fmt.Errorf("updateGithubIntegration at Repositories.Get: Integration %v: repository does not exist %v", "github", err.Error())
	}

	updateIntegration := map[string]interface{}{}
	githubDetails := githubRepoDetails{
		Repository: keys["repo"].(string),
		Branch:     keys["branch"].(string),
		Type:       keys["type"].(string),
		Owner:      keys["owner"].(string),
	}

	updateIntegration["token"] = githubIntegrationFromCache.Keys["token"]
	if repos, ok := githubIntegrationFromCache.Keys["connected_repos"].([]interface{}); ok {
		if len(repos) > 0 {
			updateIntegration["connected_repos"] = append(repos, githubDetails)
		} else {
			updateIntegration["connected_repos"] = []githubRepoDetails{githubDetails}
		}
	}

	githubIntegration, err := db.UpdateIntegration(user.TenantName, "github", updateIntegration, properties)
	if err != nil {
		return models.Integration{}, fmt.Errorf("updateGithubIntegration at UpdateIntegration: Integration %v: %v", "github", err.Error())
	}

	integrationToUpdate := models.CreateIntegration{
		Name:       githubIntegration.Name,
		Keys:       githubIntegration.Keys,
		Properties: githubIntegration.Properties,
		TenantName: githubIntegration.TenantName,
	}

	msg, err := json.Marshal(integrationToUpdate)
	if err != nil {
		return models.Integration{}, fmt.Errorf("updateGithubIntegration at json.Marshal: Integration %v: %v", "github", err.Error())
	}
	err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		return models.Integration{}, fmt.Errorf("[updateGithubIntegration at sendInternalAccountMsgWithReply: Integration %v: %v", "github", err.Error())
	}

	githubIntegration.Keys["token"] = hideIntegrationSecretKey(githubIntegration.Keys["token"].(string))
	return githubIntegration, nil
}

func getGithubClient(token string, user models.User) (*github.Client, error) {
	key := getAESKey()
	decryptedValue, err := DecryptAES(key, token)
	if err != nil {
		return &github.Client{}, fmt.Errorf("getGithubClient at DecryptAES: Integration %v: %v", "github", err.Error())
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: decryptedValue},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client, nil
}

func getSourceCodeDetails(owner, repo, tenantName string, user models.User, getAllReposSchema interface{}, actionType string) (models.Integration, interface{}, error) {
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
							integrationRes, allRepos, err := f.(func(models.Integration, interface{}, models.User) (models.Integration, interface{}, error))(githubIntegration,
								schema, user)
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
