package resultsPrinter

import (
	"fmt"
	"os"

	"github.com/allero-io/allero/pkg/rulesConfig"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
)

func PrintResults(ruleResults []*rulesConfig.RuleResult, summary rulesConfig.OutputSummary, outputFormat string) error {
	if outputFormat == "" {
		printPretty(ruleResults, summary)
		printSummary(summary)
	} else if outputFormat == "csv" {
		return printCSV(ruleResults, summary)
	}

	return nil
}

func printPretty(ruleResults []*rulesConfig.RuleResult, summary rulesConfig.OutputSummary) {
	headingColored := color.New(color.FgCyan, color.Bold, color.Underline)
	headingColored.Println("Allero Pipelines Validation Results")
	fmt.Println()

	for _, ruleResult := range ruleResults {
		fmt.Println("Rule:", ruleResult.RuleName)

		if ruleResult.Valid {
			fmt.Println("No errors found")
		} else {
			failureMessageColored := color.New(color.FgRed)
			failureMessageColored.Println("Failure Message:", ruleResult.FailureMessage)
			t := table.NewWriter()

			t.SetStyle(table.StyleBold)
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"#", "SCM Platform", "CICD Platform", "Owner Name", "Repository Name", "Pipeline Relative Path"})
			for _, schemaError := range ruleResult.SchemaErrors {
				t.AppendRow([]interface{}{"", "Github", "Github Actions", schemaError.OwnerName, schemaError.RepositryName, schemaError.WorkflowRelPath})
				t.AppendSeparator()
			}
			t.Render()
		}

		fmt.Printf("\n\n\n")

	}

	printSummary(summary)
}

func printSummary(summary rulesConfig.OutputSummary) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	fmt.Println("SUMMARY")

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
}
