package githubConnector

type GithubOwner struct {
	Name         string                       `json:"ownerName"`
	Type         string                       `json:"ownerType"`
	ID           int                          `json:"id"`
	Repositories map[string]*GithubRepository `json:"repositories"`
}

type GithubRepository struct {
	Name                   string                     `json:"name"`
	FullName               string                     `json:"fullName"`
	ID                     int                        `json:"id"`
	GithubActionsWorkflows map[string]*GithubWorkflow `json:"github-actions-workflows"`
}

type GithubWorkflow struct {
	RelativePath string      `json:"relativePath"`
	LocalPath    string      `json:"localPath"`
	Filename     string      `json:"filename"`
	Content      interface{} `json:"content"`
}
