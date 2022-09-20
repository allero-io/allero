package githubConnector

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/allero-io/allero/pkg/connectors"
	"github.com/allero-io/allero/pkg/fileManager"
	"github.com/google/go-github/github"
)

type GithubConnector struct {
	client *github.Client
}

type GithubConnectorDependencies struct {
	Client *github.Client
}

type GithubRepositoryApiResponse struct {
	Repository *github.Repository
	Error      error
}

func New(deps *GithubConnectorDependencies) *GithubConnector {
	return &GithubConnector{
		client: deps.Client,
	}
}

func (gc *GithubConnector) Get(args []string) (int, error) {
	repositoriesChan := gc.getAllRepositories(args)

	githubJsonObject := make(map[string]*GithubOwner)
	reposFetchCounter := 0
	for repo := range repositoriesChan {
		if repo.Error != nil {
			return reposFetchCounter, repo.Error
		}
		reposFetchCounter += 1
		err := gc.addRepo(githubJsonObject, repo.Repository)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = gc.processWorkflowFiles(githubJsonObject, repo.Repository)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	githubJson, err := json.MarshalIndent(githubJsonObject, "", "  ")
	if err != nil {
		return reposFetchCounter, err
	}

	alleroHomedir := fileManager.GetAlleroHomedir()
	return reposFetchCounter, fileManager.WriteToFile(fmt.Sprintf("%s/repo_files/github.json", alleroHomedir), githubJson)
}

func (gc *GithubConnector) addRepo(githubJsonObject map[string]*GithubOwner, repo *github.Repository) error {
	if strings.Contains(*repo.Name, ".") {
		return fmt.Errorf("failed fetching repo %s: should not contain a dot", *repo.FullName)
	}

	if _, ok := githubJsonObject[*repo.Owner.Login]; !ok {
		githubJsonObject[*repo.Owner.Login] = &GithubOwner{
			Name:         *repo.Owner.Login,
			Type:         *repo.Owner.Type,
			ID:           int(*repo.Owner.ID),
			Repositories: make(map[string]*GithubRepository),
		}
	}

	languages, err := gc.getProgrammingLanguages(repo)
	if err != nil {
		return err
	}

	githubJsonObject[*repo.Owner.Login].Repositories[*repo.Name] = &GithubRepository{
		Name:                   *repo.Name,
		FullName:               *repo.FullName,
		ID:                     int(*repo.ID),
		ProgrammingLanguages:   languages,
		GithubActionsWorkflows: make(map[string]*PipelineFile),
		JfrogPipelines:         make(map[string]*PipelineFile),
	}

	return nil
}

func (gc *GithubConnector) getProgrammingLanguages(repo *github.Repository) ([]string, error) {
	languagesMapping, _, err := gc.client.Repositories.ListLanguages(context.Background(), *repo.Owner.Login, *repo.Name)
	if err != nil {
		return nil, err
	}

	var languages []string
	for language := range languagesMapping {
		languages = append(languages, language)
	}

	return languages, nil
}

func (gc *GithubConnector) processWorkflowFiles(githubJsonObject map[string]*GithubOwner, repo *github.Repository) error {
	workflowFilesChan, _ := gc.getWorkflowFilesEntities(repo)
	var processingError error

	for workflowFile := range workflowFilesChan {
		content, _, _, err := gc.client.Repositories.GetContents(context.Background(), *repo.Owner.Login, *repo.Name, workflowFile.RelativePath, nil)
		if err != nil {
			processingError = fmt.Errorf("failed to get content for file %s from repository %s", workflowFile.RelativePath, *repo.FullName)
			continue
		}

		byteContent, err := base64.StdEncoding.DecodeString(*content.Content)
		if err != nil {
			processingError = fmt.Errorf("failed to decode content for file %s from repository %s", workflowFile.RelativePath, *repo.FullName)
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

		filenameWithoutPostfix := strings.Split(workflowFile.Filename, ".")[0]

		if workflowFile.Origin == "github_actions" {
			githubJsonObject[*repo.Owner.Login].Repositories[*repo.Name].GithubActionsWorkflows[filenameWithoutPostfix] = workflowFile
		} else if workflowFile.Origin == "jfrog_pipelines" {
			githubJsonObject[*repo.Owner.Login].Repositories[*repo.Name].JfrogPipelines[filenameWithoutPostfix] = workflowFile
		} else {
			processingError = fmt.Errorf("unsupported CICD platform %s for file %s from repository %s", workflowFile.Origin, workflowFile.RelativePath, *repo.FullName)
			continue
		}
	}

	return processingError
}

func (gc *GithubConnector) getWorkflowFilesEntities(repo *github.Repository) (chan *PipelineFile, error) {
	workflowFilesEntitiesChan := make(chan *PipelineFile)

	var getEntitiesErr error
	go func() {
		defer close(workflowFilesEntitiesChan)

		tree, _, err := gc.client.Git.GetTree(context.Background(), *repo.Owner.Login, *repo.Name, *repo.DefaultBranch, true)
		if err != nil {
			return
		}

		for _, cicdPlatform := range connectors.SUPPORTED_CICD_PLATFORMS {
			relevantFilesPaths := gc.matchedFiles(tree, cicdPlatform.RelevantFilesRegex)
			for _, filePath := range relevantFilesPaths {
				workflowFilesEntitiesChan <- &PipelineFile{
					RelativePath: filePath,
					Filename:     path.Base(filePath),
					Origin:       cicdPlatform.Name,
				}
			}
		}
	}()

	return workflowFilesEntitiesChan, getEntitiesErr
}

func (gc *GithubConnector) matchedFiles(tree *github.Tree, regex string) []string {
	var matchedFiles []string
	for _, fileEntry := range tree.Entries {
		// skip if entry is a folder
		if *fileEntry.Type == "tree" {
			continue
		}

		filepath := *fileEntry.Path
		if matched, _ := regexp.MatchString(regex, filepath); matched {
			matchedFiles = append(matchedFiles, filepath)
		}
	}

	return matchedFiles
}

func (gc *GithubConnector) getAllRepositories(args []string) chan *GithubRepositoryApiResponse {
	repositoriesChan := make(chan *GithubRepositoryApiResponse)

	go func() {
		defer close(repositoriesChan)

		for _, arg := range args {
			ownerWithRepo := connectors.SplitParentRepo(arg)

			if ownerWithRepo.Repo != "" {
				fmt.Printf("Start fetching repository %s/%s\n", ownerWithRepo.Owner, ownerWithRepo.Repo)
				repoMetadata, _, err := gc.client.Repositories.Get(context.Background(), ownerWithRepo.Owner, ownerWithRepo.Repo)
				if err != nil {
					err = fmt.Errorf("unable to get repository %s", arg)
				}

				repositoriesChan <- &GithubRepositoryApiResponse{
					Repository: repoMetadata,
					Error:      err,
				}
				fmt.Printf("Finished fetching repository %s/%s\n", ownerWithRepo.Owner, ownerWithRepo.Repo)

			} else {
				ownerType, err := gc.getGithubOwnerType(ownerWithRepo.Owner)
				if err != nil {
					err = fmt.Errorf("unable to get data on owner %s", ownerWithRepo.Owner)

					repositoriesChan <- &GithubRepositoryApiResponse{
						Repository: nil,
						Error:      err,
					}
				} else {
					ownerRepos, err := ListByOwnerWithPagination(gc.client, ownerWithRepo.Owner, ownerType)
					if err != nil {
						err = fmt.Errorf("unable to get repositories from owner %s", ownerWithRepo.Owner)
					}

					for repo := range ownerRepos {
						repositoriesChan <- &GithubRepositoryApiResponse{
							Repository: repo,
							Error:      err,
						}
					}
				}

			}
		}
	}()

	return repositoriesChan
}

func (gc *GithubConnector) getGithubOwnerType(owner string) (string, error) {
	metadata, _, err := gc.client.Users.Get(context.Background(), owner)
	if err != nil || metadata == nil {
		return "", err
	}

	return *metadata.Type, err
}
