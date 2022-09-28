package localConnector

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/allero-io/allero/pkg/connectors"
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

func (lc *LocalConnector) getGithub(githubJsonObject map[string]*githubConnector.GithubOwner) error {
	err := lc.addRootPathAsNewRepo(githubJsonObject)
	if err != nil {
		return err
	}

	escapedRepoName := connectors.EscapeJsonKey(lc.RootPath)
	err = lc.processWorkflowFiles(githubJsonObject, escapedRepoName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil

}

func (lc *LocalConnector) addRootPathAsNewRepo(githubJsonObject map[string]*githubConnector.GithubOwner) error {
	githubJsonObject["local_owner"] = &githubConnector.GithubOwner{
		Name:         "sudo",
		Type:         "",
		ID:           0,
		Repositories: make(map[string]*githubConnector.GithubRepository),
	}

	escapedRepoName := connectors.EscapeJsonKey(lc.RootPath)

	githubJsonObject["local_owner"].Repositories[escapedRepoName] = &githubConnector.GithubRepository{
		Name:                   escapedRepoName,
		FullName:               escapedRepoName,
		ID:                     0,
		ProgrammingLanguages:   nil,
		GithubActionsWorkflows: make(map[string]*githubConnector.PipelineFile),
		JfrogPipelines:         make(map[string]*githubConnector.PipelineFile),
	}

	return nil
}

func (lc *LocalConnector) processWorkflowFiles(githubJsonObject map[string]*githubConnector.GithubOwner, repoName string) error {
	workflowFilesChan, _ := lc.getWorkflowFilesEntities(repoName)
	var processingError error

	for workflowFile := range workflowFilesChan {
		fullPath := lc.RootPath + workflowFile.RelativePath
		content, err := fileManager.ReadFile(fullPath)
		if err != nil {
			processingError = fmt.Errorf("failed to get content for file %s", fullPath)
			continue
		}

		jsonContentBytes, err := connectors.YamlToJson(content)
		if err != nil {
			processingError = err
			continue
		}

		jsonContent := make(map[string]interface{})
		err = json.Unmarshal(jsonContentBytes, &jsonContent)
		if err != nil {
			processingError = err
			continue
		}

		workflowFile.Content = jsonContent
		escapedFilename := connectors.EscapeJsonKey(workflowFile.Filename)

		if workflowFile.Origin == "github_actions" {
			githubJsonObject["local_owner"].Repositories[repoName].GithubActionsWorkflows[escapedFilename] = workflowFile
		} else if workflowFile.Origin == "jfrog_pipelines" {
			githubJsonObject["local_owner"].Repositories[repoName].JfrogPipelines[escapedFilename] = workflowFile
		} else if workflowFile.Origin == "gitlab_ci" {
			// TODO find a way to use a channel with GilabPipeline
			// localJsonObject["local_owner"].Repositories[repoName].GitlabCi[escapedFilename] = workflowFile
		} else {
			processingError = fmt.Errorf("unsupported CICD platform %s for file %s from repository %s", workflowFile.Origin, workflowFile.RelativePath, repoName)
			continue
		}
	}

	return processingError
}

func (lc *LocalConnector) getWorkflowFilesEntities(repoName string) (chan *githubConnector.PipelineFile, error) {
	workflowFilesEntitiesChan := make(chan *githubConnector.PipelineFile)

	var getEntitiesErr error
	go func() {
		defer close(workflowFilesEntitiesChan)

		for _, cicdPlatform := range connectors.SUPPORTED_CICD_PLATFORMS {
			relevantFilesPaths, err := lc.walkAndMatchedFiles(lc.RootPath, cicdPlatform.RelevantFilesRegex)
			if err != nil {
				return
			}
			for _, filePath := range relevantFilesPaths {
				workflowFilesEntitiesChan <- &githubConnector.PipelineFile{
					RelativePath: filePath,
					Filename:     path.Base(filePath),
					Origin:       cicdPlatform.Name,
				}
			}
		}
	}()

	return workflowFilesEntitiesChan, getEntitiesErr
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
