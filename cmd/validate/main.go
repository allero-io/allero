package validate

import (
	"errors"

	"github.com/allero-io/allero/pkg/alleroBackendClient"
	"github.com/allero-io/allero/pkg/configurationManager"
	"github.com/allero-io/allero/pkg/posthog"
	"github.com/allero-io/allero/pkg/resultsPrinter"
	"github.com/allero-io/allero/pkg/rulesConfig"
	"github.com/spf13/cobra"
)

var ErrViolationsFound = errors.New("")

type ValidateCommandDependencies struct {
	RulesConfig          *rulesConfig.RulesConfig
	ConfigurationManager *configurationManager.ConfigurationManager
	PosthogClient        *posthog.PosthogClient
	AlleroBackendClient  *alleroBackendClient.AlleroBackendClient
}

func New(deps *ValidateCommandDependencies) *cobra.Command {
	var policiesCmd = &cobra.Command{
		Use:           "validate",
		Short:         "Validate set of default rules",
		Long:          "Validate set of default rules over all fetched data",
		Example:       `allero validate`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			deps.PosthogClient.PublishCmdUse("validate", args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(deps)
		},
	}

	return policiesCmd
}

func execute(deps *ValidateCommandDependencies) error {
	err := deps.RulesConfig.Initialize()
	if err != nil {
		return err
	}

	ruleResults := []*rulesConfig.RuleResult{}
	shouldPassExecution := true
	summary := deps.RulesConfig.GetSummary()
	totalRulesFailed := 0
	ruleNames := deps.RulesConfig.GetAllRuleNames()

	for _, ruleName := range ruleNames {
		rule, err := deps.RulesConfig.GetRule(ruleName)
		if err != nil {
			return err
		}

		schemaErrors, err := deps.RulesConfig.Validate(ruleName, rule)
		if err != nil {
			return err
		}

		if shouldPassExecution {
			shouldPassExecution = len(schemaErrors) == 0
		}

		isRuleValid := len(schemaErrors) == 0
		if !isRuleValid {
			totalRulesFailed++
		}

		ruleResults = append(ruleResults, &rulesConfig.RuleResult{
			RuleName:       ruleName,
			Valid:          isRuleValid,
			SchemaErrors:   schemaErrors,
			FailureMessage: rule.FailureMessage,
		})
	}

	summary.TotalRulesEvaluated = len(ruleResults)
	summary.TotalFailedRules = totalRulesFailed

	resultsPrinter.PrintResults(ruleResults, summary)

	if !shouldPassExecution {
		return ErrViolationsFound
	}

	return nil
}
