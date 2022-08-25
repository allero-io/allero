package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

type VersionCommandDependencies struct {
	CliVersion string
}

func New(deps *VersionCommandDependencies) *cobra.Command {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display current version",
		Long:  "Display current version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(deps.CliVersion)
		},
	}

	return versionCmd
}
