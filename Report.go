package main

import (
	"encoding/json"
	"html/template"
	"os"
)

type Installations struct {
	InstallationItems []installationItem `json:"installations"`
}

type installationItem struct {
	Label    string        `json:"label"`
	Services []ServiceItem `json:"services"`
	Lambdas  []string      `json:"lambdas"`
	Other    []OtherItem   `json:"other"`
	Info     []InfoItem    `json:"info"`
}

type OtherItem struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type InfoItem struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ServiceItem struct {
	Label   string `json:"label"`
	Cluster string `json:"cluster"`
	Service string `json:"service"`
	Url     string `json:"url"`
}

type Output struct {
	Installations []OutputTemplate
}

type OutputTemplate struct {
	Label    string
	Services []OutputItem
	Lambdas  []LambdaOutputItem
	Others   []OtherOutputItem
	Info     []InfoOutputItem
}

type OutputItem struct {
	Label        string
	TaskDefName  string
	Image        string
	Version      string
	DesiredCount int64
	RunningCount int64
	Url          string
}

type LambdaOutputItem struct {
	Label       string
	Version     string
	Description string
}

type OtherOutputItem struct {
	Label string
	Url   string
}

type InfoOutputItem struct {
	Name  string
	Value string
}

func GenerateReport(jsonData []byte, templateFile string) {
	var config Installations

	err := json.Unmarshal(jsonData, &config)
	assertError(err)

	output := Output{}

	for i := 0; i < len(config.InstallationItems); i++ {
		installation := config.InstallationItems[i]
		outputTemplate := OutputTemplate{
			Label: installation.Label,
		}

		for j := 0; j < len(installation.Services); j++ {
			service := installation.Services[j]
			clusterArn := GetClusterArn(service.Cluster, nil)
			serviceArn := GetServiceArn(clusterArn, service.Service, nil)
			serviceDescription := describeService(clusterArn, serviceArn, nil)

			for k := 0; k < len(serviceDescription.Services); k++ {
				realService := serviceDescription.Services[k]
				taskDefinition := describeTaskDefinition(*realService.TaskDefinition, nil)

				for l := 0; l < len(taskDefinition.TaskDefinition.ContainerDefinitions); l++ {
					version, image := ExtractVersion(*taskDefinition.TaskDefinition.ContainerDefinitions[l].Image)

					for n := 0; n < len(realService.Deployments); n++ {
						deployment := realService.Deployments[n]
						url := service.Url

						outputItem := OutputItem{
							Version:      version,
							Image:        ExtractImageName(image),
							TaskDefName:  ExtractName(deployment.TaskDefinition),
							Label:        service.Label,
							RunningCount: *deployment.RunningCount,
							DesiredCount: *deployment.DesiredCount,
							Url:          url,
						}
						outputTemplate.Services = append(outputTemplate.Services, outputItem)
					}
				}
			}
		}

		for k := 0; k < len(installation.Lambdas); k++ {
			lambdaFunction := installation.Lambdas[k]
			aliasInfo := getLambdaFunctionAliasInfo(lambdaFunction, "PRIMARY")
			functionInfo := getLambdaFunctionInfo(lambdaFunction, *aliasInfo.FunctionVersion)

			outputItem := LambdaOutputItem{
				Description: *functionInfo.Description,
				Version:     *functionInfo.Version,
				Label:       lambdaFunction,
			}
			outputTemplate.Lambdas = append(outputTemplate.Lambdas, outputItem)
		}

		for k := 0; k < len(installation.Other); k++ {
			item := installation.Other[k]
			outputItem := OtherOutputItem{
				Label: item.Name,
				Url:   item.Url,
			}
			outputTemplate.Others = append(outputTemplate.Others, outputItem)
		}

		for k := 0; k < len(installation.Info); k++ {
			item := installation.Info[k]
			infoItem := InfoOutputItem{
				Name:  item.Name,
				Value: item.Value,
			}
			outputTemplate.Info = append(outputTemplate.Info, infoItem)
		}

		output.Installations = append(output.Installations, outputTemplate)
	}

	reportTemplate, err := template.New("report").Parse(templateFile)

	assertError(err)
	err = reportTemplate.Execute(os.Stdout, output)

	assertError(err)
}
