package localConnector

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"

	"github.com/allero-io/allero/pkg/connectors"
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
	"github.com/allero-io/allero/pkg/fileManager"
)

func (lc *LocalConnector) getLocalGitlab(gitlabJsonObject map[string]*gitlabConnector.GitlabGroup) error {
	err := lc.addRootPathAsNewProject(gitlabJsonObject)
	if err != nil {
		return err
	}

	escapedRepoName := connectors.EscapeJsonKey(lc.absoluteRootPath)
	err = lc.processGitlabWorkflowFiles(gitlabJsonObject, escapedRepoName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil

}

func (lc *LocalConnector) addRootPathAsNewProject(gitlabJsonObject map[string]*gitlabConnector.GitlabGroup) error {
	gitlabJsonObject["local_group"] = &gitlabConnector.GitlabGroup{
		Name:     "sudo",
		ID:       0,
		Projects: make(map[string]*gitlabConnector.GitlabProject),
	}

	escapedRepoName := connectors.EscapeJsonKey(lc.absoluteRootPath)

	gitlabJsonObject["local_group"].Projects[escapedRepoName] = &gitlabConnector.GitlabProject{
		Name:           escapedRepoName,
		FullName:       escapedRepoName,
		ID:             0,
		GitlabCi:       make(map[string]*gitlabConnector.PipelineFile),
		JfrogPipelines: make(map[string]*gitlabConnector.PipelineFile),
	}

	return nil
}

func (lc *LocalConnector) processGitlabWorkflowFiles(gitlabJsonObject map[string]*gitlabConnector.GitlabGroup, repoName string) error {
	workflowFilesChan, _ := lc.getGitlabWorkflowFilesEntities(repoName)
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

		if workflowFile.Origin == "gitlab_ci" {
			gitlabJsonObject["local_group"].Projects[repoName].GitlabCi[escapedFilename] = workflowFile
		} else if workflowFile.Origin == "jfrog_pipelines" {
			gitlabJsonObject["local_group"].Projects[repoName].JfrogPipelines[escapedFilename] = workflowFile
		} else {
			processingError = fmt.Errorf("unsupported CICD platform %s for file %s from repository %s", workflowFile.Origin, workflowFile.RelativePath, repoName)
			continue
		}
	}

	return processingError
}

func (lc *LocalConnector) getGitlabWorkflowFilesEntities(repoName string) (chan *gitlabConnector.PipelineFile, error) {
	workflowFilesEntitiesChan := make(chan *gitlabConnector.PipelineFile)

	var getEntitiesErr error
	go func() {
		defer close(workflowFilesEntitiesChan)

		for _, cicdPlatform := range connectors.SUPPORTED_CICD_PLATFORMS {
			if !cicdPlatform.GitlabValid {
				continue
			}
			relevantFilesPaths, err := lc.walkAndMatchedFiles(lc.absoluteRootPath, cicdPlatform.RelevantFilesRegex)
			if err != nil {
				return
			}
			for _, filePath := range relevantFilesPaths {
				workflowFilesEntitiesChan <- &gitlabConnector.PipelineFile{
					RelativePath: filePath,
					Filename:     path.Base(filePath),
					Origin:       cicdPlatform.Name,
				}
			}
		}
	}()

	return workflowFilesEntitiesChan, getEntitiesErr
}
