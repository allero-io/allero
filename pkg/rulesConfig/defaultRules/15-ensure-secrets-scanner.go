package defaultRules

import (
	"encoding/json"
	"fmt"
	"strings"

	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	gitlabConnector "github.com/allero-io/allero/pkg/connectors/gitlab"
)

func EnsureSecretsScanner(githubData map[string]*githubConnector.GithubOwner, gitlabData map[string]*gitlabConnector.GitlabGroup) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)
	var err error

	if githubData != nil {
		schemaErrors, err = githubErrorsRule15(githubData)
		if err != nil {
			return nil, err
		}
	}

	if gitlabData != nil {
		schemaErrors, err = gitlabErrorsRule15(gitlabData)
		if err != nil {
			return nil, err
		}
	}

	return schemaErrors, nil
}

func githubErrorsRule15(githubData map[string]*githubConnector.GithubOwner) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)

	usesRegexExpressions := []string{
		".*trufflesecurity/trufflehog@.*",
		".*GitGuardian/ggshield/actions/secret@.*",
		".*GitGuardian/ggshield-action@.*",
		".*gitleaks/gitleaks-action@.*",
	}

	usesAndWithRegexExpressions := []map[string]string{
		{
			"uses": ".*aquasecurity/trivy-action@.*",
			"with": "security-checks:secret",
		},
	}

	runRegexExpressions := []string{
		".*^[\\S]*trufflehog.*|.*docker .* run .*(trufflesecurity/)?trufflehog.*",
		".*ggshield secret scan.*",
		".*^[\\S]*gitleaks.*|.*docker .* run .*(zricethezav/)?gitleaks.*.*",
		".*^[\\S]*trivy fs.*|.*docker .* run .*(aquasec/)?trivy fs.*",
	}
	for _, owner := range githubData {
		for _, repo := range owner.Repositories {
			foundSecretsScanner := false

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
								foundSecretsScanner = true
								break
							}
						}

						for _, regexExpression := range runRegexExpressions {
							if matchRegex(regexExpression, step.Run) {
								foundSecretsScanner = true
								break
							}
						}

						for _, regexExpression := range usesAndWithRegexExpressions {
							if matchRegex(regexExpression["uses"], step.Uses) {
								withRegex := strings.Split(regexExpression["with"], ":")
								withRegexKey, withRegexValue := withRegex[0], withRegex[1]
								for withKey, withValue := range step.With {
									withValueAsString := fmt.Sprintf("%v", withValue)
									if withKey == withRegexKey && matchRegex(withRegexValue, withValueAsString) {
										foundSecretsScanner = true
									}
								}
							}
						}
						if foundSecretsScanner {
							break
						}
					}

					if foundSecretsScanner {
						break
					}
				}

				if foundSecretsScanner {
					break
				}
			}

			if !foundSecretsScanner {
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

func gitlabErrorsRule15(gitlabData map[string]*gitlabConnector.GitlabGroup) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)

	for _, group := range gitlabData {
		for _, project := range group.Projects {
			foundScaScanner, err := findSecretsScannerRule15(project)
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

func findSecretsScannerRule15(project *gitlabConnector.GitlabProject) (bool, error) {
	scriptRegexExpressions := []string{
		".*^[\\S]*trufflehog.*|.*docker .* run .*(trufflesecurity/)?trufflehog.*",
		".*ggshield secret scan.*",
		".*^[\\S]*gitleaks.*|.*docker .* run .*(zricethezav/)?gitleaks.*.*",
		".*^[\\S]*trivy fs.*|.*docker .* run .*(aquasec/)?trivy fs.*",
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
