package clients

import (
	"github.com/allero-io/allero/pkg/configurationManager"
	"github.com/xanzy/go-gitlab"
)

func CreateGitlabClient(configurationManager configurationManager.ConfigurationManager) (*gitlab.Client, error) {
	GITLAB_TOKEN := configurationManager.GetGitlabToken()
	gitlabClient, err := gitlab.NewClient(GITLAB_TOKEN)
	if err != nil {
		return nil, err
	}

	return gitlabClient, nil
}
