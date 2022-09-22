package rulesConfig

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/allero-io/allero/pkg/configurationManager"
	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
	"github.com/allero-io/allero/pkg/fileManager"
	"github.com/allero-io/allero/pkg/jsonschemaValidator"
	"github.com/go-playground/validator"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed defaultRules/github/*
var githubRulesList embed.FS

//go:embed defaultRules/gitlab/*
var gitlabRulesList embed.FS

type Rule struct {
	Description      string                 `json:"description"`
	UniqueId         int                    `json:"uniqueId" validate:"required"`
	Schema           map[string]interface{} `json:"schema" validate:"required"`
	FailureMessage   string                 `json:"failureMessage" validate:"required"`
	EnabledByDefault bool                   `json:"enabledByDefault"`
}

type RulesConfig struct {
	configurationManager *configurationManager.ConfigurationManager
	githubData           map[string]*githubConnector.GithubOwner
	gitlabData           map[string]*gitlabConnector.GitlabGroup
}

type RulesConfigDependencies struct {
	ConfigurationManager *configurationManager.ConfigurationManager
}

type SchemaError struct {
	OwnerName       string
	RepositryName   string
	WorkflowRelPath string
	CiCdPlatform    string
	ErrorLevel      int
	ScmPlatform     string
	ErrorLevel      int
}

type RuleResult struct {
	RuleName       string
	Valid          bool
	SchemaErrors   []*SchemaError
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

func New(deps *RulesConfigDependencies) *RulesConfig {
	return &RulesConfig{
		configurationManager: deps.ConfigurationManager,
		githubData:           getGithubData(),
		gitlabData:           getGitlabData(),
	}
}

func (rc *RulesConfig) Initialize() error {
	if rc.githubData == nil && rc.gitlabData == nil {
		return fmt.Errorf("missing repository data. Run 'allero fetch -h' for more information")
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

func (rc *RulesConfig) Validate(ruleName string, rule *Rule, scmPlatform string) ([]*SchemaError, error) {
	ruleSchema, err := json.Marshal(rule.Schema)
	if err != nil {
		return nil, err
	}

	var schemaResult *gojsonschema.Result

	if scmPlatform == "github" {
		schemaResult, err = jsonschemaValidator.Validate(ruleSchema, rc.githubData)
		if err != nil {
			return nil, fmt.Errorf("error validating schema in %s for rule %s: %s", scmPlatform, ruleName, err)
		}
	} else if scmPlatform == "gitlab" {
		schemaResult, err = jsonschemaValidator.Validate(ruleSchema, rc.gitlabData)
		if err != nil {
			return nil, fmt.Errorf("error validating schema in %s for rule %s: %s", scmPlatform, ruleName, err)
		}
	}

	schemaErrors := make([]*SchemaError, 0)
	errorByField := make(map[string]bool)
	lowestErrorLevel := 999

	for _, rawSchemaError := range schemaResult.Errors() {
		if errorByField[rawSchemaError.Field()] {
			continue
		}

		errorByField[rawSchemaError.Field()] = true
		schemaError := rc.parseSchemaField(rc.githubData, rawSchemaError.Field(), scmPlatform)
		if schemaError.ErrorLevel < lowestErrorLevel {
			lowestErrorLevel = schemaError.ErrorLevel
			schemaErrors = []*SchemaError{schemaError}
		} else if schemaError.ErrorLevel == lowestErrorLevel {
			schemaErrors = append(schemaErrors, schemaError)
		}
	}

	return schemaErrors, nil
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

func (rc *RulesConfig) parseSchemaField(githubData map[string]*githubConnector.GithubOwner, field string, scmPlatform string) *SchemaError {
	keyFields := strings.Split(field, ".")
	schemaError := &SchemaError{
		ScmPlatform: scmPlatform,
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
		if schemaError.CiCdPlatform == "gitlab-ci" {
			// TODO DB complete this condition. gitlab errors contain duplications
		}
		errorLevel = 4
	}

	schemaError.ErrorLevel = errorLevel
	return schemaError
}

func (rc *RulesConfig) GetAllRuleNames(scmPlatform string) []string {
	alleroHomedir := fileManager.GetAlleroHomedir()
	rulesPath := fmt.Sprintf("%s/rules/%s", alleroHomedir, scmPlatform)

	ruleNames := []string{}

	files := fileManager.ReadFolder(rulesPath)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			ruleNames = append(ruleNames, strings.TrimSuffix(file.Name(), ".json"))
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

func (rc *RulesConfig) GetRule(ruleName string, scmPlatform string) (*Rule, error) {
	alleroHomedir := fileManager.GetAlleroHomedir()
	ruleFilename := fmt.Sprintf("%s/rules/%s/%s.json", alleroHomedir, scmPlatform, ruleName)

	content, err := os.ReadFile(ruleFilename)
	if err != nil {
		return nil, err
	}

	rule := &Rule{}
	err = json.Unmarshal(content, rule)
	if err != nil {
		return nil, err
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

func (rc *RulesConfig) validateRuleStructure(ruleName string, rule *Rule) error {
	validate := validator.New()
	err := validate.Struct(rule)
	if err != nil {
		err = fmt.Errorf("Rule %s is invalid: %s", ruleName, err.Error())
	}

	return err
}
