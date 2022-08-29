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
	HttpClient           *httpClient.HttpClient
	ConfigurationManager *configurationManager.ConfigurationManager
}

func New(deps *AlleroBackendClientDeps) (*AlleroBackendClient, error) {
	return &AlleroBackendClient{
		HttpClient:           deps.HttpClient,
		ConfigurationManager: deps.ConfigurationManager,
	}, nil
}

func (c *AlleroBackendClient) GetAlleroToken() (string, error) {
	userConfig, _, err := c.ConfigurationManager.GetUserConfig()
	if err != nil {
		return "", err
	}

	if userConfig.AlleroToken == "" {
		// TODO OY replace
		// respBody, err := c.HttpClient.Get("token")
		respBody, err := c.HttpClient.Get("")
		if err != nil {
			return "", err
		}
		userConfig.AlleroToken = string(respBody)
		err = c.ConfigurationManager.UpdateUserConfig(userConfig)

		if err != nil {
			return "", err
		}
	}
	return userConfig.AlleroToken, nil
}
