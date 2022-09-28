package connectors

import (
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"
)

type OwnerWithRepo struct {
	Repo  string
	Owner string
}

type CICDPlatform struct {
	Name               string
	RelevantFilesRegex string
	GithubValid        bool
	GitlabValid        bool
}

var SUPPORTED_CICD_PLATFORMS = []CICDPlatform{
	{
		Name:               "github_actions",
		RelevantFilesRegex: "\\.github/workflows/.*\\.ya?ml",
		GithubValid:        true,
		GitlabValid:        false,
	},
	{
		Name:               "jfrog_pipelines",
		RelevantFilesRegex: "jfrog.*\\.ya?ml",
		GithubValid:        true,
		GitlabValid:        true,
	},
	{
		Name:               "gitlab_ci",
		RelevantFilesRegex: "\\.gitlab-ci\\.ya?ml",
		GithubValid:        false,
		GitlabValid:        true,
	},
	// {
	// 	Name:               "jenkins",
	// 	RelevantFilesRegex: "(?i)jenkinsfile[^/]*$",
	// },
}

func SplitParentRepo(arg string) *OwnerWithRepo {
	splits := strings.Split(arg, "/")
	owner := splits[0]

	var repo string
	if len(splits) > 1 {
		repo = splits[1]
	}

	return &OwnerWithRepo{Owner: owner, Repo: repo}
}

func YamlToJson(byteContent []byte) ([]byte, error) {
	strContent := string(byteContent)
	modifiedStr := regexp.MustCompile(`[^$]{{.*}}`).ReplaceAllString(strContent, " DYNAMIC_VALUE")
	return yaml.YAMLToJSON([]byte(modifiedStr))
}

func EscapeJsonKey(key string) string {
	return strings.ReplaceAll(key, ".", "[ESCAPED_DOT]")
}
