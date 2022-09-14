package config

import (
	"fmt"
	"reflect"

	"github.com/allero-io/allero/pkg/configurationManager"
	"github.com/spf13/cobra"
)

var ConfigAvailableKeys = []string{"token"}

type ConfigCommandDependencies struct {
	ConfigurationManager *configurationManager.ConfigurationManager
}

func New(deps *ConfigCommandDependencies) *cobra.Command {
	configCommand := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long:  `Internal configuration management for allero config file`,
	}

	configCommand.AddCommand(NewSetCommand(deps))
	configCommand.AddCommand(NewClearCommand(deps))

	return configCommand
}

func validateKey(key string) error {
	validKeys := make(map[string]bool)

	for _, key := range ConfigAvailableKeys {
		validKeys[key] = true
	}

	if val, ok := validKeys[key]; !ok || !val {
		return fmt.Errorf("key must be one of: %s", reflect.ValueOf(validKeys).MapKeys())
	}
	return nil
}
