package resultsPrinter

import (
	"fmt"

	"github.com/allero-io/allero/pkg/rulesConfig"
	"github.com/gocarina/gocsv"
)

type OutputCsv struct {
	OwnerName       string `csv:"owner"`
	RepositoryName  string `csv:"repository"`
	WorkflowRelPath string `csv:"workflowRelPath"`
	RuleName        string `csv:"ruleName"`
	FailureMessage  string `csv:"failureMessage"`
}

func printCSV(ruleResults []*rulesConfig.RuleResult, summary rulesConfig.OutputSummary) error {
	ruleResultsCsv := []*OutputCsv{}
	for _, ruleResult := range ruleResults {
		if !ruleResult.Valid {
			for _, schemaError := range ruleResult.SchemaErrors {
				ruleResultsCsv = append(ruleResultsCsv, &OutputCsv{
					OwnerName:       schemaError.OwnerName,
					RepositoryName:  schemaError.RepositryName,
					WorkflowRelPath: schemaError.WorkflowRelPath,
					RuleName:        ruleResult.RuleName,
					FailureMessage:  ruleResult.FailureMessage,
				})
			}
		}
	}

	csvContent, err := gocsv.MarshalString(&ruleResultsCsv)
	if err != nil {
		return err
	}

	fmt.Println(csvContent)
	return nil
}
