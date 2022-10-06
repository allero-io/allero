package clients

import (
	"context"
	"net/http"

	"github.com/allero-io/allero/pkg/configurationManager"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func CreateGithubClient(configurationManager configurationManager.ConfigurationManager) *github.Client {
	githubToken := configurationManager.GetGithubToken()
	httpClient := &http.Client{}

	if githubToken != "" {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
		httpClient = oauth2.NewClient(context.Background(), tokenSource)
	}

	githubClient := github.NewClient(httpClient)

	return githubClient
}
