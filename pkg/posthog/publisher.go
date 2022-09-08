package posthog

import (
	"fmt"

	"github.com/posthog/posthog-go"
)

func (pc *PosthogClient) PublishCmdUse(command string, cmdArgs []string) error {
	args := make(map[string]any)
	args["Args"] = cmdArgs
	return pc.PublishEventWithArgs(command, args)
}

func (pc *PosthogClient) PublishEventWithArgs(command string, args map[string]interface{}) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	defer client.Close()

	err = pc.postEvent(client, command, args)
	if err != nil {
		return fmt.Errorf("failed to post command: %s", command)
	}
	return nil
}

func (pc *PosthogClient) postEvent(client posthog.Client, command string, args map[string]interface{}) error {
	properties := posthog.NewProperties().Set("Version", pc.cliVersion)
	for key, value := range args {
		properties.Set(key, value)
	}
	err := client.Enqueue(posthog.Capture{
		DistinctId: pc.userConfig.MachineId,
		Event:      command,
		Properties: properties,
	})
	if err != nil {
		return err
	}
	return nil
}
