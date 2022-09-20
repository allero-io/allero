package configurationManager

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/allero-io/allero/pkg/fileManager"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type DecodedToken struct {
	Rules    []bool `json:"rules"`
	Email    string `json:"email"`
	UniqueId string `json:"uniqueId"`
}

type ConfigurationManager struct {
	TokenGenerationUrl string
}

func New() *ConfigurationManager {
	return &ConfigurationManager{
		TokenGenerationUrl: "https://allero-mvp.webflow.io/selective-rules",
	}
}

func (cm *ConfigurationManager) initConfigFile() error {
	configHome, configName, configType, err := setViperConfig()
	if err != nil {
		return err
	}
	// workaround for creating config file when not exist
	// open issue in viper: https://github.com/spf13/viper/issues/430
	// should be fixed in pr https://github.com/spf13/viper/pull/936
	configPath := filepath.Join(configHome, configName+"."+configType)

	isDirExists, err := fileManager.IsExists(configHome)
	if err != nil {
		return err
	}
	if !isDirExists {
		osMkdirErr := os.Mkdir(configHome, os.ModePerm)
		if osMkdirErr != nil {
			return osMkdirErr
		}
	}

	isConfigExists, err := fileManager.IsExists(configPath)
	if err != nil {
		return err
	}
	if !isConfigExists {
		_, osCreateErr := os.Create(configPath)
		if osCreateErr != nil {
			return osCreateErr
		}
	}

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}

	return nil
}

func setViperConfig() (string, string, string, error) {
	configHome := fileManager.GetAlleroHomedir()

	configName := "config"
	configType := "json"

	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AddConfigPath(configHome)

	return configHome, configName, configType, nil
}

func (cm *ConfigurationManager) GetUserConfig() (*UserConfig, bool, error) {
	alleroUserConfig := fmt.Sprintf("%s/config.json", fileManager.GetAlleroHomedir())
	userConfigInstance := &UserConfig{}
	content, err := fileManager.ReadFile(alleroUserConfig)
	if err == nil {
		// An existing user
		isNewUser := false
		err = json.Unmarshal(content, &userConfigInstance)
		if err != nil {
			return nil, isNewUser, err
		}
		if userConfigInstance.MachineId == "" {
			isNewUser = true
			userConfigInstance.MachineId = calcMachineId()
			err = cm.UpdateUserConfig(userConfigInstance)
			if err != nil {
				return nil, isNewUser, err
			}
		}
		return userConfigInstance, isNewUser, nil
	}
	if os.IsNotExist(err) {
		// A new user
		isNewUser := true
		userConfigInstance.MachineId = calcMachineId()
		userConfigInstance.AlleroToken = ""
		err = cm.UpdateUserConfig(userConfigInstance)
		if err != nil {
			return nil, isNewUser, err
		}

		return userConfigInstance, isNewUser, nil
	}
	return nil, false, err
}

func (cm *ConfigurationManager) UpdateUserConfig(userConfig *UserConfig) error {
	alleroUserConfig := fmt.Sprintf("%s/config.json", fileManager.GetAlleroHomedir())
	jsonContentBytes, _ := json.MarshalIndent(userConfig, "", "  ")
	return fileManager.WriteToFile(alleroUserConfig, jsonContentBytes)
}

func (cm *ConfigurationManager) GetGithubToken() string {
	githubToken, ok := os.LookupEnv("ALLERO_GITHUB_TOKEN")
	if !ok {
		githubToken = os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			fmt.Println("We recommend providing github PAT token through environment variable ALLERO_GITHUB_TOKEN or GITHUB_TOKEN to avoid rate limit")
		}
	}
	return githubToken
}

func (cm *ConfigurationManager) GetGitlabToken() string {
	githubToken, ok := os.LookupEnv("ALLERO_GITLAB_TOKEN")
	if !ok {
		githubToken = os.Getenv("GITLAB_TOKEN")
		if githubToken == "" {
			fmt.Println("We recommend providing gitlab PAT token through environment variable ALLERO_GITLAB_TOKEN or GITLAB_TOKEN to avoid rate limit")
		}
	}
	return githubToken
}

func (cm *ConfigurationManager) SyncRules(defaultRulesList map[string][]byte) error {
	alleroRulesDir := fmt.Sprintf("%s/rules/github", fileManager.GetAlleroHomedir())

	for filename, content := range defaultRulesList {
		fileManager.WriteToFile(fmt.Sprintf("%s/%s", alleroRulesDir, filename), []byte(content))
	}

	return nil
}

func (cm *ConfigurationManager) Set(key string, value string) error {
	initConfigFileErr := cm.initConfigFile()
	if initConfigFileErr != nil {
		return initConfigFileErr
	}

	viper.Set(key, value)
	writeClientIdErr := viper.WriteConfig()
	if writeClientIdErr != nil {
		return writeClientIdErr
	}
	return nil
}

func (cm *ConfigurationManager) Get(key string) (interface{}, error) {
	err := cm.initConfigFile()
	if err != nil {
		return nil, err
	}

	return viper.Get(key), nil
}

func (cm *ConfigurationManager) Clear(key string) error {
	initConfigFileErr := cm.initConfigFile()
	if initConfigFileErr != nil {
		return initConfigFileErr
	}

	fullConfig := viper.AllSettings()
	delete(fullConfig, key)
	viper.Reset()
	setViperConfig()
	for k, v := range fullConfig {
		viper.Set(k, v)
	}

	writeClientIdErr := viper.WriteConfig()
	if writeClientIdErr != nil {
		return writeClientIdErr
	}
	return nil
}

func calcMachineId() string {
	var userMachineHashKey string
	_, flag1 := os.LookupEnv("GITHUB_ACTIONS")
	githubRepository, flag2 := os.LookupEnv("GITHUB_REPOSITORY")
	githubWorkflow, flag3 := os.LookupEnv("GITHUB_WORKFLOW")
	if flag1 && flag2 && flag3 {
		userMachineHashKey = "github_actions-" + githubRepository + "-" + githubWorkflow
	} else {
		userMachineHashKey = uuid.New().String()
	}
	keyBytes := []byte(userMachineHashKey)
	hasher := sha1.New()
	hasher.Write(keyBytes)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

func (cm *ConfigurationManager) ParseToken() (*DecodedToken, error) {
	token, err := cm.Get("token")
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, nil
	}

	rawDecodedToken, err := base64.StdEncoding.DecodeString(fmt.Sprintf("%v", token))
	if err != nil {
		return nil, fmt.Errorf(
			"error decoding token. run `allero config clear token` to clear the existing token and generate a new token using %s", cm.TokenGenerationUrl)
	}

	decodedToken := &DecodedToken{}
	err = json.Unmarshal(rawDecodedToken, decodedToken)
	if err != nil {
		return nil, err
	}
	return decodedToken, nil
}
