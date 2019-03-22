package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"regexp"
	"sort"
	"strconv"
	"time"
)

func ExtractName(a *string) string {
	//noinspection RegExpRedundantEscape
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
	//noinspection RegExpRedundantEscape
	re := regexp.MustCompile("(.*?)\\:(.+)")
	res := re.FindAllStringSubmatch(a, -1)

	if res == nil {
		return "", ""
	}

	return res[0][2], res[0][1]
}

func ListServices(clusterArn string) {
	resp := listServices(clusterArn, nil)

	for i := 0; i < len(resp.ServiceArns); i++ {
		name := ExtractName(resp.ServiceArns[i])
		fmt.Println(name)
	}
}

func ListTasks(clusterArn, serviceArn string) {
	resp, err := listTasks(clusterArn, serviceArn, nil)

	if err != nil {
		errState(err.Error())
	}

	for i := 0; i < len(resp.TaskArns); i++ {
		name := ExtractName(resp.TaskArns[i])
		fmt.Println(name)
	}
}

type ByName []*ecs.Attribute

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return *a[i].Name < *a[j].Name }

func DescribeContainerInstances(clusterArn string) {
	resp := describeContainerInstances(clusterArn, nil)

	fmt.Printf("Number of container instances for cluster: %d\n", len(resp.ContainerInstances))

	for i := 0; i < len(resp.ContainerInstances); i++ {
		instance := resp.ContainerInstances[i]
		fmt.Printf("\nEC2 Instance ID: %s\n", *instance.Ec2InstanceId)

		sort.Sort(ByName(instance.Attributes))

		for j := 0; j < len(instance.Attributes); j++ {
			attribute := instance.Attributes[j]
			if attribute.Value != nil {
				if verboseLevel < 1 {
					fmt.Printf("   %s\n", *attribute.Name)
					fmt.Printf("      %s\n", *attribute.Value)
				}
			} else {
				fmt.Printf("   %s\n", *attribute.Name)
			}
		}
	}
}

func UpdateService(clusterArn, serviceArn string) {
	message, err := updateService(clusterArn, serviceArn, nil, nil)
	assertError(err)
	fmt.Println(message)
}

type Update struct {
	Cluster       string `json:"cluster"`
	Service       string `json:"service"`
	Profile       string `json:"profile"`
	Label         string `json:"label"`
	ContainerName string `json:"containerName"`
	Region        string `json:"region"`
}

type Report struct {
	Message string
	Success bool
}

func UpdateServices(data []byte) {
	var updateConfig []Update

	err := json.Unmarshal(data, &updateConfig)
	assertError(err)

	messages := make(chan Report, len(updateConfig))
	fmt.Printf("Performing update on %d services.\n", len(updateConfig))

	for i := 0; i < len(updateConfig); i++ {
		config := updateConfig[i]

		if config.Profile == "" {
			messages <- Report{Message: config.Label + ": No Profile specified for cluster: " + config.Cluster + ", service: " + config.Service, Success: false}
		} else {
			sess, cfg := getSessionAndConfigForParams(config.Profile, config.Region)
			svc := ecs.New(sess, cfg)

			clusterArn := GetClusterArn(config.Cluster, svc)
			serviceArn := GetServiceArn(clusterArn, config.Service, svc)

			fmt.Println(config.Label + ": Updating service " + config.Service)
			//noinspection GoUnhandledErrorResult
			go updateService(clusterArn, serviceArn, messages, svc)
		}
	}

	result := len(updateConfig)
	success := true

	for {
		message, more := <-messages
		if more {
			result--
			fmt.Printf("%s, %d to go\n", message.Message, result)
			if !message.Success {
				success = false
			}
		}

		if result < 1 {
			break
		}
	}

	if !success {
		errState("Update failed for one or more services")
	}
}

func DescribeService(clusterArn, serviceArn string) {
	service := describeService(clusterArn, serviceArn, nil)

	for n := 0; n < len(service.Services); n++ {
		item := service.Services[n]

		if verboseLevel == 0 {
			fmt.Printf("Service name [%s], Running: %d, Pending: %d, Desired: %d\n", *item.ServiceName, *item.RunningCount, *item.PendingCount, *item.DesiredCount)
		}

		if verboseLevel == 1 {
			fmt.Printf("Service name [%s], Running: %d, Pending: %d, Desired: %d\n", *item.ServiceName, *item.RunningCount, *item.PendingCount, *item.DesiredCount)

			for i := 0; i < len(item.Deployments); i++ {
				deployment := item.Deployments[i]
				fmt.Printf("   %s (%s), running: %d, Pending: %d: Desired: %d\n", ExtractName(deployment.TaskDefinition), *deployment.Status, *deployment.RunningCount, *deployment.PendingCount, *deployment.DesiredCount)
			}
		}

		if verboseLevel == 2 {
			definition := describeTaskDefinition(*item.TaskDefinition, nil)

			jsonBytes, err := json.MarshalIndent(definition, "", " ")
			assertError(err)

			fmt.Println(string(jsonBytes))
		}
	}
}

func ReleaseService(clusterArn, serviceArn, version string) {
	message, err := releaseService(clusterArn, serviceArn, containerName, version, nil, nil)

	if err != nil {
		errState(err.Error())
	}

	fmt.Println(message)
}

func getContainerIndexForName(definitions []*ecs.ContainerDefinition, name string) int {
	for i := 0; i < len(definitions); i++ {
		definition := definitions[i]
		if *definition.Name == name {
			return i
		}
	}

	return -1
}

func ReleaseServices(version string, data []byte) {
	var updateConfig []Update

	err := json.Unmarshal(data, &updateConfig)
	assertError(err)

	messages := make(chan Report, len(updateConfig))
	fmt.Printf("Performing release to %s on %d services\n", version, len(updateConfig))

	for i := 0; i < len(updateConfig); i++ {
		config := updateConfig[i]

		if config.Profile == "" {
			messages <- Report{Message: config.Label + ": No Profile specified for cluster: " + config.Cluster + ", service: " + config.Service, Success: false}
		} else {
			sess, cfg := getSessionAndConfigForParams(config.Profile, config.Region)
			svc := ecs.New(sess, cfg)

			clusterArn := GetClusterArn(config.Cluster, svc)
			serviceArn := GetServiceArn(clusterArn, config.Service, svc)

			localContainerName := containerName
			if config.ContainerName != "" {
				localContainerName = config.ContainerName
			}

			fmt.Println(config.Label + ": Releasing service " + config.Service + ", containerName: " + localContainerName)
			//noinspection GoUnhandledErrorResult
			go releaseService(clusterArn, serviceArn, localContainerName, version, messages, svc)
		}
	}

	result := len(updateConfig)
	success := true

	for {
		message, more := <-messages
		if more {
			result--
			fmt.Printf("%s, %d to go\n", message.Message, result)

			if !message.Success {
				success = false
			}
		}

		if result < 1 {
			break
		}
	}

	if !success {
		errState("Release failed for one or more services")
	}
}

func describeTaskDefinition(taskDefinitionName string, svc *ecs.ECS) *ecs.DescribeTaskDefinitionOutput {
	if svc == nil {
		sess, cfg := getSessionAndConfig()
		svc = ecs.New(sess, cfg)
	}

	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionName),
	}

	result, err := svc.DescribeTaskDefinition(params)
	assertError(err)

	return result
}

func _stopTask(cluster, taskArn string, svc *ecs.ECS) error {
	if svc == nil {
		sess, cfg := getSessionAndConfig()
		svc = ecs.New(sess, cfg)
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
	clusterArns := listServices(clusterArn, svc)

	for i := 0; i < len(clusterArns.ServiceArns); i++ {
		arn := clusterArns.ServiceArns[i]
		if ExtractName(arn) == name {
			return *arn
		}
	}

	return ""
}

func listServices(cluster string, svc *ecs.ECS) *ecs.ListServicesOutput {
	if svc == nil {
		sess, cfg := getSessionAndConfig()
		svc = ecs.New(sess, cfg)
	}

	var marker = new(string)
	var result = new(ecs.ListServicesOutput)

	for marker != nil && len(result.ServiceArns) < int(maxResult) {
		if marker != nil && *marker == "" {
			marker = nil
		}

		params := &ecs.ListServicesInput{
			Cluster:    aws.String(cluster),
			NextToken:  marker,
			MaxResults: &maxResult,
		}

		resp, err := svc.ListServices(params)
		assertError(err)

		result.ServiceArns = append(result.ServiceArns, resp.ServiceArns...)
		marker = resp.NextToken
	}

	return result
}

func listTasks(cluster, service string, svc *ecs.ECS) (*ecs.ListTasksOutput, error) {
	if svc == nil {
		sess, cfg := getSessionAndConfig()
		svc = ecs.New(sess, cfg)
	}

	var marker = new(string)
	var result = new(ecs.ListTasksOutput)

	for marker != nil && len(result.TaskArns) < int(maxResult) {
		if marker != nil && *marker == "" {
			marker = nil
		}

		params := &ecs.ListTasksInput{
			Cluster:     aws.String(cluster),
			ServiceName: aws.String(service),
			MaxResults:  &maxResult,
		}

		resp, err := svc.ListTasks(params)
		if err != nil {
			return nil, err
		}

		result.TaskArns = append(result.TaskArns, resp.TaskArns...)
		marker = result.NextToken
	}

	return result, nil
}

func describeService(clusterArn, serviceArn string, svc *ecs.ECS) *ecs.DescribeServicesOutput {
	if svc == nil {
		sess, cfg := getSessionAndConfig()
		svc = ecs.New(sess, cfg)
	}

	params := &ecs.DescribeServicesInput{
		Cluster:  aws.String(clusterArn),
		Services: []*string{aws.String(serviceArn)},
	}

	result, err := svc.DescribeServices(params)
	assertError(err)

	return result
}

func describeContainerInstances(clusterArn string, svc *ecs.ECS) *ecs.DescribeContainerInstancesOutput {
	if svc == nil {
		sess, cfg := getSessionAndConfig()
		svc = ecs.New(sess, cfg)
	}

	var marker = new(string)
	var containerInstanceResult = new(ecs.ListContainerInstancesOutput)

	for marker != nil && len(containerInstanceResult.ContainerInstanceArns) < int(maxResult) {
		if marker != nil && *marker == "" {
			marker = nil
		}

		params := &ecs.ListContainerInstancesInput{
			Cluster:    aws.String(clusterArn),
			NextToken:  marker,
			MaxResults: &maxResult,
		}

		resp, err := svc.ListContainerInstances(params)
		assertError(err)

		containerInstanceResult.ContainerInstanceArns = append(containerInstanceResult.ContainerInstanceArns, resp.ContainerInstanceArns...)
		marker = resp.NextToken
	}

	params := &ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(clusterArn),
		ContainerInstances: containerInstanceResult.ContainerInstanceArns,
	}

	result, err := svc.DescribeContainerInstances(params)
	assertError(err)

	return result
}

func createTaskDefinition(taskDefinition *ecs.DescribeTaskDefinitionOutput, svc *ecs.ECS) string {
	if svc == nil {
		sess, cfg := getSessionAndConfig()
		svc = ecs.New(sess, cfg)
	}

	params := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: taskDefinition.TaskDefinition.ContainerDefinitions,
		Family:               taskDefinition.TaskDefinition.Family,
		Volumes:              taskDefinition.TaskDefinition.Volumes,
	}

	registrationResult, err := svc.RegisterTaskDefinition(params)
	assertError(err)

	taskDefinitionArn := registrationResult.TaskDefinition.TaskDefinitionArn
	return *taskDefinitionArn
}

func updateTaskDefinitionForService(newTaskDefinitionArn string, service *ecs.DescribeServicesOutput, svc *ecs.ECS) {
	if svc == nil {
		sess, cfg := getSessionAndConfig()
		svc = ecs.New(sess, cfg)
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
	assertError(err)

	//noinspection GoUnhandledErrorResult
	waitForUpdatedTaskDefinition(*clusterArn, *serviceArn, svc)
}

func waitForUpdatedTaskDefinition(cluster string, service string, svc *ecs.ECS) error {
	newTask := ""
	attempts := 240
	sleepTime := 2

	for i := 0; i < attempts; i++ {
		currentService := describeService(cluster, service, svc)

		for j := 0; j < len(currentService.Services); j++ {
			item := currentService.Services[j]

			for k := 0; k < len(item.Deployments); k++ {
				deployment := item.Deployments[k]

				if *deployment.Status == "PRIMARY" && *deployment.RunningCount == *deployment.DesiredCount {
					return nil
				}
			}
		}

		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	return errors.New("Task " + newTask + " did not start in " + strconv.Itoa(attempts*sleepTime) + " seconds")
}

func releaseService(clusterArn, serviceArn, containerName, version string, done chan Report, svc *ecs.ECS) (string, error) {
	service := describeService(clusterArn, serviceArn, svc)

	if len(service.Services) > 1 {
		errorMessage := *service.Services[0].ServiceName + " No support for multiple services"
		if done != nil {
			done <- Report{Message: errorMessage, Success: false}
		}
		return "", errors.New(errorMessage)
	}

	taskDefinitionName := *service.Services[0].TaskDefinition
	taskDefinition := describeTaskDefinition(taskDefinitionName, svc)
	containerIndex := 0

	if len(taskDefinition.TaskDefinition.ContainerDefinitions) > 1 {
		if containerName == "" {
			errorMessage := "Please specify containerName for service with multiple container definitions"
			if done != nil {
				done <- Report{Message: errorMessage, Success: false}
			}
			return "", errors.New(errorMessage)
		}

		containerIndex = getContainerIndexForName(taskDefinition.TaskDefinition.ContainerDefinitions, containerName)
		if containerIndex == -1 {
			errorMessage := "No container named " + containerName + " found in task definition"
			if done != nil {
				done <- Report{Message: errorMessage, Success: false}
			}
			return "", errors.New(errorMessage)
		}
	}

	dockerImage := *taskDefinition.TaskDefinition.ContainerDefinitions[containerIndex].Image
	currentVersion, imagePart := ExtractVersion(dockerImage)

	if verboseLevel > 0 {
		fmt.Printf("Service            [%s]\n", *service.Services[0].ServiceName)
		fmt.Printf("Task definition    [%s]\n", taskDefinitionName)
		fmt.Printf("Docker image       [%s]\n", imagePart)
		fmt.Printf("Version            [%s]\n", currentVersion)
	}

	if version == currentVersion {
		errMessage := *service.Services[0].ServiceName + " Specified version is already deployed!"
		if done != nil {
			done <- Report{Message: errMessage, Success: false}
		}
		return "", errors.New(errMessage)
	}

	minimumHealthyPercentage := *service.Services[0].DeploymentConfiguration.MinimumHealthyPercent
	desiredCount := *service.Services[0].DesiredCount
	minimumHealthyCount := desiredCount * minimumHealthyPercentage / 100

	if verboseLevel > 0 {
		fmt.Printf("HealthyPercentage [%d], DesiredCount [%d] -> Number of running instances during deployment [%d] ... ", minimumHealthyPercentage, desiredCount, minimumHealthyCount)
	}

	if desiredCount-minimumHealthyCount == 0 {
		errMessage := *service.Services[0].ServiceName + " Not possible to deploy because of too high healthy percentage"

		if done != nil {
			done <- Report{Message: errMessage, Success: false}
		}
		return "", errors.New(errMessage)
	}

	*taskDefinition.TaskDefinition.ContainerDefinitions[containerIndex].Image = imagePart + ":" + version
	newTaskDefinitionArn := createTaskDefinition(taskDefinition, svc)

	updateTaskDefinitionForService(newTaskDefinitionArn, service, svc)
	message := "Service " + *service.Services[0].ServiceName + " is released with version " + version

	if done != nil {
		done <- Report{Message: message, Success: true}
	}

	return message, nil
}

func updateService(clusterArn, serviceArn string, done chan Report, svc *ecs.ECS) (string, error) {
	tasks, err := listTasks(clusterArn, serviceArn, svc)
	if err != nil {
		if done != nil {
			done <- Report{Message: err.Error(), Success: false}
		}

		return "", err
	}

	service := describeService(clusterArn, serviceArn, svc)
	if len(service.Services) > 1 {
		errMessage := "No support for multiple services"
		if done != nil {
			done <- Report{Message: errMessage, Success: false}
		}
		return "", errors.New(errMessage)
	}

	desiredCount := *service.Services[0].DesiredCount
	if len(tasks.TaskArns) < int(desiredCount) {
		errorMessage := "The number of actual tasks " + strconv.Itoa(len(tasks.TaskArns)) + " differs from desired tasks " + strconv.FormatInt(desiredCount, 10)
		if done != nil {
			done <- Report{Message: errorMessage, Success: false}
		}

		return "", errors.New(errorMessage)
	}

	for i := 0; i < len(tasks.TaskArns); i++ {
		err = _stopTask(clusterArn, *tasks.TaskArns[i], svc)
		if err != nil {
			if done != nil {
				done <- Report{Message: err.Error(), Success: false}
			}

			return "", err
		}

		err = waitForUpdatedTaskDefinition(clusterArn, serviceArn, svc)
		if err != nil {
			if done != nil {
				done <- Report{Message: err.Error(), Success: false}
			}

			return "", err
		}
	}

	message := "Service " + *service.Services[0].ServiceName + " is updated"
	if done != nil {
		done <- Report{Message: message, Success: true}
	}

	return message, nil
}
