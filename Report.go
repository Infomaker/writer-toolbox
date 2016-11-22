package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
)

type Installations struct {
	InstallationItems []installationItem `json:"installations"`
}

type installationItem struct {
	Label       string `json:"label"`
	Services    []ServiceItem `json:"services"`
	Lambdas     []string `json:"lambdas"`
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
}

type Output struct {
	Installations []OutputTemplate
}

type OutputTemplate struct {
	Label    string
	Services []OutputItem
	Lambdas []LambdaOutputItem
}

type OutputItem struct {
	Label        string
	TaskDefName  string
	Image        string
	Version      string
	DesiredCount int64
	RunningCount int64
}

type LambdaOutputItem struct {
	Label string
	Version string
}

func GenerateReport(jsonData []byte, templateFile string) {
	var config Installations

	err := json.Unmarshal(jsonData, &config)

	if err != nil {
		fmt.Println(err.Error())
	}

	output := Output{}

	for i := 0; i < len(config.InstallationItems); i++ {
		installation := config.InstallationItems[i];

		outputTemplate := OutputTemplate{
			Label : installation.Label,
		}

		for j := 0; j < len(installation.Services); j++ {

			service := installation.Services[j]
			clusterArn := GetClusterArn(service.Cluster);
			serviceArn := GetServiceArn(clusterArn, service.Service);

			serviceDescription := _describeService(clusterArn, serviceArn)

			for k := 0; k < len(serviceDescription.Services); k++ {

				realService := serviceDescription.Services[k];

				taskDefinition := _describeTaskDefinition(*realService.TaskDefinition);
				version, image := ExtractVersion(*taskDefinition.TaskDefinition.ContainerDefinitions[0].Image)



				for l := 0; l < len(realService.Deployments); l++ {
					deployment := realService.Deployments[l];
					outputItem := OutputItem{
						Version: version,
						Image:ExtractImageName(image),
						TaskDefName:ExtractName(deployment.TaskDefinition),
						Label:service.Label,
						RunningCount:*deployment.RunningCount,
						DesiredCount:*deployment.DesiredCount,
					}
					outputTemplate.Services = append(outputTemplate.Services, outputItem)
				}
			}

		}

		for k := 0; k < len(installation.Lambdas); k++ {
			lambdaFunction := installation.Lambdas[k]

			lambdaInfo := _getLambdaFunctionAliasInfo(lambdaFunction, "PRIMARY")

			outputItem := LambdaOutputItem{
				Version: *lambdaInfo.FunctionVersion,
				Label: lambdaFunction,
			}

			outputTemplate.Lambdas = append(outputTemplate.Lambdas, outputItem)
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

