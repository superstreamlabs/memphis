package server

import (
	"context"
	"fmt"
	"memphis/models"
	"strings"

	"github.com/google/go-github/github"
)

func (s *Server) getGithubRepositories(integration models.Integration, body interface{}, user models.User) (models.Integration, interface{}, error) {
	ctx := context.Background()
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	integrationName := body.(models.GetIntegrationDetailsSchema).Name

	client, err := getGithubClient(integration.Keys["token"].(string), user)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]getGithubRepositories at getGithubClient: Integration %v: %v", user.TenantName, user.Username, integrationName, err.Error())
		return models.Integration{}, map[string]string{}, fmt.Errorf("getGithubRepositories at getGithubClient: Integration %v: %v", integrationName, err.Error())
	}
	branchesMap := make(map[string]string)

	for {
		repos, resp, err := client.Repositories.List(ctx, "", opt)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]getGithubRepositories at db.client.Repositories.List: Integration %v: %v", user.TenantName, user.Username, integrationName, err.Error())
			return models.Integration{}, map[string]string{}, fmt.Errorf("getGithubRepositories at db.client.Repositories.List: Integration %v: %v", integrationName, err.Error())
		}

		for _, repo := range repos {
			owner := repo.GetOwner().GetLogin()
			repoName := repo.GetName()
			branchesMap[repoName] = owner
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

	return integrationRes, branchesMap, nil
}

func (s *Server) getGithubBranches(integration models.Integration, body interface{}, user models.User) (models.Integration, interface{}, error) {
	branchesMap := make(map[string][]string)

	owner := body.(GetSourceCodeBranchesSchema).Owner
	repoName := body.(GetSourceCodeBranchesSchema).Name

	token := integration.Keys["token"].(string)
	connectedRepos := integration.Keys["connected_repos"].([]interface{})

	client, err := getGithubClient(token, user)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]getGithubBranches at getGithubClient: Integration %v: %v", user.TenantName, user.Username, "github", err.Error())
		return models.Integration{}, map[string][]string{}, err
	}
	branches, _, err := client.Repositories.ListBranches(context.Background(), owner, repoName, nil)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			serv.Errorf("[tenant: %v][user: %v]getGithubBranches at db.client.Repositories.ListBranches: Integration %v: %v", user.TenantName, user.Username, "github", "the repository does not exist")
			return models.Integration{}, map[string][]string{}, fmt.Errorf("the repository does not exist")
		}
		serv.Errorf("[tenant: %v][user: %v]getGithubBranches at db.client.Repositories.ListBranches: Integration %v: %v", user.TenantName, user.Username, "github", err.Error())
		return models.Integration{}, map[string][]string{}, err
	}

	// in case when connectedRepos holds multiple branches of the same repo
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

	branchInfoList := []string{}
	for _, branch := range branches {
		for repo, branches := range branchesPerRepo {
			if repo == repoName {
				isBranchExists := containsElement(branches, *branch.Name)
				if !isBranchExists {
					branchInfoList = append(branchInfoList, *branch.Name)
				}
			}
		}
	}

	if len(branchInfoList) > 0 {
		branchesMap[repoName] = branchInfoList
	}

	stringMapKeys := GetKeysAsStringMap(integration.Keys)
	cloneKeys := copyMaps(stringMapKeys)
	fmt.Print(cloneKeys)
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
