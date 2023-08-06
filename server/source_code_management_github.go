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

type githubRepoDetails struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Type       string `json:"type"`
	RepoOwner  string `json:"repo_owner"`
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
	err := testGithubAccessToken(keys["token"].(string))
	if err != nil {
		if strings.Contains(err.Error(), "access token is invalid") {
			return SHOWABLE_ERROR_STATUS_CODE, map[string]interface{}{}, err
		}
		return 500, map[string]interface{}{}, err
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
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access token is invalid") {
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

	client, err := getGithubClient(githubIntegrationFromCache.Keys["token"].(string))
	if err != nil {
		return models.Integration{}, fmt.Errorf("updateGithubIntegration at getGithubClient: %v", err.Error())
	}

	updateIntegration := map[string]interface{}{}
	connectedRepos := githubIntegrationFromCache.Keys["connected_repos"].([]interface{})
	for _, key := range keys["connected_repos"].([]interface{}) {
		keyObj := key.(map[string]interface{})
		repoOwner, ok := keyObj["repo_owner"].(string)
		if !ok {
			userDetails, _, err := client.Users.Get(context.Background(), "")
			if err != nil {
				return models.Integration{}, fmt.Errorf("updateGithubIntegration at client.Users.Get : failed getting authenticated user: %v", err.Error())
			}
			repoOwner = userDetails.GetLogin()
		}
		_, _, err = client.Repositories.Get(context.Background(), repoOwner, keyObj["repo_name"].(string))
		if err != nil {
			return models.Integration{}, fmt.Errorf("repository not found")
		}

		githubDetails := githubRepoDetails{
			Repository: keyObj["repo_name"].(string),
			Branch:     keyObj["branch"].(string),
			Type:       keyObj["type"].(string),
			RepoOwner:  keyObj["repo_owner"].(string),
		}

		branchesPerRepo := orderBranchesPerConnectedRepos(connectedRepos)
		updateIntegration["token"] = githubIntegrationFromCache.Keys["token"]
		if len(connectedRepos) > 0 {
			// ignore connected_repo that already exists
			isRepoExists := false
			for repo, branches := range branchesPerRepo {
				if repo == keyObj["repo_name"].(string) {
					isRepoExists = true
					isBranchExists := containsElement(branches, keyObj["branch"].(string))
					if !isBranchExists {
						if connectedRepositories, ok := updateIntegration["connected_repos"].([]interface{}); ok {
							updateIntegration["connected_repos"] = append(connectedRepositories, githubDetails)
						} else {
							updateIntegration["connected_repos"] = []interface{}{githubDetails}
						}
					}
				}
			}

			if !isRepoExists {
				updateIntegration["connected_repos"] = append(connectedRepos, githubDetails)
			}
		} else {
			if _, ok = updateIntegration["connected_repos"].([]githubRepoDetails); !ok {
				updateIntegration["connected_repos"] = []githubRepoDetails{}
			}
			updateIntegration["connected_repos"] = append(updateIntegration["connected_repos"].([]githubRepoDetails), githubDetails)
		}

	}

	if _, ok := updateIntegration["connected_repos"].([]interface{}); !ok {
		updateIntegration["connected_repos"] = connectedRepos
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

func testGithubAccessToken(token string) error {
	ctx := context.Background()
	client, err := getGithubClient(token)
	if err != nil {
		return err

	}
	// If the request was successful, the token is valid
	_, _, err = client.Users.Get(ctx, "")
	if err != nil {
		if strings.Contains(err.Error(), "Bad credentials") {
			return fmt.Errorf("The github access token is invalid")
		}
		return err

	}
	return nil
}

func getGithubClient(token string) (*github.Client, error) {
	key := getAESKey()
	decryptedValue, err := DecryptAES(key, token)
	if err != nil {
		return &github.Client{}, err
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: decryptedValue},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return client, nil
}

func (s *Server) getGithubRepositories(integration models.Integration, body interface{}) (models.Integration, interface{}, error) {
	ctx := context.Background()
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	client, err := getGithubClient(integration.Keys["token"].(string))
	if err != nil {
		return models.Integration{}, map[string]string{}, err
	}
	repositoriesMap := make(map[string]string)

	for {
		repos, resp, err := client.Repositories.List(ctx, "", opt)
		if err != nil {
			return models.Integration{}, map[string]string{}, err
		}

		for _, repo := range repos {
			repoOwner := repo.GetOwner().GetLogin()
			repoName := repo.GetName()
			repositoriesMap[repoName] = repoOwner
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		// Set the next page option to fetch the next page of results
		opt.Page = resp.NextPage
	}

	stringMapKeys := GetKeysAsStringMap(integration.Keys)
	cloneKeys := copyMaps(stringMapKeys)
	interfaceMapKeys := copyStringMapToInterfaceMap(cloneKeys)
	interfaceMapKeys["connected_repos"] = integration.Keys["connected_repos"]
	interfaceMapKeys["token"] = hideIntegrationSecretKey(interfaceMapKeys["token"].(string))
	integrationRes := models.Integration{
		Name:       integration.Name,
		Keys:       interfaceMapKeys,
		Properties: integration.Properties,
		TenantName: integration.TenantName,
	}

	return integrationRes, repositoriesMap, nil
}

func (s *Server) getGithubBranches(integration models.Integration, body interface{}) (models.Integration, interface{}, error) {
	branchesMap := make(map[string][]string)
	repoOwner := strings.ToLower(body.(GetSourceCodeBranchesSchema).RepoOwner)
	repoName := body.(GetSourceCodeBranchesSchema).RepoName

	token := integration.Keys["token"].(string)
	connectedRepos := integration.Keys["connected_repos"].([]interface{})

	client, err := getGithubClient(token)
	if err != nil {
		return models.Integration{}, map[string][]string{}, err
	}
	branches, _, err := client.Repositories.ListBranches(context.Background(), repoOwner, repoName, nil)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			return models.Integration{}, map[string][]string{}, fmt.Errorf("The repository does not exist %s ", repoName)
		}
		return models.Integration{}, map[string][]string{}, err
	}

	// in case when connectedRepos holds multiple branches of the same repo
	branchesPerRepo := orderBranchesPerConnectedRepos(connectedRepos)

	branchInfoList := []string{}
	for _, branch := range branches {
		isRepoExists := false
		if len(branchesPerRepo) == 0 {
			isRepoExists = true
			branchInfoList = append(branchInfoList, *branch.Name)
		}
		for repo, branches := range branchesPerRepo {
			if repo == repoName {
				isRepoExists = true
				isBranchExists := containsElement(branches, *branch.Name)
				if !isBranchExists {
					branchInfoList = append(branchInfoList, *branch.Name)
				}
			}
		}
		if !isRepoExists {
			branchInfoList = append(branchInfoList, *branch.Name)
		}
	}

	if len(branchInfoList) > 0 {
		branchesMap[repoName] = branchInfoList
	}

	stringMapKeys := GetKeysAsStringMap(integration.Keys)
	cloneKeys := copyMaps(stringMapKeys)
	interfaceMapKeys := copyStringMapToInterfaceMap(cloneKeys)
	interfaceMapKeys["connected_repos"] = integration.Keys["connected_repos"]
	interfaceMapKeys["token"] = hideIntegrationSecretKey(interfaceMapKeys["token"].(string))
	integrationRes := models.Integration{
		Name:       integration.Name,
		Keys:       interfaceMapKeys,
		Properties: integration.Properties,
		TenantName: integration.TenantName,
	}

	return integrationRes, branchesMap, nil
}

func containsElement(arr []string, val string) bool {
	for _, item := range arr {
		if item == val {
			return true
		}
	}
	return false
}
