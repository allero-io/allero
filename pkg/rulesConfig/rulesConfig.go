package rulesConfig

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/allero-io/allero/pkg/configurationManager"
	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	"github.com/allero-io/allero/pkg/fileManager"
	"github.com/allero-io/allero/pkg/jsonschemaValidator"
	"github.com/go-playground/validator"
)

//go:embed defaultRules/github/*
var githubRulesList embed.FS

type Rule struct {
	Description    string                 `json:"description"`
	UniqueId       int                    `json:"uniqueId"`
	Schema         map[string]interface{} `json:"schema" validate:"required"`
	FailureMessage string                 `json:"failureMessage" validate:"required"`
}

type RulesConfig struct {
	configurationManager *configurationManager.ConfigurationManager
	githubData           map[string]*githubConnector.GithubOwner
}

type RulesConfigDependencies struct {
	ConfigurationManager *configurationManager.ConfigurationManager
}

type SchemaError struct {
	OwnerName       string
	RepositryName   string
	WorkflowRelPath string
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

type DecodedToken struct {
	Rules    []bool `json:"rules"`
	Email    string `json:"email"`
	UniqueId string `json:"uniqueId"`
}

func New(deps *RulesConfigDependencies) *RulesConfig {
	return &RulesConfig{
		configurationManager: deps.ConfigurationManager,
		githubData:           getGithubData(),
	}
}

func (rc *RulesConfig) Initialize() error {
	if rc.githubData == nil {
		return fmt.Errorf("missing repository data. Run 'allero fetch github <org | repository | username>'")
	}

	githubRulesInDir, _ := githubRulesList.ReadDir("defaultRules/github")
	files := make(map[string][]byte)

	for _, file := range githubRulesInDir {
		ruleFilename := fmt.Sprintf("defaultRules/github/%s", file.Name())
		content, err := githubRulesList.ReadFile(ruleFilename)
		if err != nil {
			return err
		}

		files[file.Name()] = content
	}

	return rc.configurationManager.SyncRules(files)
}

func (rc *RulesConfig) Validate(ruleName string, rule *Rule) ([]*SchemaError, error) {
	ruleSchema, err := json.Marshal(rule.Schema)
	if err != nil {
		return nil, err
	}

	schemaResult, err := jsonschemaValidator.Validate(ruleSchema, rc.githubData)
	if err != nil {
		return nil, fmt.Errorf("error validating schema for rule %s: %s", ruleName, err)
	}

	schemaErrors := make([]*SchemaError, 0)
	errorByField := make(map[string]bool)
	for _, rawSchemaError := range schemaResult.Errors() {
		if errorByField[rawSchemaError.Field()] {
			continue
		}

		errorByField[rawSchemaError.Field()] = true
		schemaError := rc.parseSchemaField(rc.githubData, rawSchemaError.Field())
		schemaErrors = append(schemaErrors, schemaError)
	}

	return schemaErrors, nil
}

func (rc *RulesConfig) GetSummary() OutputSummary {
	totalOwners := len(rc.githubData)
	totalRepositories := 0
	totalPipelines := 0

	for _, owner := range rc.githubData {
		totalRepositories += len(owner.Repositories)

		for _, repository := range owner.Repositories {
			totalPipelines += len(repository.GithubActionsWorkflows)
		}
	}

	return OutputSummary{
		TotalOwners:       totalOwners,
		TotalRepositories: totalRepositories,
		TotalPipelines:    totalPipelines,
	}
}

func (rc *RulesConfig) parseSchemaField(githubData map[string]*githubConnector.GithubOwner, field string) *SchemaError {
	keyFields := strings.Split(field, ".")
	schemaError := &SchemaError{}

	if len(keyFields) >= 1 {
		schemaError.OwnerName = keyFields[0]
	}
	if len(keyFields) >= 3 {
		schemaError.RepositryName = keyFields[2]
	}
	if len(keyFields) >= 5 {
		workflowName := keyFields[4]
		relpath := githubData[schemaError.OwnerName].Repositories[schemaError.RepositryName].GithubActionsWorkflows[workflowName].RelativePath
		schemaError.WorkflowRelPath = relpath
	}

	return schemaError
}

func (rc *RulesConfig) GetAllRuleNames() []string {
	alleroHomedir := fileManager.GetAlleroHomedir()
	rulesPath := fmt.Sprintf("%s/rules/github", alleroHomedir)

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
	token, err := rc.configurationManager.Get("token")
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, nil
	}

	rawDecodedToken, err := base64.StdEncoding.DecodeString(fmt.Sprintf("%v", token))
	if err != nil {
		return nil, fmt.Errorf(
			"error decoding token. run `allero config clear token` to clear the existing token and generate a new token using %s", rc.configurationManager.TokenGenerationUrl)
	}

	decodedToken := &DecodedToken{}
	err = json.Unmarshal(rawDecodedToken, decodedToken)
	if err != nil {
		return nil, err
	}

	selectedRuleIds := make(map[int]bool)
	for i, rule := range decodedToken.Rules {
		if rule {
			selectedRuleIds[i+1] = true
		}
	}

	return selectedRuleIds, nil
}

func (rc *RulesConfig) GetRule(ruleName string) (*Rule, error) {
	alleroHomedir := fileManager.GetAlleroHomedir()
	ruleFilename := fmt.Sprintf("%s/rules/github/%s.json", alleroHomedir, ruleName)

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

func (rc *RulesConfig) validateRuleStructure(ruleName string, rule *Rule) error {
	validate := validator.New()
	err := validate.Struct(rule)
	if err != nil {
		err = fmt.Errorf("Rule %s is invalid: %s", ruleName, err.Error())
	}

	return err
}
