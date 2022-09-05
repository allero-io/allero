package configurationManager

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/allero-io/allero/pkg/fileManager"
	"github.com/google/uuid"
)

type ConfigurationManager struct{}

func New() *ConfigurationManager {
	return &ConfigurationManager{}
}

func (cm *ConfigurationManager) GetUserConfig() (*UserConfig, bool, error) {
	alleroUserConfig := fmt.Sprintf("%s/config.json", fileManager.GetAlleroHomedir())
	jsonContent := &UserConfig{}
	content, err := fileManager.ReadFile(alleroUserConfig)
	if err == nil {
		err = json.Unmarshal(content, &jsonContent)
		if err != nil {
			return nil, false, err
		}
		return jsonContent, false, nil
	}
	if os.IsNotExist(err) {
		// A new user
		machineId := calcMachineId()
		jsonContent.MachineId = machineId
		jsonContent.AlleroToken = ""
		jsonContentBytes, _ := json.MarshalIndent(jsonContent, "", "  ")
		err = fileManager.WriteToFile(alleroUserConfig, jsonContentBytes)
		if err != nil {
			return nil, false, err
		}

		return jsonContent, true, err
	}
	return nil, false, err
}

func (cm *ConfigurationManager) UpdateUserConfig(userConfig *UserConfig) error {
	alleroUserConfig := fmt.Sprintf("%s/config.json", fileManager.GetAlleroHomedir())
	jsonContentBytes, _ := json.MarshalIndent(userConfig, "", "  ")
	return fileManager.WriteToFile(alleroUserConfig, jsonContentBytes)
}

func (cm *ConfigurationManager) GetGithubToken() string {
	githubToken := os.Getenv("ALLERO_GITHUB_TOKEN")
	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			fmt.Println("Recommended to provide github PAT token through environment variable ALLERO_GITHUB_TOKEN or GITHUB_TOKEN to avoid rate limit")
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

func calcMachineId() string {
	var userMachineHashKey string
	runningWithCi := os.Getenv("GITHUB_CI_CONTEXT")
	if runningWithCi == "" {
		userMachineHashKey = uuid.New().String()
	} else {
		userMachineHashKey = os.Getenv("GITHUB_REPO_CONTEXT") + "-" + os.Getenv("GITHUB_WORKFLOW_CONTEXT")
	}
	keyBytes := []byte(userMachineHashKey)
	hasher := sha1.New()
	hasher.Write(keyBytes)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}
