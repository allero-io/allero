package fetch

import (
	"github.com/allero-io/allero/pkg/clients"
	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

type FetchGithubDependencies struct {
	GithubClient *github.Client
}

var GITHUB_TOKEN string

func NewGithubCommand(deps *FetchCommandDependencies) *cobra.Command {
	githubCmd := &cobra.Command{
		Use:   "github org/repo...",
		Short: "Fetch data of GitHub repositories",
		Long:  "Fetch data of GitHub repositories and entire organizations",
		Args:  cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, cmdArgs []string) {
			args := make(map[string]any)
			args["Platform"] = "github"
			args["Args"] = cmdArgs
			decodedToken, _ := deps.ConfigurationManager.ParseToken()
			if decodedToken != nil {
				args["User Email"] = decodedToken.Email
			}
			deps.PosthogClient.PublishEventWithArgs("data fetched", args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			GITHUB_TOKEN = deps.ConfigurationManager.GetGithubToken()
			githubClient := clients.CreateGithubClient(GITHUB_TOKEN)

			fetchGithubDeps := &FetchGithubDependencies{
				GithubClient: githubClient,
			}

			return executeGithub(fetchGithubDeps, args)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			tokenWasProvided := GITHUB_TOKEN != ""

			analyticsArgs := make(map[string]any)
			analyticsArgs["Total Fetched Repos"] = reposFetchCounter
			analyticsArgs["Token Was Provided"] = tokenWasProvided
			deps.PosthogClient.PublishEventWithArgs("data fetched summary", analyticsArgs)
		},
	}

	return githubCmd
}

func executeGithub(deps *FetchGithubDependencies, args []string) error {
	githubConnectorDeps := &githubConnector.GithubConnectorDependencies{Client: deps.GithubClient}
	githubConnector := githubConnector.New(githubConnectorDeps)
	reposFetchCounter, err = githubConnector.Get(args)
	if err != nil {
		return err
	}

	return nil
}
