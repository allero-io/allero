package localConnector

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"

	"github.com/allero-io/allero/pkg/connectors"
	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	"github.com/allero-io/allero/pkg/fileManager"
)

func (lc *LocalConnector) getLocalGithub(githubJsonObject map[string]*githubConnector.GithubOwner) error {
	err := lc.addRootPathAsNewRepo(githubJsonObject)
	if err != nil {
		return err
	}

	escapedRepoName := connectors.EscapeJsonKey(lc.absoluteRootPath)
	err = lc.processGithubWorkflowFiles(githubJsonObject, escapedRepoName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil

}

func (lc *LocalConnector) addRootPathAsNewRepo(githubJsonObject map[string]*githubConnector.GithubOwner) error {
	githubJsonObject["local_owner"] = &githubConnector.GithubOwner{
		Name:         "sudo",
		Type:         "local_github",
		ID:           0,
		Repositories: make(map[string]*githubConnector.GithubRepository),
	}

	escapedRepoName := connectors.EscapeJsonKey(lc.absoluteRootPath)

	githubJsonObject["local_owner"].Repositories[escapedRepoName] = &githubConnector.GithubRepository{
		Name:                   escapedRepoName,
		FullName:               escapedRepoName,
		ID:                     0,
		ProgrammingLanguages:   nil,
		GithubActionsWorkflows: make(map[string]*connectors.PipelineFile),
		JfrogPipelines:         make(map[string]*connectors.PipelineFile),
	}

	return nil
}

func (lc *LocalConnector) processGithubWorkflowFiles(githubJsonObject map[string]*githubConnector.GithubOwner, repoName string) error {
	workflowFilesChan, _ := lc.getWorkflowFilesEntities(repoName)
	var processingError error

	for workflowFile := range workflowFilesChan {
		fullPath := filepath.Join(lc.absoluteRootPath, workflowFile.RelativePath)
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
		} else {
			processingError = fmt.Errorf("unsupported CICD platform %s for file %s from repository %s", workflowFile.Origin, workflowFile.RelativePath, repoName)
			continue
		}
	}

	return processingError
}

func (lc *LocalConnector) getWorkflowFilesEntities(repoName string) (chan *connectors.PipelineFile, error) {
	workflowFilesEntitiesChan := make(chan *connectors.PipelineFile)

	var getEntitiesErr error
	go func() {
		defer close(workflowFilesEntitiesChan)

		for _, cicdPlatform := range connectors.SUPPORTED_CICD_PLATFORMS {
			if !cicdPlatform.GithubValid {
				continue
			}
			relevantFilesPaths, err := lc.walkAndMatchedFiles(lc.absoluteRootPath, cicdPlatform.RelevantFilesRegex)
			if err != nil {
				return
			}
			for _, filePath := range relevantFilesPaths {
				workflowFilesEntitiesChan <- &connectors.PipelineFile{
					RelativePath: filePath,
					Filename:     path.Base(filePath),
					Origin:       cicdPlatform.Name,
				}
			}
		}
	}()

	return workflowFilesEntitiesChan, getEntitiesErr
}
