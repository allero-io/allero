package githubConnector

import "github.com/allero-io/allero/pkg/connectors"

type GithubOwner struct {
	Name         string                       `json:"ownerName"`
	Type         string                       `json:"ownerType"`
	ID           int                          `json:"id"`
	Repositories map[string]*GithubRepository `json:"repositories"`
}

type GithubRepository struct {
	Name                   string                              `json:"name"`
	FullName               string                              `json:"fullName"`
	ID                     int                                 `json:"id"`
	ProgrammingLanguages   []string                            `json:"programmingLanguages"`
	GithubActionsWorkflows map[string]*connectors.PipelineFile `json:"github-actions-workflows"`
	JfrogPipelines         map[string]*connectors.PipelineFile `json:"jfrog-pipelines"`
}
