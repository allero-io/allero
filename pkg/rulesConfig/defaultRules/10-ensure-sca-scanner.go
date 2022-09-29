package defaultRules

import (
	"encoding/json"
	"fmt"

	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
)

func EnsureScaScanner(githubData map[string]*githubConnector.GithubOwner, gitlabData map[string]*gitlabConnector.GitlabGroup) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)
	var err error

	if githubData != nil {
		schemaErrors, err = githubErrorsRule10(githubData)
		if err != nil {
			return nil, err
		}
	}

	if gitlabData != nil {
		schemaErrors, err = gitlabErrorsRule10(gitlabData)
		if err != nil {
			return nil, err
		}
	}

	return schemaErrors, nil
}

func githubErrorsRule10(githubData map[string]*githubConnector.GithubOwner) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)

	usesRegexExpressions := []string{
		".*anchore/scan-action@.*",
		".*synopsys-sig/detect-action@.*",
		".*aquasecurity/trivy-action@.*",
		".*checkmarx-ts/checkmarx-cxflow-github-action@.*",
		".*snyk/actions/maven@.*",
	}

	runRegexExpressions := []string{
		".*^[\\S]*trivy.*|.*docker .* run .*(aquasec/)?trivy.*",
		"^[\\S]*grype|docker .* run .*(anchore/)?grype.*",
		"(jfrog|jf) (s|scan).*",
		"ws scan.*",
		"snyk (code | )test.*",
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

				var workflowObj Workflow
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
				schemaErrors = append(schemaErrors, &SchemaError{
					ErrorLevel:    2,
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

func gitlabErrorsRule10(gitlabData map[string]*gitlabConnector.GitlabGroup) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)

	for _, group := range gitlabData {
		for _, project := range group.Projects {
			foundScaScanner, err := findScaScannerRule10(project)
			if err != nil {
				return nil, err
			}

			if !foundScaScanner {
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

func findScaScannerRule10(project *gitlabConnector.GitlabProject) (bool, error) {

	imageRegexExpressions := []string{
		"registry.gitlab.com/secure.*",
	}

	scriptRegexExpressions := []string{
		".*^[\\S]*trivy.*|.*docker .* run .*(aquasec/)?trivy.*",
		"^[\\S]*grype|docker .* run .*(anchore/)?grype.*",
		"(jfrog|jf) (s|scan).*",
		"ws scan.*",
		"snyk ?(code | )test.*",
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
