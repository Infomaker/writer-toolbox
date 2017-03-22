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
	Label    string `json:"label"`
	Services []ServiceItem `json:"services"`
	Lambdas  []string `json:"lambdas"`
	Other    []OtherItem `json:"other"`
}

type OtherItem struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type CredentialsItem struct {
	Profile   string `json:"profile"`
	AwsKey    string `json:"awsKey"`
	AwsSecret string `json:"awsSecret"`
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

func GenerateReport(jsonData []byte, templateFile string) {
	var config Installations

	err := json.Unmarshal(jsonData, &config)

	if err != nil {
		errState(err.Error())
	}

	output := Output{}

	for i := 0; i < len(config.InstallationItems); i++ {
		installation := config.InstallationItems[i];

		outputTemplate := OutputTemplate{
			Label : installation.Label,
		}

		for j := 0; j < len(installation.Services); j++ {

			service := installation.Services[j]
			clusterArn := GetClusterArn(service.Cluster, nil);
			serviceArn := GetServiceArn(clusterArn, service.Service, nil);

			serviceDescription := _describeService(clusterArn, serviceArn, nil)

			for k := 0; k < len(serviceDescription.Services); k++ {

				realService := serviceDescription.Services[k];

				taskDefinition := _describeTaskDefinition(*realService.TaskDefinition, nil);
				version, image := ExtractVersion(*taskDefinition.TaskDefinition.ContainerDefinitions[0].Image)

				for l := 0; l < len(realService.Deployments); l++ {
					deployment := realService.Deployments[l];
					url := service.Url;

					outputItem := OutputItem{
						Version: version,
						Image:ExtractImageName(image),
						TaskDefName:ExtractName(deployment.TaskDefinition),
						Label:service.Label,
						RunningCount:*deployment.RunningCount,
						DesiredCount:*deployment.DesiredCount,
						Url:url,
					}
					outputTemplate.Services = append(outputTemplate.Services, outputItem)
				}
			}
		}

		for k := 0; k < len(installation.Lambdas); k++ {
			lambdaFunction := installation.Lambdas[k]

			aliasInfo := _getLambdaFunctionAliasInfo(lambdaFunction, "PRIMARY")

			functionInfo := _getLambdaFunctionInfo(lambdaFunction, *aliasInfo.FunctionVersion);

			outputItem := LambdaOutputItem{
				Description: *functionInfo.Description,
				Version: *functionInfo.Version,
				Label: lambdaFunction,
			}

			outputTemplate.Lambdas = append(outputTemplate.Lambdas, outputItem)
		}

		for k := 0; k < len(installation.Other); k++ {
			item := installation.Other[k]
			outputItem := OtherOutputItem{
				Label: item.Name,
				Url: item.Url,
			}
			outputTemplate.Others = append(outputTemplate.Others, outputItem)
		}

		output.Installations = append(output.Installations, outputTemplate)
	}

	reportTemplate, err := template.New("report").Parse(templateFile)

	if err != nil {
		errState(err.Error())
	}

	err = reportTemplate.Execute(os.Stdout, output)

	if err != nil {
		errState(err.Error())
	}
}

