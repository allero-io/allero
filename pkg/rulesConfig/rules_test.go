package rulesConfig

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	"github.com/allero-io/allero/pkg/fileManager"
	"github.com/allero-io/allero/pkg/rulesConfig/defaultRules"
	"gopkg.in/yaml.v2"
)

type TestFilesByRuleId = map[int]*FailAndPassTests

type FailAndPassTests struct {
	fails  []*FileWithName
	passes []*FileWithName
}

type FileWithName struct {
	name    string
	content interface{}
}

func ReadTestFileContent(filename string) (interface{}, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var data interface{}
	err = yaml.Unmarshal(buf, &data)
	return data, err
}

func getFileData(filename string) (int, bool) {
	parts := strings.Split(filename, "-")
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(err)
	}

	isPass := strings.Contains(parts[1], "pass")
	return id, isPass
}

func SetGithubData(rc *RulesConfig, filename string) {
	currDir, _ := os.Getwd()
	fullFilePath := filepath.Join(currDir, "tests", "github", filename)
	rc.githubData["dummy"].Repositories = make(map[string]*githubConnector.GithubRepository)
	rc.githubData["dummy"].Repositories["dummy"] = new(githubConnector.GithubRepository)
	rc.githubData["dummy"].Repositories["dummy"].GithubActionsWorkflows = make(map[string]*githubConnector.PipelineFile)
	rc.githubData["dummy"].Repositories["dummy"].GithubActionsWorkflows["dummy"] = new(githubConnector.PipelineFile)
	rc.githubData["dummy"].Repositories["dummy"].GithubActionsWorkflows["dummy"].Content, _ = ReadTestFileContent(fullFilePath)
	println(rc.githubData["dummy"].Repositories["dummy"].GithubActionsWorkflows["dummy"].Content)
}

func validatePassing(t *testing.T, rule *defaultRules.Rule,
	ruleName string, files []*FileWithName, shouldPass bool, rc *RulesConfig, scmPlatfrom string) {
	for _, file := range files {
		switch scmPlatfrom {
		case "github":
			SetGithubData(rc, file.name)
		}
		// mSchema, _ := json.Marshal(ruleSchema)
		// schemaResult, _ := jsonschemaValidator.Validate(mSchema, rc.githubData)
		schemaResult, _ := rc.Validate(ruleName, rule, scmPlatfrom)
		println(schemaResult)
	}
}
func getTestFilesByRuleId(t *testing.T, scmPlatfrom string) TestFilesByRuleId {
	currDir, _ := os.Getwd()
	testDirPath := filepath.Join(currDir, "tests", scmPlatfrom)
	files := fileManager.ReadFolder(testDirPath)

	testFilesByRuleId := make(TestFilesByRuleId)
	for _, file := range files {
		fileContent, err := ReadTestFileContent(filepath.Join(testDirPath, file.Name()))
		if err != nil {
			panic(err)
		}
		id, isPass := getFileData(file.Name())
		if testFilesByRuleId[id] == nil {
			testFilesByRuleId[id] = &FailAndPassTests{}
		}

		fileWithName := &FileWithName{name: file.Name(), content: fileContent}
		if isPass {
			testFilesByRuleId[id].passes = append(testFilesByRuleId[id].passes, fileWithName)
		} else {
			testFilesByRuleId[id].fails = append(testFilesByRuleId[id].fails, fileWithName)
		}
	}

	return testFilesByRuleId
}

func TestDefaultRulesValidation(t *testing.T) {
	var rc RulesConfig
	rc.githubData = make(map[string]*githubConnector.GithubOwner)
	rc.githubData["dummy"] = new(githubConnector.GithubOwner)
	err := rc.Initialize()
	if err != nil {
		panic(err)
	}
	scmPlatfrom := "github"
	ruleNames := rc.GetAllRuleNames(scmPlatfrom)
	testFilesByRuleId := getTestFilesByRuleId(t, scmPlatfrom)
	for _, ruleName := range ruleNames {
		println(ruleName)
		rule, err := rc.GetRule(ruleName, scmPlatfrom)
		if err != nil {
			panic(err)
		}

		validatePassing(t, rule, ruleName, testFilesByRuleId[rule.UniqueId].passes, true, &rc, scmPlatfrom)
		validatePassing(t, rule, ruleName, testFilesByRuleId[rule.UniqueId].fails, false, &rc, scmPlatfrom)
	}
}
