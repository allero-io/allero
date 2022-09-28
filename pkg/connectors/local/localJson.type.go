package localConnector

import (
	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
)

type LocalRoot struct {
	GithubData map[string]*githubConnector.GithubOwner `json:"local_github"`
	GitlabData map[string]*gitlabConnector.GitlabGroup `json:"local_gitlab"`
}
