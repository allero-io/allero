package fetch

import (
	"fmt"
	"os"

	"github.com/allero-io/allero/pkg/clients"
	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

var (
	reposFetchCounter int
)

type FetchGithubDependencies struct {
	GithubClient *github.Client
}

func NewGithubCommand(deps *FetchCommandDependencies) *cobra.Command {
	githubCmd := &cobra.Command{
		Use:   "github org/repo...",
		Short: "Fetch data of GitHub repositories",
		Long:  "Fetch data of GitHub repositories and entire organizations",
		Args:  cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			argsHead := []string{
				"github",
			}
			args = append(argsHead, args...)
			deps.PosthogClient.PublishCmdUse("data fetched", args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			githubToken := deps.ConfigurationManager.GetGithubToken()
			githubClient := clients.CreateGithubClient(githubToken)

			fetchGithubDeps := &FetchGithubDependencies{
				GithubClient: githubClient,
			}

			return execute(fetchGithubDeps, args)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			_, tokenWasProvided := os.LookupEnv("GITHUB_TOKEN")
			analyticsArgs := []string{
				"reposFetchedCounter=" + fmt.Sprint(reposFetchCounter),
				"tokenWasProvided=" + fmt.Sprint(tokenWasProvided),
			}
			deps.PosthogClient.PublishCmdUse("data fetched summary", analyticsArgs)
		},
	}

	return githubCmd
}

func execute(deps *FetchGithubDependencies, args []string) error {

	githubConnectorDeps := &githubConnector.GithubConnectorDependencies{Client: deps.GithubClient}
	githubConnector := githubConnector.New(githubConnectorDeps)
	var err error
	reposFetchCounter, err = githubConnector.Get(args)
	if err != nil {
		return err
	}

	return nil
}
