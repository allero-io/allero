package validate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/allero-io/allero/pkg/configurationManager"
	localConnector "github.com/allero-io/allero/pkg/connectors/local"
	"github.com/allero-io/allero/pkg/mapStructureEncoder"
	"github.com/allero-io/allero/pkg/posthog"
	"github.com/allero-io/allero/pkg/resultsPrinter"
	"github.com/allero-io/allero/pkg/rulesConfig"
	"github.com/allero-io/allero/pkg/rulesConfig/defaultRules"
	"github.com/spf13/cobra"
)

var ErrViolationsFound = errors.New("")
var (
	summary rulesConfig.OutputSummary
)

type ValidateCommandDependencies struct {
	RulesConfig             *rulesConfig.RulesConfig
	ConfigurationManager    *configurationManager.ConfigurationManager
	PosthogClient           *posthog.PosthogClient
	LocalRepositoriesClient *localConnector.LocalConnector
}

type validateCommandOptions struct {
	output              string
	ignoreToken         bool
	localPathToValidate string
}

type wrapper struct {
	err error
}

var SCM_PLATFORMS = []string{"github", "gitlab"}

func (w *wrapper) Run(f func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		err := f(cmd, args)
		w.err = err
	}
}

func New(deps *ValidateCommandDependencies) *cobra.Command {
	cmdWrap := wrapper{}
	var policiesCmd = &cobra.Command{
		Use:   "validate [OPTIONAL] PATH",
		Short: "Validate set of default rules",
		Long:  "Validate set of default rules over fetched data or the given path",
		Example: `allero validate                     Validate over fetched repositories
allero validate .                   Validate over current directory
allero validate ~/my-repo-dir       Validate over local directory`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MaximumNArgs(1),
		PreRun: func(cmd *cobra.Command, cmdArgs []string) {
			args := make(map[string]any)

			if len(cmdArgs) > 0 {
				absolutePath, err := filepath.Abs(cmdArgs[0])
				if err == nil {
					cmdArgs[0] = absolutePath
				}
			}
			args["Args"] = cmdArgs

			decodedToken, _ := deps.ConfigurationManager.ParseToken()
			if decodedToken != nil {
				args["User Email"] = decodedToken.Email
			}

			args["Github Actor"] = os.Getenv("GITHUB_ACTOR")
			args["Github Repository"] = os.Getenv("GITHUB_REPOSITORY")

			deps.PosthogClient.PublishEventWithArgs("data validated", args)
		},
		Run: cmdWrap.Run(func(cmd *cobra.Command, args []string) error {
			output := cmd.Flag("output").Value.String()
			if !validateOutputFlag(output) {
				return fmt.Errorf("invalid output flag %s", output)
			}

			ignoreToken := cmd.Flag("ignore-token").Value.String() == "true"
			localPathToValidate := ""
			if len(args) > 0 {
				localPathToValidate = args[0]
			}

			validateCommandFlags := &validateCommandOptions{
				output:              output,
				ignoreToken:         ignoreToken,
				localPathToValidate: localPathToValidate,
			}

			return execute(deps, validateCommandFlags)
		}),
		PostRunE: func(cmd *cobra.Command, args []string) error {
			summaryProperties, err := mapStructureEncoder.Encode(summary)
			if err != nil {
				return err
			}
			deps.PosthogClient.PublishEventWithArgs("data validated summary", summaryProperties)
			return cmdWrap.err
		},
	}

	policiesCmd.Flags().StringP("output", "o", "", "Define output format. Can be 'csv'")
	policiesCmd.Flags().Bool("ignore-token", false, "Ignore token and run as anonymous user")

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

func execute(deps *ValidateCommandDependencies, option *validateCommandOptions) error {
	isLocal := option.localPathToValidate != ""
	var err error

	if isLocal {
		err = deps.LocalRepositoriesClient.Get(option.localPathToValidate)
		if err == nil {
			fmt.Printf("Running validation over %s\n", option.localPathToValidate)
			deps.RulesConfig.ReadLocalData()
		}
	}
	if err != nil {
		return err
	}

	err = deps.RulesConfig.Initialize()
	if err != nil {
		return err
	}

	shouldPassExecution := true
	summary = deps.RulesConfig.GetSummary(option.localPathToValidate != "")
	totalRulesFailed := 0

	ruleNamesByScmPlatform := map[string][]string{}
	for _, scmPlatform := range SCM_PLATFORMS {
		ruleNamesByScmPlatform[scmPlatform] = deps.RulesConfig.GetAllRuleNames(scmPlatform)
	}

	hasToken := false
	selectedRuleIds := make(map[int]bool)

	if !option.ignoreToken {
		selectedRuleIds, err = deps.RulesConfig.GetSelectedRuleIds()
		if err != nil {
			return err
		}
		hasToken = selectedRuleIds != nil
	}

	ruleResultsById := map[int]*rulesConfig.RuleResult{}
	disabledRules := map[string]bool{}

	for scmPlatform, ruleNames := range ruleNamesByScmPlatform {
		for _, ruleName := range ruleNames {
			rule, err := deps.RulesConfig.GetRule(ruleName, scmPlatform)
			if err != nil {
				return err
			}

			isCustomRule := rule.UniqueId >= 1000

			if isRuleDisabled(hasToken, selectedRuleIds, rule, isCustomRule) {
				disabledRules[ruleName] = true
				continue
			}

			schemaErrors, err := deps.RulesConfig.Validate(ruleName, rule, scmPlatform)
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

			ruleResult := &rulesConfig.RuleResult{
				RuleName:       ruleName,
				Valid:          isRuleValid,
				SchemaErrors:   schemaErrors,
				FailureMessage: rule.FailureMessage,
			}

			if ruleResultsById[rule.UniqueId] == nil {
				ruleResultsById[rule.UniqueId] = ruleResult
			} else {
				ruleResultsById[rule.UniqueId].SchemaErrors = append(ruleResultsById[rule.UniqueId].SchemaErrors, schemaErrors...)
				if !isRuleValid {
					ruleResultsById[rule.UniqueId].Valid = false
				}
			}
		}
	}

	summary.TotalRulesEvaluated = len(ruleResultsById)
	summary.TotalFailedRules = totalRulesFailed
	summary.URL = deps.ConfigurationManager.TokenGenerationUrl

	if isLocal {
		ruleResultsById = reduceLocalRuleResults(ruleResultsById)
	}

	err = resultsPrinter.PrintResults(ruleResultsById, summary, disabledRules, option.output, isLocal)
	if err != nil {
		return err
	}
	if !shouldPassExecution {
		return ErrViolationsFound
	}
	return nil
}

func isRuleDisabled(hasToken bool, selectedRuleIds map[int]bool, rule *defaultRules.Rule, isCustomRule bool) bool {
	if hasToken && !selectedRuleIds[rule.UniqueId] && !isCustomRule {
		return true
	} else if !hasToken && !rule.EnabledByDefault {
		return true
	}

	return false
}

func reduceLocalRuleResults(ruleResultsById map[int]*rulesConfig.RuleResult) map[int]*rulesConfig.RuleResult {
	reducedRuleResultsById := map[int]*rulesConfig.RuleResult{}
	for uniqueId, ruleResult := range ruleResultsById {
		if ruleResult.Valid {
			reducedRuleResultsById[uniqueId] = ruleResult
			continue
		}

		schemaErrorsByScmPlatform := map[string][]*defaultRules.SchemaError{}
		for _, schemaError := range ruleResult.SchemaErrors {
			schemaErrorsByScmPlatform[schemaError.ScmPlatform] = append(schemaErrorsByScmPlatform[schemaError.ScmPlatform], schemaError)
		}

		maxErrors := 0
		schemaErrors := []*defaultRules.SchemaError{}
		for _, scmSchemaErrors := range schemaErrorsByScmPlatform {
			if len(scmSchemaErrors) > maxErrors {
				maxErrors = len(scmSchemaErrors)
				schemaErrors = scmSchemaErrors
			}
		}

		reducedRuleResultsById[uniqueId] = &rulesConfig.RuleResult{
			RuleName:       ruleResult.RuleName,
			Valid:          false,
			SchemaErrors:   schemaErrors,
			FailureMessage: ruleResult.FailureMessage,
		}
	}
	return reducedRuleResultsById
}
