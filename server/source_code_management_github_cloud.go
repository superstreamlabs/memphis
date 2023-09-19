package server

import (
	"github.com/memphisdev/memphis/models"
)

type githubRepoDetails struct {
	RepoName  string `json:"repo_name"`
	Branch    string `json:"branch"`
	Type      string `json:"type"`
	RepoOwner string `json:"repo_owner"`
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
