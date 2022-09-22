package defaultRules

import (
	"encoding/json"

	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
)

func EnsureScaScanner(githubData map[string]*githubConnector.GithubOwner) ([]*SchemaError, error) {
	schemaErrors := make([]*SchemaError, 0)

	usesRegexExpressions := []string{
		".*anchore/scan-action@.*",
		".*synopsys-sig/detect-action@.*",
		".*aquasecurity/trivy-action@.*",
	}

	runRegexExpressions := []string{
		".*^[\\S]*trivy.*|.*docker .* run .*(aquasec/)?trivy.*",
		"^[\\S]*grype|docker .* run .*(anchore/)?grype.*",
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
					ErrorLevel:    1,
					RepositryName: repo.Name,
					CiCdPlatform:  "github-actions-workflows",
					OwnerName:     owner.Name,
				})
			}
		}
	}

	return schemaErrors, nil
}
