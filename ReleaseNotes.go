package main

import (
	"encoding/json"
	"html/template"
	"os"
)

type IssuesJson struct {
	Issues []Issues `json:"issues"`
}

type Issues struct {
	Key    string
	Self   string
	Fields Fields `json:"fields"`
}

type Fields interface {
}

type OutputData struct {
	Version      string
	IssueTypes   map[string][]Issues
	Dependencies []Dependencies
}

type Dependencies struct {
	Dependency string `json:"dependency"`
	Versions   []string `json:"versions"`
}

func GenerateReleaseNotes(jsonData []byte, templateFile, version, dependenciesData string) {

	var issues IssuesJson
	var dependencies []Dependencies

	err := json.Unmarshal(jsonData, &issues)
	assertError(err);

	if (dependenciesData != "") {
		err = json.Unmarshal([]byte(dependenciesData), &dependencies)
		assertError(err);
	}

	var output = OutputData{
		Version: version,
		IssueTypes: make(map[string][]Issues),
		Dependencies: dependencies,
	}

	for i := 0; i < len(issues.Issues); i++ {
		issue := issues.Issues[i]
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

	reportTemplate, err := template.New("report").Parse(templateFile)

	assertError(err);

	err = reportTemplate.Execute(os.Stdout, output)

	assertError(err);


}
