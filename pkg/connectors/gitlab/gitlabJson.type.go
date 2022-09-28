package gitlabConnector

type GitlabGroup struct {
	Name     string                    `json:"groupName"`
	ID       int                       `json:"id"`
	Projects map[string]*GitlabProject `json:"projects"`
}

type GitlabProject struct {
	Name           string                   `json:"name"`
	FullName       string                   `json:"fullName"`
	ID             int                      `json:"id"`
	GitlabCi       map[string]*PipelineFile `json:"gitlab-ci"`
	JfrogPipelines map[string]*PipelineFile `json:"jfrog-pipelines"`
}

type PipelineFile struct {
	RelativePath string                 `json:"relativePath"`
	Filename     string                 `json:"filename"`
	Origin       string                 `json:"origin"`
	Content      map[string]interface{} `json:"content"`
}
