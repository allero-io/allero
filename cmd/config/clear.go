package config

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewClearCommand(deps *ConfigCommandDependencies) *cobra.Command {
	setCommand := &cobra.Command{
		Use:   "clear",
		Short: "Clear configuration value",
		Long:  `Remove value for specific key in allero config.json file. Defaults to $HOME/.allero/config.json`,
		Run: func(cmd *cobra.Command, args []string) {
			err := clear(deps, args[0])
			if err != nil {
				fmt.Printf("Failed setting %s with value %s. Error: %s", args[0], args[1], err)
			}
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires exactly 1 argument")
			}

			return validateKey(args[0])
		},
	}

	return setCommand
}

func clear(deps *ConfigCommandDependencies, key string) error {
	return deps.ConfigurationManager.Clear(key)
}
