package config

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewSetCommand(deps *ConfigCommandDependencies) *cobra.Command {
	setCommand := &cobra.Command{
		Use:   "set",
		Short: "Set configuration value",
		Long:  `Apply value for specific key in allero config.json file. Defaults to $HOME/.allero/config.json`,
		Run: func(cmd *cobra.Command, args []string) {
			err := set(deps, args[0], args[1])
			if err != nil {
				fmt.Printf("Failed setting %s with value %s. Error: %s", args[0], args[1], err)
			}
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("requires exactly 2 arguments")
			}

			return validateKey(args[0])
		},
	}

	return setCommand
}

func set(deps *ConfigCommandDependencies, key string, value string) error {
	return deps.ConfigurationManager.Set(key, value)
}
