package rulesConfig

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/allero-io/allero/pkg/configurationManager"
	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
	localConnector "github.com/allero-io/allero/pkg/connectors/local"
	"github.com/allero-io/allero/pkg/fileManager"
	"github.com/allero-io/allero/pkg/rulesConfig/defaultRules"
	"github.com/go-playground/validator"
)

//go:embed defaultRules/github/*
var githubRulesList embed.FS

//go:embed defaultRules/gitlab/*
var gitlabRulesList embed.FS

type RulesConfig struct {
	configurationManager *configurationManager.ConfigurationManager
	githubData           map[string]*githubConnector.GithubOwner
	gitlabData           map[string]*gitlabConnector.GitlabGroup
}

type RulesConfigDependencies struct {
	ConfigurationManager *configurationManager.ConfigurationManager
}

type RuleResult struct {
	RuleName       string
	Valid          bool
	SchemaErrors   []*defaultRules.SchemaError
	FailureMessage string
}
type OutputSummary struct {
	TotalOwners         int    `mapstructure:"Total Owners"`
	TotalRepositories   int    `mapstructure:"Total Repositories"`
	TotalPipelines      int    `mapstructure:"Total Pipelines"`
	TotalRulesEvaluated int    `mapstructure:"Total Rules Evaluated"`
	TotalFailedRules    int    `mapstructure:"Total Failed Rules"`
	URL                 string `mapstructure:"URL"`
}

var ErrNoFetchedData = errors.New("missing repository data. Use PATH option to validate local directory or fetch data from remote first. Run 'allero fetch -h' for more information about remote data")

func New(deps *RulesConfigDependencies) *RulesConfig {
	return &RulesConfig{
		configurationManager: deps.ConfigurationManager,
		githubData:           getGithubData(),
		gitlabData:           getGitlabData(),
	}
}

func (rc *RulesConfig) Initialize() error {
	if rc.githubData == nil && rc.gitlabData == nil {
		return ErrNoFetchedData
	}

	githubFiles, err := rc.GetRulesFiles("github", githubRulesList)
	if err != nil {
		return err
	}

	err = rc.configurationManager.SyncRules(githubFiles, "github")
	if err != nil {
		return err
	}

	gitlabFiles, err := rc.GetRulesFiles("gitlab", gitlabRulesList)
	if err != nil {
		return err
	}

	err = rc.configurationManager.SyncRules(gitlabFiles, "gitlab")
	if err != nil {
		return err
	}

	return nil
}

func (rc *RulesConfig) GetRulesFiles(folderName string, rulesList embed.FS) (map[string][]byte, error) {
	files := make(map[string][]byte)
	rulesInDir, err := rulesList.ReadDir(fmt.Sprintf("defaultRules/%s", folderName))
	if err != nil {
		return nil, err
	}

	for _, file := range rulesInDir {
		ruleFilename := fmt.Sprintf("defaultRules/%s/%s", folderName, file.Name())
		content, err := rulesList.ReadFile(ruleFilename)
		if err != nil {
			return nil, err
		}

		files[file.Name()] = content
	}

	return files, nil
}

func (rc *RulesConfig) GetSummary() OutputSummary {
	totalOwners := len(rc.githubData) + len(rc.gitlabData)
	totalRepositories := 0
	totalPipelines := 0

	for _, owner := range rc.githubData {
		totalRepositories += len(owner.Repositories)

		for _, repository := range owner.Repositories {
			totalPipelines += len(repository.GithubActionsWorkflows)
			totalPipelines += len(repository.JfrogPipelines)
		}
	}

	for _, group := range rc.gitlabData {
		totalRepositories += len(group.Projects)

		for _, project := range group.Projects {
			totalPipelines += len(project.GitlabCi)
			totalPipelines += len(project.JfrogPipelines)
		}
	}

	return OutputSummary{
		TotalOwners:       totalOwners,
		TotalRepositories: totalRepositories,
		TotalPipelines:    totalPipelines,
	}
}

func (rc *RulesConfig) parseSchemaFieldGithub(githubData map[string]*githubConnector.GithubOwner, field string) *defaultRules.SchemaError {
	keyFields := strings.Split(field, ".")
	schemaError := &defaultRules.SchemaError{
		ScmPlatform: "github",
	}
	errorLevel := 0

	if len(keyFields) >= 1 {
		schemaError.OwnerName = keyFields[0]
		errorLevel = 1
	}
	if len(keyFields) >= 3 {
		schemaError.RepositryName = keyFields[2]
		errorLevel = 2
	}
	if len(keyFields) >= 4 {
		schemaError.CiCdPlatform = keyFields[3]
		errorLevel = 3
	}
	if len(keyFields) >= 5 {
		workflowName := keyFields[4]

		if schemaError.CiCdPlatform == "github-actions-workflows" {
			schemaError.WorkflowRelPath = githubData[schemaError.OwnerName].Repositories[schemaError.RepositryName].GithubActionsWorkflows[workflowName].RelativePath
		}
		if schemaError.CiCdPlatform == "jfrog-pipelines" {
			schemaError.WorkflowRelPath = githubData[schemaError.OwnerName].Repositories[schemaError.RepositryName].JfrogPipelines[workflowName].RelativePath
		}
		errorLevel = 4
	}

	schemaError.ErrorLevel = errorLevel
	return schemaError
}

func (rc *RulesConfig) parseSchemaFieldGitlab(gitlabData map[string]*gitlabConnector.GitlabGroup, field string) *defaultRules.SchemaError {
	keyFields := strings.Split(field, ".")
	schemaError := &defaultRules.SchemaError{
		ScmPlatform: "github",
	}
	errorLevel := 0

	if len(keyFields) >= 1 {
		schemaError.OwnerName = keyFields[0]
		errorLevel = 1
	}
	if len(keyFields) >= 3 {
		schemaError.RepositryName = keyFields[2]
		errorLevel = 2
	}
	if len(keyFields) >= 4 {
		schemaError.CiCdPlatform = keyFields[3]
		errorLevel = 3
	}
	if len(keyFields) >= 5 {
		workflowName := keyFields[4]

		if schemaError.CiCdPlatform == "gitlab-ci" {
			schemaError.WorkflowRelPath = gitlabData[schemaError.OwnerName].Projects[schemaError.RepositryName].GitlabCi[workflowName].RelativePath
		}
		if schemaError.CiCdPlatform == "jfrog-pipelines" {
			schemaError.WorkflowRelPath = gitlabData[schemaError.OwnerName].Projects[schemaError.RepositryName].JfrogPipelines[workflowName].RelativePath
		}
		errorLevel = 4
	}

	schemaError.ErrorLevel = errorLevel
	return schemaError
}

func (rc *RulesConfig) GetAllRuleNames(scmPlatform string) []string {
	alleroHomedir := fileManager.GetAlleroHomedir()
	rulesPath := fmt.Sprintf("%s/rules/%s", alleroHomedir, scmPlatform)
	customRulesPath := fmt.Sprintf("%s/rules/%s/custom", alleroHomedir, scmPlatform)

	ruleNames := []string{}

	files := fileManager.ReadFolder(rulesPath)
	customFiles := fileManager.ReadFolder(customRulesPath)

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			ruleNames = append(ruleNames, strings.TrimSuffix(file.Name(), ".json"))
		}
	}

	for _, file := range customFiles {
		if strings.HasSuffix(file.Name(), ".json") {
			ruleName := "custom/" + strings.TrimSuffix(file.Name(), ".json")
			ruleNames = append(ruleNames, ruleName)
		}
	}

	return ruleNames
}

func (rc *RulesConfig) GetSelectedRuleIds() (map[int]bool, error) {
	decodedToken, err := rc.configurationManager.ParseToken()
	if err != nil {
		return nil, err
	}

	if decodedToken == nil {
		return nil, nil
	}

	selectedRuleIds := make(map[int]bool)
	for i, rule := range decodedToken.Rules {
		if rule {
			selectedRuleIds[i+1] = true
		}
	}

	return selectedRuleIds, nil
}

func (rc *RulesConfig) GetRule(ruleName string, scmPlatform string) (*defaultRules.Rule, error) {
	isCustomRule := strings.HasPrefix(ruleName, "custom/")

	alleroHomedir := fileManager.GetAlleroHomedir()
	ruleFilename := fmt.Sprintf("%s/rules/%s/%s.json", alleroHomedir, scmPlatform, ruleName)

	content, err := os.ReadFile(ruleFilename)
	if err != nil {
		return nil, err
	}

	rule := &defaultRules.Rule{}
	err = json.Unmarshal(content, rule)
	if err != nil {
		return nil, err
	}

	if isCustomRule {
		rule.UniqueId = rule.UniqueId + 1000
	}

	return rule, rc.validateRuleStructure(ruleName, rule)
}

func getGithubData() map[string]*githubConnector.GithubOwner {
	githubData := make(map[string]*githubConnector.GithubOwner)
	alleroHomedir := fileManager.GetAlleroHomedir()
	githubDataFilename := fmt.Sprintf("%s/repo_files/github.json", alleroHomedir)

	content, err := os.ReadFile(githubDataFilename)
	if err != nil {
		return nil
	}

	json.Unmarshal(content, &githubData)
	return githubData
}

func (rc *RulesConfig) ReadLocalData() error {
	var localData localConnector.LocalRoot
	alleroHomedir := fileManager.GetAlleroHomedir()
	localDataFilename := fmt.Sprintf("%s/repo_files/local.json", alleroHomedir)

	content, err := os.ReadFile(localDataFilename)
	if err != nil {
		return nil
	}

	json.Unmarshal(content, &localData)
	rc.githubData = localData.GithubData
	rc.gitlabData = localData.GitlabData
	return nil
}

func getGitlabData() map[string]*gitlabConnector.GitlabGroup {
	gitlabData := make(map[string]*gitlabConnector.GitlabGroup)
	alleroHomedir := fileManager.GetAlleroHomedir()
	gitlabDataFilename := fmt.Sprintf("%s/repo_files/gitlab.json", alleroHomedir)

	content, err := os.ReadFile(gitlabDataFilename)
	if err != nil {
		return nil
	}

	json.Unmarshal(content, &gitlabData)
	return gitlabData
}

func (rc *RulesConfig) validateRuleStructure(ruleName string, rule *defaultRules.Rule) error {
	validate := validator.New()
	err := validate.Struct(rule)
	if err != nil {
		err = fmt.Errorf("rule %s is invalid: %s", ruleName, err.Error())
	}

	return err
}
