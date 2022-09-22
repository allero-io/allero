package rulesConfig

import (
	"encoding/json"
	"fmt"

	"github.com/allero-io/allero/pkg/jsonschemaValidator"
	"github.com/allero-io/allero/pkg/rulesConfig/defaultRules"
	"github.com/xeipuuv/gojsonschema"
)

func (rc *RulesConfig) Validate(ruleName string, rule *defaultRules.Rule, scmPlatform string) ([]*defaultRules.SchemaError, error) {
	if rule.InCodeImplementation {
		return rc.InCodeValidate(rule)
	}

	return rc.JSONSchemaValidate(ruleName, rule, scmPlatform)
}

func (rc *RulesConfig) JSONSchemaValidate(ruleName string, rule *defaultRules.Rule, scmPlatform string) ([]*defaultRules.SchemaError, error) {
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

	for _, rawSchemaError := range schemaResult.Errors() {
		if errorByField[rawSchemaError.Field()] {
			continue
		}

		errorByField[rawSchemaError.Field()] = true
		schemaError := rc.parseSchemaField(rc.githubData, rawSchemaError.Field(), scmPlatform)
		if schemaError.ErrorLevel < lowestErrorLevel {
			lowestErrorLevel = schemaError.ErrorLevel
			schemaErrors = []*defaultRules.SchemaError{schemaError}
		} else if schemaError.ErrorLevel == lowestErrorLevel {
			schemaErrors = append(schemaErrors, schemaError)
		}
	}

	return schemaErrors, nil
}

func (rc *RulesConfig) InCodeValidate(rule *defaultRules.Rule) ([]*defaultRules.SchemaError, error) {
	return defaultRules.Validate(rule, rc.githubData)
}
