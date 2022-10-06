package fetch

import (
	"strings"

	"github.com/allero-io/allero/pkg/clients"
	"github.com/allero-io/allero/pkg/configurationManager"
	"github.com/allero-io/allero/pkg/connectors"
	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
	"github.com/allero-io/allero/pkg/posthog"
	"github.com/spf13/cobra"
)

var (
	reposFetchCounter int
)

var scmPrefixes = map[string]string{
	"github": "https://github.com/",
	"gitlab": "https://gitlab.com/",
}

type FetchCommandDependencies struct {
	ConfigurationManager *configurationManager.ConfigurationManager
	PosthogClient        *posthog.PosthogClient
}

func New(deps *FetchCommandDependencies) *cobra.Command {
	fetchCmd := &cobra.Command{
		Use:     "fetch",
		Short:   "Fetch data of repositories",
		Long:    "Fetch data of repositories and entire organizations from several SCM platforms",
		Example: `allero fetch https://github.com/allero-io/allero`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeFetch(deps, args)
		},
	}

	fetchCmd.AddCommand(NewGithubCommand(deps))
	fetchCmd.AddCommand(NewGitlabCommand(deps))

	return fetchCmd
}

func executeFetch(deps *FetchCommandDependencies, args []string) error {
	scmMapping := getAllOwnersWithRepo(args)
	var err error

	for scm, ownersWithRepos := range scmMapping {
		switch scm {

		case "github":
			githubConnectorDeps := &githubConnector.GithubConnectorDependencies{Client: clients.CreateGithubClient(*deps.ConfigurationManager)}
			githubConnector := githubConnector.New(githubConnectorDeps)

			reposFetchCounter, err = githubConnector.Get(ownersWithRepos)
			if err != nil {
				return err
			}

		case "gitlab":
			gitlabClient, err := clients.CreateGitlabClient(*deps.ConfigurationManager)
			if err != nil {
				return err
			}

			gitlabConnectorDeps := &gitlabConnector.GitlabConnectorDependencies{Client: gitlabClient}
			gitlabConnector := gitlabConnector.New(gitlabConnectorDeps)

			reposFetchCounter, err = gitlabConnector.Get(ownersWithRepos)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getAllOwnersWithRepo(args []string) map[string][]*connectors.OwnerWithRepo {
	scmToArgs := getArgsPerScm(args)
	scmMapping := getReposPerScm(scmToArgs)

	return scmMapping
}

func getArgsPerScm(args []string) map[string][]string {
	scmToArgs := make(map[string][]string)
	for _, arg := range args {
		for scm, prefix := range scmPrefixes {
			if strings.HasPrefix(arg, prefix) {
				trimmedArg := strings.TrimPrefix(arg, prefix)
				scmToArgs[scm] = append(scmToArgs[scm], trimmedArg)
			}
		}
	}

	return scmToArgs
}

func getReposPerScm(scmToArgs map[string][]string) map[string][]*connectors.OwnerWithRepo {
	scmMapping := make(map[string][]*connectors.OwnerWithRepo)

	for scm, _ := range scmToArgs {
		ownersWithRepos := connectors.SplitParentRepo(scmToArgs[scm])
		scmMapping[scm] = ownersWithRepos
	}

	return scmMapping
}
