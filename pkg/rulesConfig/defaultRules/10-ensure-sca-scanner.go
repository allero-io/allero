package defaultRules

import (
	"encoding/json"
	"fmt"

	"github.com/allero-io/allero/pkg/connectors"
	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
)

func EnsureScaScanner(githubData map[string]*githubConnector.GithubOwner, gitlabData map[string]*gitlabConnector.GitlabGroup) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)
	var err error

	sharedRegexExpressions := []string{
		"^[\\S]*trivy.*|.*docker( .*)? run .*(aquasec/)?trivy.*",
		"^[\\S]*grype.*|.*docker( .*)? run .*(anchore/)?grype.*",
		"(jfrog|jf) (s|scan).*",
		"ws scan.*",
		"snyk (code |)test.*",
		"(jfrog|jf) (xr).*",
	}

	if githubData != nil {
		schemaErrors, err = githubErrorsRule10(githubData, sharedRegexExpressions)
		if err != nil {
			return nil, err
		}
	}

	if gitlabData != nil {
		schemaErrors, err = gitlabErrorsRule10(gitlabData, sharedRegexExpressions)
		if err != nil {
			return nil, err
		}
	}

	return schemaErrors, nil
}

func githubErrorsRule10(githubData map[string]*githubConnector.GithubOwner, runRegexExpressions []string) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)

	usesRegexExpressions := []string{
		".*anchore/scan-action@.*",
		".*synopsys-sig/detect-action@.*",
		".*aquasecurity/trivy-action@.*",
		".*checkmarx-ts/checkmarx-cxflow-github-action@.*",
		".*snyk/actions/maven@.*",
	}

	for _, owner := range githubData {
		for _, repo := range owner.Repositories {
			foundScaScanner := false

			for _, workflow := range repo.GithubActionsWorkflows {
				content := workflow.Content
				contentByteArr, err := json.Marshal(content)
				if err != nil {
					return nil, err
				}

				var workflowObj GithubWorkflow
				err = json.Unmarshal(contentByteArr, &workflowObj)
				if err != nil {
					return nil, err
				}

				for _, job := range workflowObj.Jobs {
					for _, step := range job.Steps {
						for _, regexExpression := range usesRegexExpressions {
							if matchRegex(regexExpression, step.Uses) {
								foundScaScanner = true
								break
							}
						}

						for _, regexExpression := range runRegexExpressions {
							if matchRegex(regexExpression, step.Run) {
								foundScaScanner = true
								break
							}
						}

						if foundScaScanner {
							break
						}
					}

					if foundScaScanner {
						break
					}
				}

				if foundScaScanner {
					break
				}
			}

			if !foundScaScanner {
				var err error
				foundScaScanner, err = findJfrogScanScannerRule10(repo.JfrogPipelines, runRegexExpressions)

				if err != nil {
					return nil, err
				}
			}

			if !foundScaScanner {
				schemaErrors = append(schemaErrors, &SchemaError{
					ErrorLevel:    2,
					RepositryName: repo.Name,
					CiCdPlatform:  "",
					OwnerName:     owner.Name,
					ScmPlatform:   "github",
				})
			}
		}
	}

	return schemaErrors, nil
}

func gitlabErrorsRule10(gitlabData map[string]*gitlabConnector.GitlabGroup, scriptsRegexExpressions []string) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)

	for _, group := range gitlabData {
		for _, project := range group.Projects {
			foundScaScanner, err := findGitlabScaScannerRule10(project, scriptsRegexExpressions)
			if err != nil {
				return nil, err
			}

			if !foundScaScanner {
				schemaErrors = append(schemaErrors, &SchemaError{
					ErrorLevel:    2,
					RepositryName: project.Name,
					CiCdPlatform:  "",
					OwnerName:     group.Name,
					ScmPlatform:   "gitlab",
				})
			}
		}
	}

	return schemaErrors, nil
}

func findGitlabScaScannerRule10(project *gitlabConnector.GitlabProject, scriptsRegexExpressions []string) (bool, error) {

	imageRegexExpressions := []string{
		"registry.gitlab.com/secure.*",
	}

	for _, pipeline := range project.GitlabCi {

		for key, value := range pipeline.Content {
			if key == "image" {
				imageValue := fmt.Sprintf("%v", value)
				for _, imageRegexExpression := range imageRegexExpressions {
					if matchRegex(imageRegexExpression, imageValue) {
						return true, nil
					}
				}
			}
			stageBytes, err := json.Marshal(value)
			if err != nil {
				return false, err
			}

			var stageWithSingleScript GitlabStageScript
			var stageWithScripts GitlabStageScripts

			err = json.Unmarshal(stageBytes, &stageWithSingleScript)
			if err != nil {
				err = json.Unmarshal(stageBytes, &stageWithScripts)
				if err != nil {
					continue
				}
			}

			if stageWithSingleScript.Script != "" {
				for _, regexExpression := range scriptsRegexExpressions {
					if matchRegex(regexExpression, stageWithSingleScript.Script) {
						return true, nil
					}
				}
			}

			if stageWithScripts.Scripts != nil {
				for _, script := range stageWithScripts.Scripts {
					for _, regexExpression := range scriptsRegexExpressions {
						if matchRegex(regexExpression, script) {
							return true, nil
						}
					}
				}
			}
		}
	}

	foundInJfrog, err := findJfrogScanScannerRule10(project.JfrogPipelines, scriptsRegexExpressions)
	if err != nil {
		return false, err
	}

	return foundInJfrog, nil
}

func findJfrogScanScannerRule10(jfrogPipelines map[string]*connectors.PipelineFile, executionRegexExpressions []string) (bool, error) {
	jfrogPiplineFiles := make([]*JfrogPipelineFile, 0)

	for _, pipeline := range jfrogPipelines {
		content := pipeline.Content
		contentByteArr, err := json.Marshal(content)
		if err != nil {
			return false, err
		}

		var jfrogPipelineFile JfrogPipelineFile
		err = json.Unmarshal(contentByteArr, &jfrogPipelineFile)
		if err != nil {
			return false, err
		}

		jfrogPiplineFiles = append(jfrogPiplineFiles, &jfrogPipelineFile)
	}

	for _, jfrogPipelineFile := range jfrogPiplineFiles {
		for _, pipeline := range jfrogPipelineFile.Pipelines {
			for _, step := range pipeline.Steps {
				for _, executionCommand := range step.Execution.OnExecute {
					for _, regexExpression := range executionRegexExpressions {
						if matchRegex(regexExpression, executionCommand) {
							return true, nil
						}
					}
				}
			}
		}
	}

	return false, nil
}
