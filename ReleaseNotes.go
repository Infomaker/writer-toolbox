package main

import (
	"encoding/json"
	"html/template"
	"os"
	"net/http"
	"bytes"
	"io/ioutil"
)

// Config structs

type ConfigContainer struct {
	IssuesUrl           string `json:"issuesUrl"`
	Jql                 string `json:"jql"`
	Fields              []string `json:"fields"`
	HighlightField      string `json:"highlightField"`
	HeadsUpField        string `json:"headsUpField"`
	MaxResults          int64 `json:"maxResults"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	IssueTypesSortOrder []string `json:"issueTypesSortOrder"`
}

// Issues imports structs

type IssuesJson struct {
	StartAt   int64 `json:"startAt"`
	MaxResult int64 `json:"maxResults"`
	Total     int64 `json:"total"`
	Issues    []Issues `json:"issues"`
}

// -- Issues structs

type Issues struct {
	Key    string
	Self   string
	Fields Fields `json:"fields"`
}

type Fields interface {
}

// --- Output structs
type OutputData struct {
	Version             string
	IssueTypes          map[string][]Issues
	Dependencies        []Dependencies
	IssueTypesSortOrder []string
}

type Dependencies struct {
	Dependency string `json:"dependency"`
	Versions   []string `json:"versions"`
}

func GenerateReleaseNotes(configData []byte, templateFile, version, dependenciesData string) {

	var dependencies []Dependencies
	var config ConfigContainer

	err := json.Unmarshal(configData, &config)
	assertError(err);

	issues := getIssuesFromUrl(config)

	if (dependenciesData != "") {
		err = json.Unmarshal([]byte(dependenciesData), &dependencies)
		assertError(err);
	}

	var output = OutputData{
		Version: version,
		IssueTypes: make(map[string][]Issues),
		Dependencies: dependencies,
	}

	if (len(config.IssueTypesSortOrder) == 0) {
		config.IssueTypesSortOrder = getDefaultSortOrder()
	}

	for i := 0; i < len(issues); i++ {
		issue := issues[i]
		fields := issue.Fields.(map[string]interface{})

		if (fields["issuetype"] != nil) {
			issueType := fields["issuetype"].(map[string]interface{})

			var typeKey = "";
			if (issueType["name"] != nil) {
				typeKey = string(issueType["name"].(string))
			}
			if output.IssueTypes[typeKey] == nil {
				output.IssueTypes[typeKey] = []Issues{issue}
			} else {
				output.IssueTypes[typeKey] = append(output.IssueTypes[typeKey], issue)
			}
		} else {
			output.IssueTypes[""] = []Issues{issue}
		}
	}

	output.IssueTypesSortOrder = config.IssueTypesSortOrder

	reportTemplate, err := template.New("report").Parse(templateFile)
	assertError(err);

	err = reportTemplate.Execute(os.Stdout, output)
	assertError(err);
}
func getDefaultSortOrder() []string {
	return []string{"New Feature", "Task", "Bug", "Epic"}
}

func getIssuesFromUrl(config ConfigContainer) []Issues {

	type JsonData struct {
		Jql        string `json:"jql"`
		Fields     []string `json:"fields"`
		MaxResults int64 `json:"maxResults"`
	}

	var data = JsonData{
		Jql: config.Jql,
		Fields: config.Fields,
		MaxResults: config.MaxResults,
	}

	body, err := json.Marshal(data)
	assertError(err)

	req, err := http.NewRequest("POST", config.IssuesUrl, bytes.NewBuffer(body))
	assertError(err)

	req.SetBasicAuth(config.Username, config.Password)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	assertError(err);

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	assertError(err)

	var issues IssuesJson

	err = json.Unmarshal(responseBody, &issues)

	return issues.Issues
}