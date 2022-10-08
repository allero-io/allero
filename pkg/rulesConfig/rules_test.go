package rulesConfig

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	githubConnector "github.com/allero-io/allero/pkg/connectors/github"
	"github.com/allero-io/allero/pkg/fileManager"
	"github.com/allero-io/allero/pkg/rulesConfig/defaultRules"
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

func ReadTestFileContent(filename string) (map[string]*githubConnector.GithubOwner, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	data := make(map[string]*githubConnector.GithubOwner)
	err = json.Unmarshal([]byte(buf), &data)
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
	rc.githubData, _ = ReadTestFileContent(fullFilePath)
}

func validatePassing(t *testing.T, rule *defaultRules.Rule,
	ruleName string, files []*FileWithName, shouldPass bool, rc *RulesConfig, scmPlatfrom string) {
	for _, file := range files {
		// TODO: add gitlab support
		switch scmPlatfrom {
		case "github":
			SetGithubData(rc, file.name)
		}
		schemaResult, _ := rc.Validate(ruleName, rule, scmPlatfrom)
		if len(schemaResult) > 0 && shouldPass {
			t.Errorf("Expected validation for rule name %s to pass, but it failed for file %s\n", ruleName, file.name)
		}
		if len(schemaResult) == 0 && !shouldPass {
			t.Errorf("Expected validation for rule name %s to pass, but it passed for file %s\n", ruleName, file.name)
		}
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
		// test file will be looking like 1-pass.json where 1 is the rule id to test
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
	// githubData have to be initialize to be able to init RuleConfig
	rc.githubData = make(map[string]*githubConnector.GithubOwner)
	rc.githubData["dummy"] = new(githubConnector.GithubOwner)
	err := rc.Initialize()
	if err != nil {
		panic(err)
	}
	for _, scmPlatform := range []string{"github", "gitlab"} {
		ruleNames := rc.GetAllRuleNames(scmPlatform)
		testFilesByRuleId := getTestFilesByRuleId(t, scmPlatform)
		// go over all rules
		for _, ruleName := range ruleNames {
			rule, err := rc.GetRule(ruleName, scmPlatform)
			if err != nil {
				panic(err)
			}
			testFileByRuleId := testFilesByRuleId[rule.UniqueId]
			if testFileByRuleId != nil {
				println("test - " + ruleName)
				validatePassing(t, rule, ruleName, testFileByRuleId.passes, true, &rc, scmPlatform)
				validatePassing(t, rule, ruleName, testFileByRuleId.fails, false, &rc, scmPlatform)
			}
		}
	}
}
