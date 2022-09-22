package defaultRules

import (
	"fmt"
	"regexp"

	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
)

type SchemaError struct {
	OwnerName       string
	RepositryName   string
	WorkflowRelPath string
	CiCdPlatform    string
	ErrorLevel      int
	ScmPlatform     string
}

type Rule struct {
	Description          string                 `json:"description"`
	UniqueId             int                    `json:"uniqueId" validate:"required"`
	Schema               map[string]interface{} `json:"schema"`
	FailureMessage       string                 `json:"failureMessage" validate:"required"`
	EnabledByDefault     bool                   `json:"enabledByDefault"`
	InCodeImplementation bool                   `json:"inCodeImplementation"`
}

type Workflow struct {
	Jobs map[string]Job `json:"jobs"`
}

type Job struct {
	Steps []Step `json:"steps"`
}

type Step struct {
	Uses string `json:"uses"`
	Run  string `json:"run"`
}

func Validate(rule *Rule, githubData map[string]*githubConnector.GithubOwner) ([]*SchemaError, error) {
	if rule.UniqueId == 10 {
		return EnsureScaScanner(githubData)
	}
	if rule.UniqueId == 11 {
		return EnsureTerraformScanner(githubData)
	}

	return nil, fmt.Errorf("missing implementation for rule %d", rule.UniqueId)
}

func matchRegex(regexExpression string, str string) bool {
	r, err := regexp.Compile(regexExpression)
	if err != nil {
		return false
	}
	return r.MatchString(str)
}
