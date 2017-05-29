package main

import (
	"regexp"

	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"time"
	"encoding/json"
	"errors"
	"strconv"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

func ExtractName(a *string) string {
	re := regexp.MustCompile("\\w+:\\w+:\\w+:[\\w-]+:\\d+:[^/]+\\/(.+)")

	res := re.FindAllStringSubmatch(*a, -1)

	if res == nil {
		return "No match for " + *a
	}

	return res[0][1]
}

func ExtractImageName(a string) string {
	re := regexp.MustCompile("[^/]+/(.+)")

	res := re.FindAllStringSubmatch(a, -1)

	if res == nil {
		return "No match for " + a
	}

	return res[0][1]
}

func ExtractVersion(a string) (string, string) {
	re := regexp.MustCompile("(.*?)\\:(.+)")

	res := re.FindAllStringSubmatch(a, -1)

	if res == nil {
		return "", ""
	}

	return res[0][2], res[0][1]
}

func _listServices(cluster string, svc *ecs.ECS) *ecs.ListServicesOutput {
	if svc == nil {
		svc = ecs.New(session.New(), _getAwsConfig())
	}

	params := &ecs.ListServicesInput{
		Cluster:    aws.String(cluster),
	}

	resp, err := svc.ListServices(params)
	assertError(err);

	return resp
}

func _listTasks(cluster, service string, svc *ecs.ECS) (*ecs.ListTasksOutput, error) {
	if svc == nil {
		svc = ecs.New(session.New(), _getAwsConfig())
	}

	params := &ecs.ListTasksInput{
		Cluster:     aws.String(cluster),
		ServiceName: aws.String(service),
	}

	resp, err := svc.ListTasks(params)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func _describeService(clusterArn, serviceArn string, svc *ecs.ECS) *ecs.DescribeServicesOutput {
	if svc == nil {
		svc = ecs.New(session.New(), _getAwsConfig())
	}

	params := &ecs.DescribeServicesInput{
		Cluster:  aws.String(clusterArn),
		Services: []*string{aws.String(serviceArn)},
	}

	result, err := svc.DescribeServices(params)
	assertError(err);

	return result
}

func _describeTasks(clusterArn string, tasks []*string, svc *ecs.ECS) (*ecs.DescribeTasksOutput, error) {
	if (svc == nil) {
		svc = ecs.New(session.New(), _getAwsConfig())
	}

	params := &ecs.DescribeTasksInput{
		Cluster: aws.String(clusterArn),
		Tasks:   tasks,
	}

	result, err := svc.DescribeTasks(params)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func _createTaskDefinition(taskDefinition *ecs.DescribeTaskDefinitionOutput, svc *ecs.ECS) string {
	if (svc == nil) {
		svc = ecs.New(session.New(), _getAwsConfig())
	}

	params := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: taskDefinition.TaskDefinition.ContainerDefinitions,
		Family:               taskDefinition.TaskDefinition.Family,
		Volumes:              taskDefinition.TaskDefinition.Volumes,
	}

	registrationResult, err := svc.RegisterTaskDefinition(params)

	assertError(err);

	taskDefinitionArn := registrationResult.TaskDefinition.TaskDefinitionArn

	return *taskDefinitionArn
}

func _updateTaskDefinitionForService(newTaskDefinitionArn string, service *ecs.DescribeServicesOutput, svc *ecs.ECS) {
	if (svc == nil) {
		svc = ecs.New(session.New(), _getAwsConfig())
	}

	clusterArn := service.Services[0].ClusterArn
	serviceArn := service.Services[0].ServiceArn

	params := &ecs.UpdateServiceInput{
		Cluster:        aws.String(*clusterArn),
		DesiredCount:   aws.Int64(*service.Services[0].DesiredCount),
		Service:        aws.String(*serviceArn),
		TaskDefinition: aws.String(newTaskDefinitionArn),
	}

	_, err := svc.UpdateService(params)

	assertError(err);

	_waitForUpdatedTaskDefinition(*clusterArn, *serviceArn, svc)
}

func _waitForUpdatedTaskDefinition(cluster string, service string, svc *ecs.ECS) error {
	newTask := ""

	attempts := 240
	sleepTime := 2

	for i := 0; i < attempts; i++ {

		currentService := _describeService(cluster, service, svc);

		for j := 0; j < len(currentService.Services); j++ {
			item := currentService.Services[j]

			for k := 0; k < len(item.Deployments); k++ {
				deployment := item.Deployments[k]

				if (*deployment.Status == "PRIMARY" && *deployment.RunningCount == *deployment.DesiredCount) {
					return nil
				}
			}
		}

		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	return errors.New("Task " + newTask + " did not start in " + strconv.Itoa(attempts * sleepTime) + " seconds")

}

func ListServices(clusterArn string) {

	resp := _listServices(clusterArn, nil)
	for i := 0; i < len(resp.ServiceArns); i++ {
		name := ExtractName(resp.ServiceArns[i])
		fmt.Println(name)
	}
}

func ListTasks(clusterArn, serviceArn string) {
	resp, err := _listTasks(clusterArn, serviceArn, nil)
	if (err != nil) {
		errState(err.Error())
	}
	for i := 0; i < len(resp.TaskArns); i++ {
		name := ExtractName(resp.TaskArns[i])
		fmt.Println(name)
	}
}

func GetSvcForCredentials(awsKey, secretKey, profile string) *ecs.ECS {

	var tempAuth = new(Auth)

	if profile != "" {
		key, secret := getAwsCredentialsFromProfile(profile)
		tempAuth = new(Auth)
		tempAuth.key = key
		tempAuth.secret = secret
	}

	if (awsKey != "" && secretKey != "") {
		tempAuth = new(Auth)
		tempAuth.key = awsKey
		tempAuth.secret = secretKey
	}

	config := &aws.Config{Region: aws.String(region), Credentials: credentials.NewStaticCredentials(tempAuth.key, tempAuth.secret, "")}

	return ecs.New(session.New(), config)
}

func _updateService(clusterArn, serviceArn string, done chan Report, svc *ecs.ECS) (string, error) {

	tasks, err := _listTasks(clusterArn, serviceArn, svc)
	if (err != nil) {
		if (done != nil) {
			done <- Report{Message:err.Error(), Success:false}
		}
		return "", err
	}
	service := _describeService(clusterArn, serviceArn, svc)

	if len(service.Services) > 1 {
		errMessage := "No support for multiple services"
		if (done != nil) {
			done <- Report{Message: errMessage, Success:false}
		}
		return "", errors.New(errMessage)
	}

	desiredCount := *service.Services[0].DesiredCount

	if len(tasks.TaskArns) < int(desiredCount) {
		errorMessage := "The number of actual tasks " + strconv.Itoa(len(tasks.TaskArns)) + " differs from desired tasks " + strconv.FormatInt(desiredCount, 10)
		if (done != nil) {
			done <- Report{Message:errorMessage, Success:false}
		}
		return "", errors.New(errorMessage)
	}

	for i := 0; i < len(tasks.TaskArns); i++ {
		err = _stopTask(clusterArn, *tasks.TaskArns[i], svc)
		if err != nil {
			if (done != nil) {
				done <- Report{Message:err.Error(), Success:false}
			}
			return "", err
		}
		err = _waitForUpdatedTaskDefinition(clusterArn, serviceArn, svc)
		if err != nil {
			if (done != nil) {
				done <- Report{Message:err.Error(), Success:false}
			}
			return "", err
		}
	}
	message := "Service " + *service.Services[0].ServiceName + " is updated"

	if done != nil {
		done <- Report{Message:message, Success:true}
	}

	return message, nil
}

func UpdateService(clusterArn, serviceArn string) {
	message, err := _updateService(clusterArn, serviceArn, nil, nil)

	assertError(err);

	fmt.Println(message)
}

type Update struct {
	Cluster       string `json:"cluster"`
	Service       string `json:"service"`
	AwsKey        string `json:"awsKey"`
	AwsSecret     string `json:"awsSecret"`
	Profile       string `json:"profile"`
	Label         string `json:"label"`
	ContainerName string `json:"containerName"`
}

type Report struct {
	Message string
	Success bool
}

func UpdateServices(data []byte) {

	var updateConfig []Update

	err := json.Unmarshal(data, &updateConfig)

	assertError(err);

	messages := make(chan Report, len(updateConfig))

	fmt.Printf("Performing update on %d services.\n", len(updateConfig))

	for i := 0; i < len(updateConfig); i++ {
		config := updateConfig[i]
		if (config.AwsKey == "" && config.Profile == "") {
			messages <- Report{Message:config.Label + ": No AwsKey or Profile specified for cluster: " + config.Cluster + ", service: " + config.Service, Success:false}
		} else {
			svc := GetSvcForCredentials(config.AwsKey, config.AwsSecret, config.Profile)
			clusterArn := GetClusterArn(config.Cluster, svc)
			serviceArn := GetServiceArn(clusterArn, config.Service, svc)
			fmt.Println(config.Label + ": Updating service " + config.Service)
			go _updateService(clusterArn, serviceArn, messages, svc)
		}
	}

	result := len(updateConfig);
	success := true
	for {
		message, more := <-messages
		if more {
			result--;
			fmt.Printf("%s, %d to go\n", message.Message, result)
			if !message.Success {
				success = false
			}
		}

		if (result < 1) {
			break;
		}
	}

	if !success {
		errState("Update failed for one or more services")
	}

}

func DescribeService(clusterArn, serviceArn string) {

	service := _describeService(clusterArn, serviceArn, nil)

	for n := 0; n < len(service.Services); n++ {
		item := service.Services[n]

		fmt.Printf("Service name [%s], Running: %d, Pending: %d, Desired: %d\n", *item.ServiceName, *item.RunningCount, *item.PendingCount, *item.DesiredCount)

		if verboseLevel == 1 {
			for i := 0; i < len(item.Deployments); i++ {
				deployment := item.Deployments[i]
				fmt.Printf("   %s (%s), running: %d, Pending: %d: Desired: %d\n", ExtractName(deployment.TaskDefinition), *deployment.Status, *deployment.RunningCount, *deployment.PendingCount, *deployment.DesiredCount)
			}
		}

		if verboseLevel == 2 {
			fmt.Printf("Deployment configuration -> MaximumPercent: %d, MinimumHealthyPercent: %d\n", *item.DeploymentConfiguration.MaximumPercent, *item.DeploymentConfiguration.MinimumHealthyPercent)
			definition := _describeTaskDefinition(*item.TaskDefinition, nil)
			fmt.Println(definition)
		}
	}
}

func ReleaseService(clusterArn, serviceArn, version string) {
	message, err := _releaseService(clusterArn, serviceArn, containerName, version, nil, nil)

	if (err != nil) {
		errState(err.Error())
	}

	fmt.Println(message)
}

func _releaseService(clusterArn, serviceArn, containerName, version string, done chan Report, svc *ecs.ECS) (string, error) {
	service := _describeService(clusterArn, serviceArn, svc)

	if len(service.Services) > 1 {
		errorMessage := *service.Services[0].ServiceName + " No support for multiple services"
		if (done != nil) {
			done <- Report{Message: errorMessage, Success:false}
		}
		return "", errors.New(errorMessage)
	}

	taskDefinitionName := *service.Services[0].TaskDefinition
	taskDefinition := _describeTaskDefinition(taskDefinitionName, svc)

	containerIndex := 0

	if (len(taskDefinition.TaskDefinition.ContainerDefinitions) > 1) {
		if containerName == "" {
			errorMessage := "Please specify containerName for service with multiple container definitions";
			if (done != nil) {
				done <- Report{Message: errorMessage, Success:false}
			}
			return "", errors.New(errorMessage)
		}

		containerIndex = getContainerIndexForName(taskDefinition.TaskDefinition.ContainerDefinitions, containerName)

		if (containerIndex == -1) {
			errorMessage := "No container named " + containerName + " found in task definition";
			if (done != nil) {
				done <- Report{Message: errorMessage, Success:false}
			}
			return "", errors.New(errorMessage)
		}
	}

	dockerImage := *taskDefinition.TaskDefinition.ContainerDefinitions[containerIndex].Image
	currentVersion, imagePart := ExtractVersion(dockerImage)

	if (verboseLevel > 0) {
		fmt.Printf("Service            [%s]\n", *service.Services[0].ServiceName)
		fmt.Printf("Task definition    [%s]\n", taskDefinitionName)
		fmt.Printf("Docker image       [%s]\n", imagePart)
		fmt.Printf("Version            [%s]\n", currentVersion)
	}

	if version == currentVersion {
		errMessage := *service.Services[0].ServiceName + " Specified version is already deployed!"
		if (done != nil) {
			done <- Report{Message: errMessage, Success:false}
		}
		return "", errors.New(errMessage)
	}

	minimumHealthyPercentage := *service.Services[0].DeploymentConfiguration.MinimumHealthyPercent
	desiredCount := *service.Services[0].DesiredCount
	minimumHealthyCount := desiredCount * minimumHealthyPercentage / 100

	if (verboseLevel > 0) {
		fmt.Printf("HealthyPercentage [%d], DesiredCount [%d] -> Number of running instances during deployment [%d] ... ", minimumHealthyPercentage, desiredCount, minimumHealthyCount)
	}

	if desiredCount - minimumHealthyCount == 0 {

		errMessage := *service.Services[0].ServiceName + " Not possible to deploy because of too high healthy percentage"
		if (done != nil) {
			done <- Report{Message: errMessage, Success:false}
		}
		return "", errors.New(errMessage)
	}

	*taskDefinition.TaskDefinition.ContainerDefinitions[containerIndex].Image = imagePart + ":" + version

	newTaskDefinitionArn := _createTaskDefinition(taskDefinition, svc)

	_updateTaskDefinitionForService(newTaskDefinitionArn, service, svc)

	message := "Service " + *service.Services[0].ServiceName + " is released with version " + version

	if done != nil {
		done <- Report{Message:message, Success:true}
	}

	return message, nil
}

func getContainerIndexForName(definitions []*ecs.ContainerDefinition, name string) int {
	for i := 0; i < len(definitions); i++ {
		definition := definitions[i]
		if (*definition.Name == name) {
			return i;
		}
	}

	return -1;
}

func ReleaseServices(version string, data []byte) {
	var updateConfig []Update

	err := json.Unmarshal(data, &updateConfig)

	assertError(err);

	messages := make(chan Report, len(updateConfig))

	fmt.Printf("Performing release to %s on %d services\n", version, len(updateConfig))

	for i := 0; i < len(updateConfig); i++ {
		config := updateConfig[i]

		if (config.AwsKey == "" && config.Profile == "") {
			messages <- Report{Message:config.Label + ": No AwsKey or Profile specified for cluster: " + config.Cluster + ", service: " + config.Service, Success:false}
		} else {
			svc := GetSvcForCredentials(config.AwsKey, config.AwsSecret, config.Profile)
			clusterArn := GetClusterArn(config.Cluster, svc)
			serviceArn := GetServiceArn(clusterArn, config.Service, svc)
			localContainerName := containerName
			if (config.ContainerName != "") {
				localContainerName = config.ContainerName
			}
			fmt.Println(config.Label + ": Releasing service " + config.Service + ", containerName: " + localContainerName)
			go _releaseService(clusterArn, serviceArn, localContainerName, version, messages, svc)
		}
	}

	result := len(updateConfig);
	success := true
	for {
		message, more := <-messages
		if more {
			result--;
			fmt.Printf("%s, %d to go\n", message.Message, result)
			if !message.Success {
				success = false
			}
		}

		if (result < 1) {
			break;
		}
	}

	if !success {
		errState("Release failed for one or more services")
	}

}

func _describeTaskDefinition(taskDefinitionName string, svc *ecs.ECS) *ecs.DescribeTaskDefinitionOutput {
	if svc == nil {
		svc = ecs.New(session.New(), _getAwsConfig())
	}

	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionName),
	}

	result, err := svc.DescribeTaskDefinition(params)

	assertError(err);

	return result
}

func _stopTask(cluster, taskArn string, svc *ecs.ECS) error {
	if svc == nil {
		svc = ecs.New(session.New(), _getAwsConfig())
	}

	params := &ecs.StopTaskInput{
		Cluster: aws.String(cluster),
		Reason:  aws.String("Stopped by WriterTool"),
		Task:    aws.String(taskArn),
	}

	_, err := svc.StopTask(params)

	if err != nil {
		return err
	}

	return nil
}

func GetServiceArn(clusterArn, name string, svc *ecs.ECS) string {
	clusterArns := _listServices(clusterArn, svc)

	for i := 0; i < len(clusterArns.ServiceArns); i++ {
		arn := clusterArns.ServiceArns[i]
		if ExtractName(arn) == name {
			return *arn
		}
	}

	return ""
}
