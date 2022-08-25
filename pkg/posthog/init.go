package posthog

import (
	"fmt"
	"os"
	"runtime"

	"github.com/allero-io/allero/pkg/configurationManager"
	"github.com/posthog/posthog-go"
)

type PosthogClientDependencies struct {
	ConfigurationManager *configurationManager.ConfigurationManager
	CliVersion           string
}
type PosthogClient struct {
	cliVersion string
	userConfig *configurationManager.UserConfig
}

func New(deps *PosthogClientDependencies) (*PosthogClient, error) {
	userConfig, newUser, err := deps.ConfigurationManager.GetUserConfig()
	if err != nil {
		return nil, err
	}
	if newUser {
		// A new user
		osName := runtime.GOOS
		osArch := runtime.GOARCH
		osHost, _ := os.Hostname()

		client, err := getClient()
		if err != nil {
			return nil, err
		}
		defer client.Close()

		client.Enqueue(posthog.Identify{
			DistinctId: userConfig.MachineId,
			Properties: posthog.NewProperties().
				Set("Os Name", osName).
				Set("Os Architecture", osArch).
				Set("Os Host", osHost),
		})
	}

	return &PosthogClient{cliVersion: deps.CliVersion, userConfig: userConfig}, nil
}

func getClient() (posthog.Client, error) {
	if posthogProjectToken == "" {
		return nil, fmt.Errorf("analytics token is not set")
	}
	return posthog.NewWithConfig(
		posthogProjectToken,
		posthog.Config{
			Endpoint: posthogProjectHost,
		},
	)
}
