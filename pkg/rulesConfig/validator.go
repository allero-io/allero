package rulesConfig

import (
	"encoding/json"
	"fmt"
	"strings"

	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
	"github.com/allero-io/allero/pkg/jsonschemaValidator"
	"github.com/allero-io/allero/pkg/rulesConfig/defaultRules"
	"github.com/xeipuuv/gojsonschema"
)

func (rc *RulesConfig) Validate(ruleName string, rule *defaultRules.Rule, scmPlatform string) ([]*defaultRules.SchemaError, error) {
	if rule.InCodeImplementation {
		if scmPlatform == "github" {
			return rc.InCodeValidate(rule, rc.githubData, nil)
		} else if scmPlatform == "gitlab" {
			return rc.InCodeValidate(rule, nil, rc.gitlabData)
		}
	}

	return rc.JSONSchemaValidate(ruleName, rule, scmPlatform)
}

func (rc *RulesConfig) JSONSchemaValidate(ruleName string, rule *defaultRules.Rule, scmPlatform string) ([]*defaultRules.SchemaError, error) {
	if scmPlatform == "github" && rc.githubData == nil {
		return []*defaultRules.SchemaError{}, nil
	} else if scmPlatform == "gitlab" && rc.gitlabData == nil {
		return []*defaultRules.SchemaError{}, nil
	}

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

	schemaErrors := make([]*defaultRules.SchemaError, 0)
	errorByField := make(map[string]bool)
	lowestErrorLevel := 999

	for _, schemaErrorField := range rc.createUniqueErrors(schemaResult) {
		errorFields := strings.Split(schemaErrorField, ".")
		var trimedErrorField string
		if len(errorFields) > 4 {
			trimedErrorField = strings.Join(errorFields[:5], ".")
		} else {
			trimedErrorField = strings.Join(errorFields[:], ".")
		}
		if errorByField[trimedErrorField] {
			continue
		}
		errorByField[trimedErrorField] = true
		var schemaError *defaultRules.SchemaError

		if scmPlatform == "github" {
			schemaError = rc.parseSchemaFieldGithub(rc.githubData, schemaErrorField)
		} else if scmPlatform == "gitlab" {
			schemaError = rc.parseSchemaFieldGitlab(rc.gitlabData, schemaErrorField)
		}

		if schemaError.ErrorLevel < lowestErrorLevel {
			lowestErrorLevel = schemaError.ErrorLevel
			schemaErrors = []*defaultRules.SchemaError{schemaError}
		} else if schemaError.ErrorLevel == lowestErrorLevel {
			schemaErrors = append(schemaErrors, schemaError)
		}
	}

	return schemaErrors, nil
}

func (rc *RulesConfig) createUniqueErrors(schemaResult *gojsonschema.Result) []string {
	uniqueMapping := make(map[string]bool)

	for _, schemaError := range schemaResult.Errors() {
		if schemaError.Type() == "number_all_of" {
			continue
		}
		if ok := uniqueMapping[schemaError.Field()]; !ok {
			uniqueMapping[schemaError.Field()] = true
		}
	}

	uniqueErrorsField := make([]string, 0, len(uniqueMapping))
	for k := range uniqueMapping {
		uniqueErrorsField = append(uniqueErrorsField, k)
	}
	return uniqueErrorsField
}

func (rc *RulesConfig) InCodeValidate(rule *defaultRules.Rule, githubData map[string]*githubConnector.GithubOwner, gitlabData map[string]*gitlabConnector.GitlabGroup) ([]*defaultRules.SchemaError, error) {
	return defaultRules.Validate(rule, githubData, gitlabData)
}
