package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	html_report "github.com/daveshanley/vacuum/html-report"
	"github.com/daveshanley/vacuum/model"
	"github.com/daveshanley/vacuum/model/reports"
	"github.com/daveshanley/vacuum/motor"
	"github.com/daveshanley/vacuum/rulesets"
	"github.com/daveshanley/vacuum/statistics"
	"github.com/pb33f/libopenapi/datamodel"
	"github.com/pb33f/libopenapi/index"
	"github.com/pterm/pterm"
)

func BuildResults(rulesetFlag string, specBytes []byte) (*model.RuleResultSet, *motor.RuleSetExecutionResult, error) {

	// read spec and parse
	defaultRuleSets := rulesets.BuildDefaultRuleSets()

	// default is recommended rules, based on spectral (for now anyway)
	selectedRS := defaultRuleSets.GenerateOpenAPIRecommendedRuleSet()

	// fmt.Printf("Linting against %d rules: %s\n", len(selectedRS.Rules), selectedRS.DocumentationURI)

	ruleset := motor.ApplyRulesToRuleSet(&motor.RuleSetExecution{
		RuleSet: selectedRS,
		Spec:    specBytes,
		// CustomFunctions: customFunctions,
	})

	resultSet := model.NewRuleResultSet(ruleset.Results)
	resultSet.SortResultsByLineNumber()
	return resultSet, ruleset, nil
}

func GenerateHtml(url string) {

	response, _err := http.Get(url)
	if _err != nil {
		panic(_err.Error())
	}
	// convert repsonse body to byte array
	specBytes, _err := ioutil.ReadAll(response.Body)
	if _err != nil {
		panic(_err.Error())
	}

	reportOutput := "vacuum-reports/"
	// create the report output directory if it doesn't exist
	if _, err := os.Stat(reportOutput); os.IsNotExist(err) {
		os.Mkdir(reportOutput, 0755)
	}
	// get the repo name from the url by splitting between /bundesAPI/ and /main
	repoName := strings.Split(strings.Split(url, "/bundesAPI/")[1], "/main")[0]

	reportOutput = reportOutput + repoName + ".html"

	var err error

	var resultSet *model.RuleResultSet
	var ruleset *motor.RuleSetExecutionResult
	var specIndex *index.SpecIndex
	var specInfo *datamodel.SpecInfo
	var stats *reports.ReportStatistics

	resultSet, ruleset, err = BuildResults("", specBytes)
	if err != nil {
		pterm.Error.Printf("Failed to generate report: %v\n\n", err)
		panic(err)
	}
	specIndex = ruleset.Index
	specInfo = ruleset.SpecInfo

	specInfo.Generated = time.Now()
	stats = statistics.CreateReportStatistics(specIndex, specInfo, resultSet)

	// generate html report
	report := html_report.NewHTMLReport(specIndex, specInfo, resultSet, stats)

	generatedBytes := report.GenerateReport(false)

	err = os.WriteFile(reportOutput, generatedBytes, 0664)

	if err != nil {
		pterm.Error.Printf("Unable to write HTML report file: '%s': %s\n", reportOutput, err.Error())
		pterm.Println()
		panic(err)
	}

	pterm.Success.Printf("HTML Report generated for '%s', written to '%s'\n", repoName, reportOutput)
	pterm.Println()

}
