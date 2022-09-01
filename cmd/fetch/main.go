package fetch

import (
	"github.com/allero-io/allero/pkg/alleroBackendClient"
	"github.com/allero-io/allero/pkg/configurationManager"
	"github.com/allero-io/allero/pkg/posthog"
	"github.com/spf13/cobra"
)

type FetchCommandDependencies struct {
	ConfigurationManager *configurationManager.ConfigurationManager
	PosthogClient        *posthog.PosthogClient
	AlleroBackendClient  *alleroBackendClient.AlleroBackendClient
}

var fetchCmd = &cobra.Command{
	Use:     "fetch",
	Short:   "Fetch data of repositories",
	Long:    "Fetch data of repositories and entire organizations from a specified SCM platform",
	Example: `allero fetch github allero-io dapr/dapr`,
}

func New(deps *FetchCommandDependencies) *cobra.Command {
	fetchCmd.AddCommand(NewGithubCommand(deps))

	return fetchCmd
}
