package localConnector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	"github.com/allero-io/allero/pkg/fileManager"
)

type LocalConnector struct {
	RootPath string
}

func New() *LocalConnector {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return &LocalConnector{
		RootPath: path,
	}
}

func (lc *LocalConnector) Get() error {
	var localJsonObject LocalRoot
	githubJsonObject := make(map[string]*githubConnector.GithubOwner)
	err := lc.getGithub(githubJsonObject)
	if err != nil {
		return err
	}
	localJsonObject.GithubData = githubJsonObject

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
			relativePath := strings.TrimPrefix(path, lc.RootPath)
			allFiles = append(allFiles, relativePath)
		}

		return nil
	})

	return allFiles, err
}
