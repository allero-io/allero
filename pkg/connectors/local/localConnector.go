package localConnector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
	"github.com/allero-io/allero/pkg/fileManager"
)

type LocalConnector struct {
	runningPath      string
	absoluteRootPath string
}

func New() *LocalConnector {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return &LocalConnector{
		runningPath:      path,
		absoluteRootPath: "",
	}
}

func (lc *LocalConnector) Get(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	lc.absoluteRootPath = abs
	var localJsonObject LocalRoot
	githubJsonObject := make(map[string]*githubConnector.GithubOwner)
	err = lc.getGithub(githubJsonObject)
	if err != nil {
		return err
	}
	localJsonObject.GithubData = githubJsonObject

	gitlabJsonObject := make(map[string]*gitlabConnector.GitlabGroup)
	err = lc.getGitlab(gitlabJsonObject)
	if err != nil {
		return err
	}
	localJsonObject.GitlabData = gitlabJsonObject

	localJson, err := json.MarshalIndent(localJsonObject, "", "  ")
	if err != nil {
		return err
	}

	alleroHomedir := fileManager.GetAlleroHomedir()
	return fileManager.WriteToFile(fmt.Sprintf("%s/repo_files/local.json", alleroHomedir), localJson)
}

func (lc *LocalConnector) walkAndMatchedFiles(dir string, regex string) ([]string, error) {

	var allFiles []string
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if matched, _ := regexp.MatchString(regex, path); matched {
			relativePath := strings.TrimPrefix(path, lc.absoluteRootPath+"/")
			allFiles = append(allFiles, relativePath)
		}

		return nil
	})

	return allFiles, err
}
