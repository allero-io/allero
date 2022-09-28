package localConnector

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/allero-io/allero/pkg/connectors"
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

	localJsonObject := make(map[string]*LocalOwner)

	err := lc.addRootPathAsRepo(localJsonObject)
	if err != nil {
		return err
	}

	// err = lc.processWorkflowFiles(localJsonObject, "local")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	localJson, err := json.MarshalIndent(localJsonObject, "", "  ")
	if err != nil {
		return err
	}

	alleroHomedir := fileManager.GetAlleroHomedir()
	return fileManager.WriteToFile(fmt.Sprintf("%s/repo_files/local.json", alleroHomedir), localJson)
}

func (lc *LocalConnector) addRootPathAsRepo(localJsonObject map[string]*LocalOwner) error {
	localJsonObject["local"] = &LocalOwner{
		Name:         "sudo",
		Type:         "",
		ID:           0,
		Repositories: make(map[string]*LocalRepository),
	}

	escapedRepoName := connectors.EscapeJsonKey(lc.RootPath)

	localJsonObject["local"].Repositories[escapedRepoName] = &LocalRepository{
		Name:                  escapedRepoName,
		FullName:              escapedRepoName,
		ID:                    0,
		ProgrammingLanguages:  nil,
		LocalActionsWorkflows: make(map[string]*PipelineFile),
		JfrogPipelines:        make(map[string]*PipelineFile),
	}

	return nil
}

func (lc *LocalConnector) processWorkflowFiles(localJsonObject map[string]*LocalOwner, rootPath string) error {
	// workflowFilesChan, _ := lc.getWorkflowFilesEntities(repo)
	// var processingError error

	// for workflowFile := range workflowFilesChan {
	// 	content, _, _, err := lc.client.Repositories.GetContents(context.Background(), *repo.Owner.Login, *repo.Name, workflowFile.RelativePath, nil)
	// 	if err != nil {
	// 		processingError = fmt.Errorf("failed to get content for file %s from repository %s", workflowFile.RelativePath, *repo.FullName)
	// 		continue
	// 	}

	// 	byteContent, err := base64.StdEncoding.DecodeString(*content.Content)
	// 	if err != nil {
	// 		processingError = fmt.Errorf("failed to decode content for file %s from repository %s", workflowFile.RelativePath, *repo.FullName)
	// 		continue
	// 	}

	// 	jsonContentBytes, err := connectors.YamlToJson(byteContent)
	// 	if err != nil {
	// 		processingError = err
	// 		continue
	// 	}

	// 	jsonContent := make(map[string]interface{})
	// 	err = json.Unmarshal(jsonContentBytes, &jsonContent)
	// 	if err != nil {
	// 		processingError = err
	// 		continue
	// 	}

	// 	workflowFile.Content = jsonContent
	// 	escapedFilename := connectors.EscapeJsonKey(workflowFile.Filename)

	// 	if workflowFile.Origin == "local_actions" {
	// 		localJsonObject[*repo.Owner.Login].Repositories[*repo.Name].LocalActionsWorkflows[escapedFilename] = workflowFile
	// 	} else if workflowFile.Origin == "jfrog_pipelines" {
	// 		localJsonObject[*repo.Owner.Login].Repositories[*repo.Name].JfrogPipelines[escapedFilename] = workflowFile
	// 	} else {
	// 		processingError = fmt.Errorf("unsupported CICD platform %s for file %s from repository %s", workflowFile.Origin, workflowFile.RelativePath, *repo.FullName)
	// 		continue
	// 	}
	// }

	// return processingError
	return nil
}

// func (lc *LocalConnector) getWorkflowFilesEntities(repo *local.Repository) (chan *PipelineFile, error) {
// 	workflowFilesEntitiesChan := make(chan *PipelineFile)

// 	var getEntitiesErr error
// 	go func() {
// 		defer close(workflowFilesEntitiesChan)

// 		tree, _, err := lc.client.Git.GetTree(context.Background(), *repo.Owner.Login, *repo.Name, *repo.DefaultBranch, true)
// 		if err != nil {
// 			return
// 		}

// 		for _, cicdPlatform := range connectors.SUPPORTED_CICD_PLATFORMS {
// 			relevantFilesPaths := lc.matchedFiles(tree, cicdPlatform.RelevantFilesRegex)
// 			for _, filePath := range relevantFilesPaths {
// 				workflowFilesEntitiesChan <- &PipelineFile{
// 					RelativePath: filePath,
// 					Filename:     path.Base(filePath),
// 					Origin:       cicdPlatform.Name,
// 				}
// 			}
// 		}
// 	}()

// 	return workflowFilesEntitiesChan, getEntitiesErr
// }

// func (lc *LocalConnector) matchedFiles(tree *local.Tree, regex string) []string {
// 	var matchedFiles []string
// 	for _, fileEntry := range tree.Entries {
// 		// skip if entry is a folder
// 		if *fileEntry.Type == "tree" {
// 			continue
// 		}

// 		filepath := *fileEntry.Path
// 		if matched, _ := regexp.MatchString(regex, filepath); matched {
// 			matchedFiles = append(matchedFiles, filepath)
// 		}
// 	}

// 	return matchedFiles
// }
