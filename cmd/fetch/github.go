package fetch

import (
	"github.com/allero-io/allero/pkg/clients"
	"github.com/allero-io/allero/pkg/connectors"
	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

type FetchGithubDependencies struct {
	GithubClient    *github.Client
	OwnersWithRepos []*connectors.OwnerWithRepo
}

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
			ownersWithRepos := connectors.SplitParentRepo(args)
			githubClient := clients.CreateGithubClient(*deps.ConfigurationManager)

			fetchGithubDeps := &FetchGithubDependencies{
				GithubClient:    githubClient,
				OwnersWithRepos: ownersWithRepos,
			}

			return executeGithub(fetchGithubDeps)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			githubToken := deps.ConfigurationManager.GetGithubToken()
			tokenWasProvided := githubToken != ""

			analyticsArgs := make(map[string]any)
			analyticsArgs["Total Fetched Repos"] = reposFetchCounter
			analyticsArgs["Token Was Provided"] = tokenWasProvided
			deps.PosthogClient.PublishEventWithArgs("data fetched summary", analyticsArgs)
		},
	}

	return githubCmd
}

func executeGithub(deps *FetchGithubDependencies) error {
	githubConnectorDeps := &githubConnector.GithubConnectorDependencies{Client: deps.GithubClient}
	githubConnector := githubConnector.New(githubConnectorDeps)

	var err error
	reposFetchCounter, err = githubConnector.Get(deps.OwnersWithRepos)

	if err != nil {
		return err
	}

	return nil
}
