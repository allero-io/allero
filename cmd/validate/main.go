package validate

import (
	"errors"
	"fmt"

	"github.com/allero-io/allero/pkg/configurationManager"
	localConnector "github.com/allero-io/allero/pkg/connectors/local"
	"github.com/allero-io/allero/pkg/mapStructureEncoder"
	"github.com/allero-io/allero/pkg/posthog"
	"github.com/allero-io/allero/pkg/resultsPrinter"
	"github.com/allero-io/allero/pkg/rulesConfig"
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

type ValidateCommandFlags struct {
	output        string
	ignoreToken   bool
	failOnNoFetch bool
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
		Use:           "validate",
		Short:         "Validate set of default rules",
		Long:          "Validate set of default rules over all fetched data",
		Example:       `allero validate`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, cmdArgs []string) {
			args := make(map[string]any)
			args["Args"] = cmdArgs
			decodedToken, _ := deps.ConfigurationManager.ParseToken()
			if decodedToken != nil {
				args["User Email"] = decodedToken.Email
			}
			deps.PosthogClient.PublishEventWithArgs("data validated", args)
		},
		Run: cmdWrap.Run(func(cmd *cobra.Command, args []string) error {
			output := cmd.Flag("output").Value.String()
			if !validateOutputFlag(output) {
				return fmt.Errorf("invalid output flag %s", output)
			}

			ignoreToken := cmd.Flag("ignore-token").Value.String() == "true"
			failOnNoFetch := cmd.Flag("fail-on-no-fetch").Value.String() == "true"

			validateCommandFlags := &ValidateCommandFlags{
				output:        output,
				ignoreToken:   ignoreToken,
				failOnNoFetch: failOnNoFetch,
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
	policiesCmd.Flags().Bool("fail-on-no-fetch", false, "Fail if validate command is run without any fetched data")

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
	if !flags.failOnNoFetch && err == rulesConfig.NoFetchedDataError {
		err := deps.LocalRepositoriesClient.Get()
		if err != nil {
			return err
		} else {
			fmt.Printf("No fetched data found. Running validation on local data fetch from %s.\n", deps.LocalRepositoriesClient.RootPath)
			deps.RulesConfig.ReadLocalData()
		}
	} else if err != nil {
		return err
	}

	shouldPassExecution := true
	summary = deps.RulesConfig.GetSummary()
	totalRulesFailed := 0

	ruleNamesByScmPlatform := map[string][]string{}
	for _, scmPlatform := range SCM_PLATFORMS {
		ruleNamesByScmPlatform[scmPlatform] = deps.RulesConfig.GetAllRuleNames(scmPlatform)
	}

	hasToken := false
	selectedRuleIds := make(map[int]bool)

	if !flags.ignoreToken {
		selectedRuleIds, err = deps.RulesConfig.GetSelectedRuleIds()
		if err != nil {
			return err
		}
		hasToken = selectedRuleIds != nil
	}

	ruleResultsById := map[int]*rulesConfig.RuleResult{}

	for scmPlaform, ruleNames := range ruleNamesByScmPlatform {
		for _, ruleName := range ruleNames {
			rule, err := deps.RulesConfig.GetRule(ruleName, scmPlaform)
			if err != nil {
				return err
			}

			if hasToken && !selectedRuleIds[rule.UniqueId] {
				continue
			} else if !hasToken && !rule.EnabledByDefault {
				continue
			}

			schemaErrors, err := deps.RulesConfig.Validate(ruleName, rule, scmPlaform)
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

	if !hasToken {
		summary.URL = deps.ConfigurationManager.TokenGenerationUrl
	}

	err = resultsPrinter.PrintResults(ruleResultsById, summary, flags.output)
	if err != nil {
		return err
	}
	if !shouldPassExecution {
		return ErrViolationsFound
	}
	return nil
}
