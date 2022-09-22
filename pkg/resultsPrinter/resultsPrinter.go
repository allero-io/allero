package resultsPrinter

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/allero-io/allero/pkg/rulesConfig"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
)

var CICD_PLATFORMS = map[string]string{
	"github-actions-workflows": "Github Actions",
	"jfrog-pipelines":          "Jfrog Pipelines",
}

var SCM_PLATFORMS = map[string]string{
	"github": "Github",
	"gitlab": "Gitlab",
}

const colorRed = "\033[31m"
const colorReset = "\033[0m"
const colorBlue = "\033[34m"

func PrintResults(ruleResults map[int]*rulesConfig.RuleResult, summary rulesConfig.OutputSummary, outputFormat string) error {
	if outputFormat == "" {
		printPretty(ruleResults, summary)
		printSummary(ruleResults, summary)
	} else if outputFormat == "csv" {
		return printCSV(ruleResults, summary)
	}

	return nil
}

func printPretty(ruleResults map[int]*rulesConfig.RuleResult, summary rulesConfig.OutputSummary) {
	ruleIds := sortRulesOrder(ruleResults)

	headingColored := color.New(color.FgCyan, color.Bold, color.Underline)
	headingColored.Println("Allero Pipelines Validation Results")
	fmt.Println()

	for _, id := range ruleIds {
		ruleResult := ruleResults[id]
		fmt.Println("Rule:", ruleResult.RuleName)

		if ruleResult.Valid {
			fmt.Println("No errors found")
		} else {
			failureMessageColored := color.New(color.FgRed)
			failureMessageColored.Println("Failure Message:", ruleResult.FailureMessage)
			t := table.NewWriter()

			t.SetStyle(table.StyleBold)
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"SCM Platform", "CICD Platform", "Owner Name", "Repository Name", "Pipeline Relative Path"})
			for _, schemaError := range ruleResult.SchemaErrors {
				uneascapedRepoName := unescapeValue(schemaError.RepositryName)
				uneascapedFilepath := unescapeValue(schemaError.WorkflowRelPath)

				t.AppendRow([]interface{}{SCM_PLATFORMS[schemaError.ScmPlatform], CICD_PLATFORMS[schemaError.CiCdPlatform], schemaError.OwnerName, uneascapedRepoName, uneascapedFilepath})
				t.AppendSeparator()
			}
			t.Render()
		}

		fmt.Printf("\n\n\n")

	}
}

func sortRulesOrder(ruleResults map[int]*rulesConfig.RuleResult) []int {
	var keys []int
	for k := range ruleResults {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	return keys
}

func unescapeValue(value string) string {
	return strings.ReplaceAll(value, "[ESCAPED_DOT]", ".")
}

func printSummary(ruleResults map[int]*rulesConfig.RuleResult, summary rulesConfig.OutputSummary) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	fmt.Println("Summary")

	t.AppendRow([]interface{}{"Owners", summary.TotalOwners})
	t.AppendSeparator()
	t.AppendRow([]interface{}{"Repositories", summary.TotalRepositories})
	t.AppendSeparator()
	t.AppendRow([]interface{}{"Pipelines", summary.TotalPipelines})
	t.AppendSeparator()
	t.AppendRow([]interface{}{"Total rules evaluated", summary.TotalRulesEvaluated})
	t.AppendSeparator()
	t.AppendRow([]interface{}{"Failed rules", summary.TotalFailedRules})
	t.Render()

	if summary.URL != "" {
		fmt.Println()
		fmt.Println("Select your own rules:", string(colorBlue), summary.URL, string(colorReset))
	}

	if summary.TotalFailedRules > 0 {
		fmt.Println()
		fmt.Println("Failed rules summary:")
		for _, ruleResult := range ruleResults {
			if !ruleResult.Valid {
				fmt.Println(string(colorRed), "\r", ruleResult.RuleName, string(colorReset))
			}
		}
	}
}
