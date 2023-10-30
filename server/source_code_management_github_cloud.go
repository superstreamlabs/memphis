package server

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/memphisdev/memphis/models"
	"gopkg.in/yaml.v2"

	"github.com/google/go-github/github"
)

const (
	memphisDevFunctionsRepoName   = "memphis-dev-functions"
	memphisDevFunctionsOwnerName  = "memphisdev"
	memphisDevFunctionsBranchName = "master"
)

var memphisFunctions = map[string]interface{}{
	"repo_name":  memphisDevFunctionsRepoName,
	"branch":     "master",
	"type":       "functions",
	"repo_owner": memphisDevFunctionsOwnerName,
}

type githubRepoDetails struct {
	RepoName  string `json:"repo_name"`
	Branch    string `json:"branch"`
	Type      string `json:"type"`
	RepoOwner string `json:"repo_owner"`
}

func (it IntegrationsHandler) handleCreateGithubIntegration(tenantName string, keys map[string]interface{}) (models.Integration, int, error) {
	return models.Integration{}, 0, nil
}

func (it IntegrationsHandler) handleUpdateGithubIntegration(user models.User, body models.CreateIntegrationSchema) (models.Integration, int, error) {
	return models.Integration{}, 0, nil

}
func cacheDetailsGithub(keys map[string]interface{}, properties map[string]bool, tenantName string) {
	return
}

func getGithubClientWithoutAccessToken() *github.Client {
	client := github.NewClient(nil)
	return client
}

func getGithubClient(tenantName string) (string, string, *github.Client, error) {
	client := getGithubClientWithoutAccessToken()
	return "", "", client, nil
}

func testGithubIntegration(installationId string) error {
	return nil
}

func (s *Server) getGithubRepositories(integration models.Integration, body interface{}) (models.Integration, interface{}, error) {
	return models.Integration{}, nil, nil
}

func (s *Server) getGithubBranches(integration models.Integration, body interface{}) (models.Integration, interface{}, error) {
	return models.Integration{}, nil, nil
}

func containsElement(arr []string, val string) bool {
	for _, item := range arr {
		if item == val {
			return true
		}
	}
	return false
}

func GetGithubContentFromConnectedRepo(connectedRepo map[string]interface{}, functionsDetails map[string][]functionDetails, tenantName string) (map[string][]functionDetails, error) {
	branch := connectedRepo["branch"].(string)
	repo := connectedRepo["repo_name"].(string)
	owner := connectedRepo["repo_owner"].(string)

	var client *github.Client
	var err error
	client = getGithubClientWithoutAccessToken()
	_, repoContent, _, err := client.Repositories.GetContents(context.Background(), owner, repo, "", &github.RepositoryContentGetOptions{
		Ref: branch})
	if err != nil {
		return functionsDetails, err
	}

	countFunctions := 0
	for _, directoryContent := range repoContent {
		// In order to restrict the api calls per repo
		if countFunctions == 10 {
			break
		}
		if directoryContent.GetType() == "dir" {
			_, filesContent, _, err := client.Repositories.GetContents(context.Background(), owner, repo, *directoryContent.Path, &github.RepositoryContentGetOptions{
				Ref: branch})
			if err != nil {
				continue
			}

			isValidFileYaml := false
			for _, fileContent := range filesContent {
				var content *github.RepositoryContent
				var commit *github.RepositoryCommit
				var contentMap map[string]interface{}
				if *fileContent.Type == "file" && *fileContent.Name == "memphis.yaml" {
					content, _, _, err = client.Repositories.GetContents(context.Background(), owner, repo, *fileContent.Path, &github.RepositoryContentGetOptions{
						Ref: branch})
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

					if _, ok := contentMap["memory"]; !ok || contentMap["memory"] == "" {
						contentMap["memory"] = int64(128) * 1024 * 1024
					}

					if _, ok := contentMap["storage"]; !ok || contentMap["storage"] == "" {
						contentMap["storage"] = int64(512) * 1024 * 1024
					}

					if contentMap["dependencies"].(string) == "" {
						switch contentMap["language"] {
						case "go":
							contentMap["dependencies"] = "go.mod"
						case "nodejs":
							contentMap["dependencies"] = "package.json"
						case "python":
							contentMap["dependencies"] = "requirements.txt"
						}
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
						countFunctions++
						fileDetails := functionDetails{
							Commit:       commit,
							ContentMap:   contentMap,
							RepoName:     repo,
							Branch:       branch,
							Owner:        owner,
							DirectoryUrl: directoryContent.HTMLURL,
							TenantName:   tenantName,
						}
						functionsDetails["other"] = append(functionsDetails["other"], fileDetails)
						break
					}
					splitPath := strings.Split(*fileContent.Path, "/")
					path := strings.TrimSpace(splitPath[0])
					if path != contentMap["function_name"].(string) {
						serv.Warnf("[tenant: %v]]GetGithubContentFromConnectedRepo: In the repository %s, function name %s in git doesn't match the function_name field %s in YAML file.", tenantName, repo, splitPath[0], contentMap["function_name"].(string))
						continue
					}
					if strings.Contains(path, "") {
						serv.Warnf("[tenant: %v]GetGithubContentFromConnectedRepo: In the repository %s, the function name %s in the YAML file cannot contain spaces", tenantName, repo, contentMap["function_name"].(string))
						continue
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

func deleteInstallationForAuthenticatedGithubApp(tenantName string) error {
	return nil
}
func getGithubKeys(githubIntegrationDetails map[string]interface{}, repoOwner, repo, branch, repoType string) map[string]interface{} {
	return map[string]interface{}{}
}

func retrieveGithubAppName() string {
	return ""
}
