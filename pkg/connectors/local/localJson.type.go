package localConnector

type LocalOwner struct {
	Name         string                      `json:"ownerName"`
	Type         string                      `json:"ownerType"`
	ID           int                         `json:"id"`
	Repositories map[string]*LocalRepository `json:"repositories"`
}

type LocalRepository struct {
	Name                  string                   `json:"name"`
	FullName              string                   `json:"fullName"`
	ID                    int                      `json:"id"`
	ProgrammingLanguages  []string                 `json:"programmingLanguages"`
	LocalActionsWorkflows map[string]*PipelineFile `json:"local-actions-workflows"`
	JfrogPipelines        map[string]*PipelineFile `json:"jfrog-pipelines"`
}

type PipelineFile struct {
	RelativePath string      `json:"relativePath"`
	Filename     string      `json:"filename"`
	Origin       string      `json:"origin"`
	Content      interface{} `json:"content"`
}
