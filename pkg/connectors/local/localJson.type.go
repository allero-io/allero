package localConnector

type LocalOwner struct {
	Name         string                      `json:"ownerName"`
	Type         string                      `json:"ownerType"`
	ID           int                         `json:"id"`
	Repositories map[string]*LocalRepository `json:"repositories"`
}

type LocalRepository struct {
	Name                   string                         `json:"name"`
	FullName               string                         `json:"fullName"`
	ID                     int                            `json:"id"`
	ProgrammingLanguages   []string                       `json:"programmingLanguages"`
	GithubActionsWorkflows map[string]*PipelineFile       `json:"github-actions-workflows"`
	GitlabCi               map[string]*GitlabPipelineFile `json:"gitlab-ci"`
	JfrogPipelines         map[string]*PipelineFile       `json:"jfrog-pipelines"`
}

type PipelineFile struct {
	RelativePath string      `json:"relativePath"`
	Filename     string      `json:"filename"`
	Origin       string      `json:"origin"`
	Content      interface{} `json:"content"`
}

type GitlabPipelineFile struct {
	Filename string                 `json:"filename"`
	Origin   string                 `json:"origin"`
	Content  map[string]interface{} `json:"content"`
}
