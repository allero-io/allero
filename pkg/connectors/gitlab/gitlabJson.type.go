package gitlabConnector

import "github.com/allero-io/allero/pkg/connectors"

type GitlabGroup struct {
	Name     string                    `json:"groupName"`
	ID       int                       `json:"id"`
	Projects map[string]*GitlabProject `json:"projects"`
}

type GitlabProject struct {
	Name           string                              `json:"name"`
	FullName       string                              `json:"fullName"`
	ID             int                                 `json:"id"`
	GitlabCi       map[string]*connectors.PipelineFile `json:"gitlab-ci"`
	JfrogPipelines map[string]*connectors.PipelineFile `json:"jfrog-pipelines"`
}
