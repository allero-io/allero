package clients

import (
	"context"
	"net/http"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func CreateGithubClient(githubToken string) *github.Client {
	httpClient := &http.Client{}

	if githubToken != "" {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
		httpClient = oauth2.NewClient(context.Background(), tokenSource)
	}

	githubClient := github.NewClient(httpClient)

	return githubClient
}
