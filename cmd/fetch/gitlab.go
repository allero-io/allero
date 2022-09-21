package fetch

import (
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

type FetchGitlabDependencies struct {
	GitlabClient *gitlab.Client
}

var GITLAB_TOKEN string

func NewGitlabCommand(deps *FetchCommandDependencies) *cobra.Command {
	githubCmd := &cobra.Command{
		Use:   "gitlab org/repo...",
		Short: "Fetch data of Gitlab repositories",
		Long:  "Fetch data of Gitlab repositories and entire organizations",
		Args:  cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, cmdArgs []string) {
			args := make(map[string]any)
			args["Platform"] = "gitlab"
			args["Args"] = cmdArgs
			decodedToken, _ := deps.ConfigurationManager.ParseToken()
			if decodedToken != nil {
				args["User Email"] = decodedToken.Email
			}
			deps.PosthogClient.PublishEventWithArgs("data fetched", args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			GITLAB_TOKEN = deps.ConfigurationManager.GetGitlabToken()
			gitlabClient, err := gitlab.NewClient(GITLAB_TOKEN)
			if err != nil {
				return err
			}

			fetchGitlabDeps := &FetchGitlabDependencies{
				GitlabClient: gitlabClient,
			}
			return executeGitlab(fetchGitlabDeps, args)
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

func executeGitlab(deps *FetchGitlabDependencies, args []string) error {
	gitlabConnectorDeps := &gitlabConnector.GitlabConnectorDependencies{Client: deps.GitlabClient}
	gitlabConnector := gitlabConnector.New(gitlabConnectorDeps)

	reposFetchCounter, err = gitlabConnector.Get(args)
	return err
}
