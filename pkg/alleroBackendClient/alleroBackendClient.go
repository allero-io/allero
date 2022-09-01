package alleroBackendClient

import (
	"github.com/allero-io/allero/pkg/configurationManager"
	"github.com/allero-io/allero/pkg/httpClient"
)

type AlleroBackendClientDeps struct {
	HttpClient           *httpClient.HttpClient
	ConfigurationManager *configurationManager.ConfigurationManager
}

type AlleroBackendClient struct {
	httpClient           *httpClient.HttpClient
	configurationManager *configurationManager.ConfigurationManager
	AlleroToken          string
}

func New(deps *AlleroBackendClientDeps) (*AlleroBackendClient, error) {
	alleroToken, err := getAlleroToken(deps.HttpClient, deps.ConfigurationManager)
	if err != nil {
		return nil, err
	}
	return &AlleroBackendClient{
		httpClient:           deps.HttpClient,
		configurationManager: deps.ConfigurationManager,
		AlleroToken:          alleroToken,
	}, nil
}

func getAlleroToken(hc *httpClient.HttpClient, cm *configurationManager.ConfigurationManager) (string, error) {
	userConfig, _, err := cm.GetUserConfig()
	if err != nil {
		return "", err
	}

	if userConfig.AlleroToken == "" {
		// TODO OY replace
		// respBody, err := c.HttpClient.Get("token")
		respBody, err := hc.Get("")
		if err != nil {
			return "", err
		}
		userConfig.AlleroToken = string(respBody)
		err = cm.UpdateUserConfig(userConfig)

		if err != nil {
			return "", err
		}
	}
	return userConfig.AlleroToken, nil
}
