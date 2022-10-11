package defaultRules

import (
	"encoding/json"

	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
)

func EnsureCodeQualityScanner(githubData map[string]*githubConnector.GithubOwner, gitlabData map[string]*gitlabConnector.GitlabGroup) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)
	var err error

	if githubData != nil {
		schemaErrors, err = githubErrorsRule17(githubData)
		if err != nil {
			return nil, err
		}
	}

	if gitlabData != nil {
		schemaErrors, err = gitlabErrorsRule17(gitlabData)
		if err != nil {
			return nil, err
		}
	}

	return schemaErrors, nil
}

func githubErrorsRule17(githubData map[string]*githubConnector.GithubOwner) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)

	usesRegexExpressions := []string{
		".*paambaati/codeclimate-action@.*",
		".*kitabisa/sonarqube-action@.*",
		".*sonarsource/sonarcloud-github-action@.*",
	}

	runRegexExpressions := []string{
		".*docker .* run .*codeclimate/codeclimate analyze.*",
	}

	for _, owner := range githubData {
		for _, repo := range owner.Repositories {
			foundCodeQualityScanner := false

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
								foundCodeQualityScanner = true
								break
							}
						}

						for _, regexExpression := range runRegexExpressions {
							if matchRegex(regexExpression, step.Run) {
								foundCodeQualityScanner = true
								break
							}
						}

						if foundCodeQualityScanner {
							break
						}
					}

					if foundCodeQualityScanner {
						break
					}
				}

				if foundCodeQualityScanner {
					break
				}
			}

			if !foundCodeQualityScanner {
				schemaErrors = append(schemaErrors, &SchemaError{
					ErrorLevel:    1,
					RepositryName: repo.Name,
					CiCdPlatform:  "github-actions-workflows",
					OwnerName:     owner.Name,
					ScmPlatform:   "github",
				})
			}
		}
	}

	return schemaErrors, nil
}

func gitlabErrorsRule17(gitlabData map[string]*gitlabConnector.GitlabGroup) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)

	for _, group := range gitlabData {
		for _, project := range group.Projects {
			foundCodeQualityScanner, err := findCodeQualityScannerRule17(project)
			if err != nil {
				return nil, err
			}

			if !foundCodeQualityScanner {
				schemaErrors = append(schemaErrors, &SchemaError{
					ErrorLevel:    2,
					RepositryName: project.Name,
					CiCdPlatform:  "gitlab-ci",
					OwnerName:     group.Name,
					ScmPlatform:   "gitlab",
				})
			}
		}
	}

	return schemaErrors, nil
}

func findCodeQualityScannerRule17(project *gitlabConnector.GitlabProject) (bool, error) {
	scriptRegexExpressions := []string{
		".*docker .* run .*codeclimate/codeclimate analyze.*",
	}

	for _, pipeline := range project.GitlabCi {

		for _, stage := range pipeline.Content {
			stageBytes, err := json.Marshal(stage)
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
				for _, regexExpression := range scriptRegexExpressions {
					if matchRegex(regexExpression, stageWithSingleScript.Script) {
						return true, nil
					}
				}
			}

			if stageWithScripts.Scripts != nil {
				for _, script := range stageWithScripts.Scripts {
					for _, regexExpression := range scriptRegexExpressions {
						if matchRegex(regexExpression, script) {
							return true, nil
						}
					}
				}
			}
		}
	}

	return false, nil
}
