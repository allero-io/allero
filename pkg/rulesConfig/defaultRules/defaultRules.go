package defaultRules

import (
	"fmt"
	"regexp"

	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
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

type GithubWorkflow struct {
	Jobs map[string]GithubJob `json:"jobs"`
}

type GithubJob struct {
	Steps []GithubStep `json:"steps"`
}

type GithubStep struct {
	Uses string         `json:"uses"`
	Run  string         `json:"run"`
	With map[string]any `json:"with"`
}

type GitlabStageScript struct {
	Script string `json:"script"`
}

type GitlabStageScripts struct {
	Scripts []string `json:"script"`
}

type JfrogPipelineFile struct {
	Pipelines []JfrogPipeline `json:"pipelines"`
}

type JfrogPipeline struct {
	Steps []JfrogPipelineStep `json:"steps"`
}

type JfrogPipelineStep struct {
	Execution JfrogPipelineStepExecution `json:"execution"`
}

type JfrogPipelineStepExecution struct {
	OnExecute []string `json:"onExecute"`
}

func Validate(rule *Rule, githubData map[string]*githubConnector.GithubOwner, gitlabData map[string]*gitlabConnector.GitlabGroup) ([]*SchemaError, error) {
	if rule.UniqueId == 10 {
		return EnsureScaScanner(githubData, gitlabData)
	}
	if rule.UniqueId == 11 {
		return EnsureTerraformScanner(githubData, gitlabData)
	}
	if rule.UniqueId == 14 {
		return EnsureCodeCoverageChecker(githubData, gitlabData)
	}
	if rule.UniqueId == 15 {
		return EnsureSecretsScanner(githubData, gitlabData)
	}
	if rule.UniqueId == 16 {
		return EnsureLinter(githubData, gitlabData)
	}
	if rule.UniqueId == 17 {
		return EnsureCodeQualityScanner(githubData, gitlabData)
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
