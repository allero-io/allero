package gitlabConnector

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/allero-io/allero/pkg/connectors"
	"github.com/allero-io/allero/pkg/fileManager"
	"github.com/xanzy/go-gitlab"
)

type GitlabConnectorDependencies struct {
	Client *gitlab.Client
}

type GitlabConnector struct {
	client *gitlab.Client
}

type GitlabProjectApiResponse struct {
	Project *gitlab.Project
	Error   error
}

func New(deps *GitlabConnectorDependencies) *GitlabConnector {
	return &GitlabConnector{
		client: deps.Client,
	}
}

func (gc *GitlabConnector) Get(args []string) (int, error) {
	projectsChan := gc.getAllProjects(args)
	reposFetchCounter := 0
	gitlabJsonObject := make(map[string]*GitlabGroup)

	for project := range projectsChan {
		if project.Error != nil {
			return reposFetchCounter, project.Error
		}

		reposFetchCounter += 1
		err := gc.addProject(gitlabJsonObject, project.Project)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = gc.processPipelineFiles(gitlabJsonObject, project.Project)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	gitlabJson, err := json.MarshalIndent(gitlabJsonObject, "", "  ")
	if err != nil {
		return reposFetchCounter, err
	}

	alleroHomedir := fileManager.GetAlleroHomedir()
	return reposFetchCounter, fileManager.WriteToFile(fmt.Sprintf("%s/repo_files/gitlab.json", alleroHomedir), gitlabJson)
}

func (gc *GitlabConnector) processPipelineFiles(gitlabJsonObject map[string]*GitlabGroup, project *gitlab.Project) error {
	workflowFilesChan, _ := gc.getPipelineFiles(project)
	var processingError error

	for workflowFile := range workflowFilesChan {
		byteContent, _, err := gc.client.RepositoryFiles.GetRawFile(project.ID, workflowFile.Filename, &gitlab.GetRawFileOptions{})
		if err != nil {
			processingError = fmt.Errorf("failed to get content for file %s from repository %s", workflowFile.Filename, project.PathWithNamespace)
			continue
		}

		jsonContentBytes, err := connectors.YamlToJson(byteContent)
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
			gitlabJsonObject[project.Namespace.Name].Projects[project.Path].GitlabCi[escapedFilename] = workflowFile
		} else if workflowFile.Origin == "jfrog_pipelines" {
			gitlabJsonObject[project.Namespace.Name].Projects[project.Path].JfrogPipelines[escapedFilename] = workflowFile
		} else {
			processingError = fmt.Errorf("unsupported CICD platform %s for file %s from repository %s", workflowFile.Origin, workflowFile.Filename, project.PathWithNamespace)
			continue
		}
	}

	return processingError
}

func (gc *GitlabConnector) getPipelineFiles(project *gitlab.Project) (chan *PipelineFile, error) {
	pipelineFilesChan := make(chan *PipelineFile)

	var getEntitiesErr error
	go func() {
		defer close(pipelineFilesChan)

		treeNodes, _, err := gc.client.Repositories.ListTree(project.ID, &gitlab.ListTreeOptions{})
		if err != nil {
			getEntitiesErr = err
			return
		}

		for _, cicdPlatform := range connectors.SUPPORTED_CICD_PLATFORMS {
			relevantFilesPaths := gc.matchedFiles(treeNodes, cicdPlatform.RelevantFilesRegex)
			for _, filePath := range relevantFilesPaths {
				pipelineFilesChan <- &PipelineFile{
					Filename: filePath,
					Origin:   cicdPlatform.Name,
				}
			}
		}
	}()

	return pipelineFilesChan, getEntitiesErr
}

func (gc *GitlabConnector) matchedFiles(treeNodes []*gitlab.TreeNode, regex string) []string {
	var matchedFiles []string
	for _, fileEntry := range treeNodes {
		// skip if entry is a folder
		if fileEntry.Type == "tree" {
			continue
		}

		filepath := fileEntry.Path
		if matched, _ := regexp.MatchString(regex, filepath); matched {
			matchedFiles = append(matchedFiles, filepath)
		}
	}

	return matchedFiles
}

func (gc *GitlabConnector) addProject(gitlabJsonObject map[string]*GitlabGroup, project *gitlab.Project) error {
	groupName := project.Namespace.Name
	projectName := project.Path

	fullName := project.PathWithNamespace

	if _, ok := gitlabJsonObject[groupName]; !ok {
		gitlabJsonObject[groupName] = &GitlabGroup{
			Name:     groupName,
			ID:       int(project.Namespace.ID),
			Projects: make(map[string]*GitlabProject),
		}
	}

	escapedProjectName := connectors.EscapeJsonKey(projectName)

	gitlabJsonObject[groupName].Projects[escapedProjectName] = &GitlabProject{
		Name:           projectName,
		FullName:       fullName,
		ID:             project.ID,
		GitlabCi:       make(map[string]*PipelineFile),
		JfrogPipelines: make(map[string]*PipelineFile),
	}

	return nil
}

func (gc *GitlabConnector) getAllProjects(args []string) chan *GitlabProjectApiResponse {
	projectsChan := make(chan *GitlabProjectApiResponse)

	go func() {
		defer close(projectsChan)

		for _, arg := range args {
			ownerWithRepo := connectors.SplitParentRepo(arg)
			group, _, err := gc.client.Groups.GetGroup(ownerWithRepo.Owner, nil)
			if err != nil {
				projectsChan <- &GitlabProjectApiResponse{
					Project: nil,
					Error:   err,
				}

				continue
			}

			var projects []*gitlab.Project

			if ownerWithRepo.Repo == "" {
				projects = group.Projects
			} else {
				for _, project := range group.Projects {
					if project.PathWithNamespace == strings.ToLower(arg) {
						projects = append(projects, project)
						break
					}
				}
			}

			for _, project := range projects {
				projectsChan <- &GitlabProjectApiResponse{
					Project: project,
					Error:   nil,
				}
			}
		}
	}()

	return projectsChan
}
