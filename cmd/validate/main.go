package validate

import (
	"errors"
	"fmt"

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

type ValidateCommandFlags struct {
	output string
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
			output := cmd.Flag("output").Value.String()
			if !validateOutputFlag(output) {
				return fmt.Errorf("invalid output flag %s", output)
			}

			validateCommandFlags := &ValidateCommandFlags{
				output: output,
			}

			return execute(deps, validateCommandFlags)
		},
	}

	policiesCmd.Flags().StringP("output", "o", "", "Define output format. Can be 'csv'")

	return policiesCmd
}

func validateOutputFlag(output string) bool {
	if output == "" {
		return true
	}

	outputFormats := []string{"csv"}
	for _, verifiedOutput := range outputFormats {
		if output == verifiedOutput {
			return true
		}
	}

	return false
}

func execute(deps *ValidateCommandDependencies, flags *ValidateCommandFlags) error {
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

	err = resultsPrinter.PrintResults(ruleResults, summary, flags.output)
	if err != nil {
		return err
	}

	if !shouldPassExecution {
		return ErrViolationsFound
	}

	return nil
}
