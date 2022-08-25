package posthog

import (
	"fmt"

	"github.com/posthog/posthog-go"
)

func (pc *PosthogClient) PublishCmdUse(command string, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	defer client.Close()

	err = pc.postCmd(client, command, args)
	if err != nil {
		return fmt.Errorf("failed to post command: %s", command)
	}
	return nil
}

func (pc *PosthogClient) postCmd(client posthog.Client, command string, args []string) error {
	err := client.Enqueue(posthog.Capture{
		DistinctId: pc.userConfig.MachineId,
		Event:      command,
		Properties: posthog.NewProperties().
			Set("Version", pc.cliVersion).Set("Args", args),
	})
	if err != nil {
		return err
	}
	return nil
}
