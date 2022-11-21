package clients

import (
	"fmt"
	"github.com/allero-io/allero/pkg/configurationManager"
	"github.com/xanzy/go-gitlab"
)

func CreateGitlabClient(configurationManager configurationManager.ConfigurationManager) (*gitlab.Client, error) {
	GITLAB_TOKEN := configurationManager.GetGitlabToken()
	gitlabPrivateServerUrl := "" 

	gitlabCustomServerValue, err := configurationManager.Get("gitlabClientURL")
	if err != nil {
		return nil, err
	}

	if gitlabCustomServerValue != nil {
		gitlabPrivateServerUrl = fmt.Sprintf("%s", gitlabCustomServerValue)
	}

	gitlabClient, err := generateGitlabClient(GITLAB_TOKEN, gitlabPrivateServerUrl) 
	if err != nil {
		return nil, err
	}

	return gitlabClient, nil
}

func generateGitlabClient(gitlabToken, gitlabPrivateServerUrl string) (*gitlab.Client, error) {
	if gitlabPrivateServerUrl != "" {
		gitlabClient, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabPrivateServerUrl))
		if err != nil {
			return nil, err
		}
		return gitlabClient, nil
	}

	gitlabClient, err := gitlab.NewClient(gitlabToken)
	if err != nil {
		return nil, err
	}
	return gitlabClient, nil
}