package cmd

import (
	"os"

	"github.com/allero-io/allero/cmd/config"
	"github.com/allero-io/allero/cmd/fetch"
	"github.com/allero-io/allero/cmd/validate"
	"github.com/allero-io/allero/cmd/version"
	"github.com/allero-io/allero/pkg/configurationManager"
	"github.com/allero-io/allero/pkg/posthog"
	"github.com/allero-io/allero/pkg/rulesConfig"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "allero",
	Short: "Protecting your production pipelines",
	Long:  `Protecting your production pipelines`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil && err != validate.ErrViolationsFound {
		rootCmd.SilenceUsage = false
		rootCmd.SilenceErrors = false

		c := color.New(color.FgHiRed)
		c.Println("Error:", err)
	}
	if err != nil {
		os.Exit(1)
	}
}

var CliVersion string

func init() {
	configurationManager := configurationManager.New()

	posthogClient, _ := posthog.New(&posthog.PosthogClientDependencies{
		ConfigurationManager: configurationManager,
		CliVersion:           CliVersion,
	})

	rulesConfig := rulesConfig.New(&rulesConfig.RulesConfigDependencies{
		ConfigurationManager: configurationManager,
	})

	rootCmd.AddCommand(fetch.New(&fetch.FetchCommandDependencies{
		ConfigurationManager: configurationManager,
		PosthogClient:        posthogClient,
	}))

	rootCmd.AddCommand(validate.New(&validate.ValidateCommandDependencies{
		ConfigurationManager: configurationManager,
		PosthogClient:        posthogClient,
		RulesConfig:          rulesConfig,
	}))

	rootCmd.AddCommand(version.New(&version.VersionCommandDependencies{
		CliVersion: CliVersion,
	}))

	rootCmd.AddCommand(config.New(&config.ConfigCommandDependencies{
		ConfigurationManager: configurationManager,
	}))
}
