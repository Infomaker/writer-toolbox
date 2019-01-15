package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"net/http"
)

// Config structs
type ConfigContainer struct {
	IssuesUrl               string   `json:"issuesUrl"`
	Jql                     string   `json:"jql"`
	Fields                  []string `json:"fields"`
	HighlightField          string   `json:"highlightField"`
	HeadsUpField            string   `json:"headsUpField"`
	MaxResults              int64    `json:"maxResults"`
	Username                string   `json:"username"`
	Password                string   `json:"password"`
	IssueTypesSortOrder     []string `json:"issueTypesSortOrder"`
	ReleaseDescriptionLabel string   `json:"releaseDescriptionLabel"`
}

// Issues imports structs
type IssuesJson struct {
	StartAt   int64    `json:"startAt"`
	MaxResult int64    `json:"maxResults"`
	Total     int64    `json:"total"`
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
	ReleaseDate         string
	IssueTypes          map[string][]Issues
	Dependencies        []Dependencies
	IssueTypesSortOrder []string
	ReleaseDescription  string
}

type Dependencies struct {
	Dependency string   `json:"dependency"`
	Versions   []string `json:"versions"`
}

func GenerateReleaseNotes(configData []byte, templateFile, version, releaseDate, dependenciesData string) {
	var dependencies []Dependencies
	var config ConfigContainer

	configInfo, err2 := template.New("hello").Parse(string(configData))
	assertError(err2)
	processedConfig := new(bytes.Buffer)

	type ConfigData struct {
		Version  string
		Login    string
		Password string
	}

	err3 := configInfo.Execute(processedConfig, ConfigData{Version: version, Login: login, Password: password})
	assertError(err3)

	err := json.Unmarshal(processedConfig.Bytes(), &config)
	assertError(err)
	issues := getIssuesFromUrl(config)

	if dependenciesData != "" {
		err = json.Unmarshal([]byte(dependenciesData), &dependencies)
		assertError(err)
	}

	var output = OutputData{
		Version:      version,
		ReleaseDate:  releaseDate,
		IssueTypes:   make(map[string][]Issues),
		Dependencies: dependencies,
	}

	if len(config.IssueTypesSortOrder) == 0 {
		config.IssueTypesSortOrder = getDefaultSortOrder()
	}

	for i := 0; i < len(issues); i++ {
		issue := issues[i]
		fields := issue.Fields.(map[string]interface{})

		if fields["issuetype"] != nil {
			issueType := fields["issuetype"].(map[string]interface{})

			var typeKey = ""
			if issueType["name"] != nil {
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

		if fields["labels"] != nil {
			if config.ReleaseDescriptionLabel != "" {
				labels := fields["labels"].([]interface{})
				for j := 0; j < len(labels); j++ {
					if labels[j] == config.ReleaseDescriptionLabel {
						if fields["description"] != nil {
							output.ReleaseDescription = fields["description"].(string)
						}
					}
				}
			}
		}
	}

	output.IssueTypesSortOrder = config.IssueTypesSortOrder
	reportTemplate, err := template.New("report").Parse(templateFile)
	assertError(err)

	var out bytes.Buffer
	err = reportTemplate.Execute(&out, output)
	fmt.Println(html.UnescapeString(out.String()))
	assertError(err)
}

func getDefaultSortOrder() []string {
	return []string{"New Feature", "Task", "Bug", "Epic"}
}

func getIssuesFromUrl(config ConfigContainer) []Issues {
	type JsonData struct {
		Jql        string   `json:"jql"`
		Fields     []string `json:"fields"`
		MaxResults int64    `json:"maxResults"`
	}

	var data = JsonData{
		Jql:        config.Jql,
		Fields:     config.Fields,
		MaxResults: config.MaxResults,
	}

	body, err := json.Marshal(data)
	assertError(err)

	req, err := http.NewRequest("POST", config.IssuesUrl, bytes.NewBuffer(body))
	assertError(err)

	req.SetBasicAuth(config.Username, config.Password)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	assertError(err)
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	assertError(err)

	var issues IssuesJson
	err = json.Unmarshal(responseBody, &issues)

	return issues.Issues
}
