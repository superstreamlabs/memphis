package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"memphis/db"
	"memphis/models"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type githubRepoDetails struct {
	RepoName  string `json:"repo_name"`
	Branch    string `json:"branch"`
	Type      string `json:"type"`
	RepoOwner string `json:"repo_owner"`
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
	if _, ok := keys["token"]; !ok {
		keys["token"] = ""
	}
	if keys["token"] == "" {
		if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(tenantName); ok {
			if githubIntegrationFromCache, ok := tenantInetgrations["github"].(models.Integration); ok {
				keys["token"] = githubIntegrationFromCache.Keys["token"].(string)
			}
			if !ok || keys["token"] == "" {
				exist, integrationFromDb, err := db.GetIntegration("github", tenantName)
				if err != nil {
					return 500, map[string]interface{}{}, err
				}
				if !exist {
					statusCode = SHOWABLE_ERROR_STATUS_CODE
					return SHOWABLE_ERROR_STATUS_CODE, map[string]interface{}{}, errors.New("github integration does not exist")
				}
				keys["token"] = integrationFromDb.Keys["token"]
			}
		}
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
	if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(user.TenantName); ok {
		if _, ok = tenantInetgrations["github"].(models.Integration); !ok {
			return models.Integration{}, fmt.Errorf("github integration does not exist")
		}
	} else if !ok {
		return models.Integration{}, fmt.Errorf("github integration does not exist")
	}

	client, err := getGithubClient(keys["token"].(string))
	if err != nil {
		return models.Integration{}, err
	}

	updateIntegration := map[string]interface{}{}
	updateIntegration["token"] = keys["token"].(string)
	for _, key := range keys["connected_repos"].([]interface{}) {
		connectedRepoDetails := key.(map[string]interface{})
		var repoOwner string
		repoOwnerInterface, ok := connectedRepoDetails["repo_owner"].([]interface{})
		if ok {
			for _, owner := range repoOwnerInterface {
				repoOwner = owner.(string)
			}
		} else {
			userDetails, _, err := client.Users.Get(context.Background(), "")
			if err != nil {
				return models.Integration{}, err
			}
		} else {
			repoOwner = repoOwnerStr
		}

		_, _, err = client.Repositories.Get(context.Background(), repoOwner, connectedRepoDetails["repo_name"].(string))
		if err != nil {
			if strings.Contains(err.Error(), "Not Found") {
				updateIntegration["connected_repos"] = []githubRepoDetails{}
				continue
			} else {
				return models.Integration{}, fmt.Errorf("repository %s not found", connectedRepoDetails["repo_name"].(string))
			}
		}

		githubDetails := githubRepoDetails{
			RepoName:  connectedRepoDetails["repo_name"].(string),
			Branch:    connectedRepoDetails["branch"].(string),
			Type:      connectedRepoDetails["type"].(string),
			RepoOwner: repoOwner,
		}

		if connectedRepositories, ok := updateIntegration["connected_repos"].([]githubRepoDetails); !ok {
			updateIntegration["connected_repos"] = []githubRepoDetails{}
			updateIntegration["connected_repos"] = append(connectedRepositories, githubDetails)
		} else {
			updateIntegration["connected_repos"] = append(connectedRepositories, githubDetails)
		}
	}

	if len(keys["connected_repos"].([]interface{})) == 0 {
		updateIntegration["connected_repos"] = []githubRepoDetails{}
	}

	githubIntegration, err := db.UpdateIntegration(user.TenantName, "github", updateIntegration, properties)
	if err != nil {
		return models.Integration{}, err
	}

	integrationToUpdate := models.CreateIntegration{
		Name:       githubIntegration.Name,
		Keys:       githubIntegration.Keys,
		Properties: githubIntegration.Properties,
		TenantName: githubIntegration.TenantName,
	}

	msg, err := json.Marshal(integrationToUpdate)
	if err != nil {
		return models.Integration{}, err
	}
	err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		return models.Integration{}, err
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
	repositoriesMap := make(map[string][]string)

	for {
		repos, resp, err := client.Repositories.List(ctx, "", opt)
		if err != nil {
			return models.Integration{}, map[string]string{}, err
		}

		for _, repo := range repos {
			repoOwner := repo.GetOwner().GetLogin()
			repoName := repo.GetName()
			if _, exists := repositoriesMap[repoName]; exists {
				repositoriesMap[repoName] = append(repositoriesMap[repoName], repoOwner)
			} else {
				repositoriesMap[repoName] = []string{repoOwner}
			}
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

	opts := &github.ListOptions{PerPage: 100}
	var branches []*github.Branch
	var resp *github.Response
	for {
		branches, resp, err = client.Repositories.ListBranches(context.Background(), repoOwner, repoName, opts)
		if err != nil {
			if strings.Contains(err.Error(), "Not Found") {
				return models.Integration{}, map[string][]string{}, fmt.Errorf("The repository %s does not exist", repoName)
			}
			return models.Integration{}, map[string][]string{}, err
		}

		if resp.NextPage == 0 {
			// No more pages
			break
		}
		opts.Page = resp.NextPage
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
				for owner := range branches {
					if owner == repoOwner {
						isRepoExists = true
						isBranchExists := containsElement(branches[owner], *branch.Name)
						if !isBranchExists {
							branchInfoList = append(branchInfoList, *branch.Name)
						}
					}
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

func GetGithubContentFromConnectedRepo(githubIntegration models.Integration, connectedRepo map[string]interface{}, functionsDetails []functionDetails) ([]functionDetails, error) {
	token := githubIntegration.Keys["token"].(string)
	branch := connectedRepo["branch"].(string)
	repo := connectedRepo["repo_name"].(string)
	owner := connectedRepo["repo_owner"].(string)

	client, err := getGithubClient(token)
	if err != nil {
		return []functionDetails{}, err
	}

	_, repoContent, _, err := client.Repositories.GetContents(context.Background(), owner, repo, "", nil)
	if err != nil {
		return functionsDetails, err
	}

	for _, directoryContent := range repoContent {
		if directoryContent.GetType() == "dir" {
			_, filesContent, _, err := client.Repositories.GetContents(context.Background(), owner, repo, *directoryContent.Path, nil)
			if err != nil {
				continue
			}

			isValidFileYaml := false
			for _, fileContent := range filesContent {
				var content *github.RepositoryContent
				var commit *github.RepositoryCommit
				var contentMap map[string]interface{}
				if *fileContent.Type == "file" && strings.HasSuffix(*fileContent.Name, ".yaml") {
					content, _, _, err = client.Repositories.GetContents(context.Background(), owner, repo, *fileContent.Path, nil)
					if err != nil {
						continue
					}

					decodedContent, err := base64.StdEncoding.DecodeString(*content.Content)
					if err != nil {
						continue
					}

					err = yaml.Unmarshal(decodedContent, &contentMap)
					if err != nil {
						continue
					}

					err = validateYamlContent(contentMap)
					if err != nil {
						isValidFileYaml = false
						continue
					}
					isValidFileYaml = true

					commit, _, err = client.Repositories.GetCommit(context.Background(), owner, repo, branch)
					if err != nil {
						continue
					}

					if isValidFileYaml {
						fileDetails := functionDetails{
							Content:         content,
							Commit:          commit,
							ContentMap:      contentMap,
							RepoName:        repo,
							Branch:          branch,
							IntegrationName: githubIntegration.Name,
							Owner:           owner,
						}
						functionsDetails = append(functionsDetails, fileDetails)
						break
					}
				}
			}
			if !isValidFileYaml {
				continue
			}
		}
	}

	return functionsDetails, nil
}
