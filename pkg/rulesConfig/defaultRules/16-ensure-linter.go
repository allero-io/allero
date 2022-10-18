package defaultRules

import (
	"encoding/json"

	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
)

func EnsureLinter(githubData map[string]*githubConnector.GithubOwner, gitlabData map[string]*gitlabConnector.GitlabGroup) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)
	var err error

	if githubData != nil {
		schemaErrors, err = githubErrorsRule16(githubData)
		if err != nil {
			return nil, err
		}
	}

	if gitlabData != nil {
		schemaErrors, err = gitlabErrorsRule16(gitlabData)
		if err != nil {
			return nil, err
		}
	}

	return schemaErrors, nil
}

func githubErrorsRule16(githubData map[string]*githubConnector.GithubOwner) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)

	usesRegexExpressions := []string{
		".*wemake-services/wemake-python-styleguide@.*",
		".*github/super-linter@.*",
		".*oxsecurity/megalinter@.*",
	}

	runRegexExpressions := []string{
		".*^[\\S]*pip install wemake-python-styleguide.*",
		".*^[\\S]*flake8 .*",
		".*^[\\S]*{tool}.*",
		".*docker .* run .*({tool}/)?renovate.*",
		".*mega-linter-runner.*",
	}

	for _, owner := range githubData {
		for _, repo := range owner.Repositories {
			foundLinter := false

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
								foundLinter = true
								break
							}
						}

						for _, regexExpression := range runRegexExpressions {
							if matchRegex(regexExpression, step.Run) {
								foundLinter = true
								break
							}
						}

						if foundLinter {
							break
						}
					}

					if foundLinter {
						break
					}
				}

				if foundLinter {
					break
				}
			}

			if !foundLinter {
				schemaErrors = append(schemaErrors, &SchemaError{
					ErrorLevel:    1,
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

func gitlabErrorsRule16(gitlabData map[string]*gitlabConnector.GitlabGroup) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)

	for _, group := range gitlabData {
		for _, project := range group.Projects {
			foundLinter, err := findLinterRule16(project)
			if err != nil {
				return nil, err
			}

			if !foundLinter {
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

func findLinterRule16(project *gitlabConnector.GitlabProject) (bool, error) {
	scriptRegexExpressions := []string{
		".*^[\\S]*pip install wemake-python-styleguide.*",
		".*^[\\S]*flake8 .*",
		".*^[\\S]*{tool}.*",
		".*docker .* run .*({tool}/)?renovate.*",
		".*mega-linter-runner.*",
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
