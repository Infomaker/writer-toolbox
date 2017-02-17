package main

import (
	"regexp"

	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"os"
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
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		os.Exit(1)
	}

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
	if err != nil {
		errState(err.Error())
	}

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

func _createTaskDefinition(taskDefinition *ecs.DescribeTaskDefinitionOutput) string {
	svc := ecs.New(session.New(), _getAwsConfig())

	params := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: taskDefinition.TaskDefinition.ContainerDefinitions,
		Family:               taskDefinition.TaskDefinition.Family,
		Volumes:              taskDefinition.TaskDefinition.Volumes,
	}

	registrationResult, err := svc.RegisterTaskDefinition(params)

	if err != nil {
		errState(err.Error())
	}

	taskDefinitionArn := registrationResult.TaskDefinition.TaskDefinitionArn

	return *taskDefinitionArn
}

func _updateTaskDefinitionForService(newTaskDefinitionArn string, service *ecs.DescribeServicesOutput) {
	svc := ecs.New(session.New(), _getAwsConfig())

	params := &ecs.UpdateServiceInput{
		Cluster:        aws.String(*service.Services[0].ClusterArn),
		DesiredCount:   aws.Int64(*service.Services[0].DesiredCount),
		Service:        aws.String(*service.Services[0].ServiceArn),
		TaskDefinition: aws.String(newTaskDefinitionArn),
	}

	_, err := svc.UpdateService(params)

	if err != nil {
		errState(err.Error())
	}
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
			done <- Report{Message:err.Error(),Success:false}
		}
		return "", err
	}
	service := _describeService(clusterArn, serviceArn, svc)

	if len(service.Services) > 1 {
		return "", errors.New("No support for multiple services.")
	}

	desiredCount := *service.Services[0].DesiredCount

	if len(tasks.TaskArns) < int(desiredCount) {
		errorMessage := "The number of actual tasks " + strconv.Itoa(len(tasks.TaskArns)) + " differs from desired tasks " + strconv.FormatInt(desiredCount, 10)
		if (done != nil) {
			done <- Report{Message:errorMessage,Success:false}
		}
		return "", errors.New(errorMessage)
	}

	for i := 0; i < len(tasks.TaskArns); i++ {
		currentTasks, err := _listTasks(clusterArn, serviceArn, svc)
		if (err != nil) {
			if (done != nil) {
				done <- Report{Message:err.Error(),Success:false}
			}
			return "", err
		}
		err = _stopTask(clusterArn, *tasks.TaskArns[i], svc)
		if err != nil {
			if (done != nil) {
				done <- Report{Message:err.Error(),Success:false}
			}
			return "", err
		}
		err = _waitForNewTask(clusterArn, serviceArn, currentTasks.TaskArns, svc)
		if err != nil {
			if (done != nil) {
				done <- Report{Message:err.Error(),Success:false}
			}
			return "", err
		}
	}
	message := "Service " + *service.Services[0].ServiceName + " is updated"

	if done != nil {
		done <- Report{Message:message,Success:true}
	}

	return message, nil
}

func UpdateService(clusterArn, serviceArn string) {
	message, err := _updateService(clusterArn, serviceArn, nil, nil)

	if err != nil {
		errState(err.Error())
	}

	fmt.Println(message)
}

type Update struct {
	Cluster   string `json:"cluster"`
	Service   string `json:"service"`
	AwsKey    string `json:"awsKey"`
	AwsSecret string `json:"awsSecret"`
	Profile   string `json:"profile"`
	Label     string `json:"label"`
}

type Report struct {
	Message string
	Success bool
}

func UpdateServices(data []byte) {

	var updateConfig []Update

	err := json.Unmarshal(data, &updateConfig)

	if err != nil {
		errState(err.Error())
	}

	messages := make(chan Report, len(updateConfig))

	fmt.Printf("Performing update on %d services.\n", len(updateConfig))

	for i := 0; i < len(updateConfig); i++ {
		config := updateConfig[i]
		if (config.AwsKey == "" && config.Profile == "") {
			messages <- Report{Message:config.Label + ": No AwsKey or Profile specified for cluster: " + config.Cluster + ", service: " + config.Service,Success:false}
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
			definition := _describeTaskDefinition(*item.TaskDefinition)
			fmt.Println(definition)
		}
	}
}

func ReleaseService(clusterArn, serviceArn, version string) {
	service := _describeService(clusterArn, serviceArn, nil)

	if len(service.Services) > 1 {
		errState("No support for multiple services.")
	}

	taskDefinitionName := *service.Services[0].TaskDefinition
	taskDefinition := _describeTaskDefinition(taskDefinitionName)
	dockerImage := *taskDefinition.TaskDefinition.ContainerDefinitions[0].Image
	currentVersion, imagePart := ExtractVersion(dockerImage)

	fmt.Printf("Service            [%s]\n", *service.Services[0].ServiceName)
	fmt.Printf("Task definition    [%s]\n", taskDefinitionName)
	fmt.Printf("Docker image       [%s]\n", imagePart)
	fmt.Printf("Version            [%s]\n", currentVersion)

	if version == currentVersion {
		errUsage("Specified version is already deployed!")
	}

	minimumHealthyPercentage := *service.Services[0].DeploymentConfiguration.MinimumHealthyPercent
	desiredCount := *service.Services[0].DesiredCount
	minimumHealthyCount := desiredCount * minimumHealthyPercentage / 100

	fmt.Printf("HealthyPercentage [%d], DesiredCount [%d] -> Number of running instances during deployment [%d] ... ", minimumHealthyPercentage, desiredCount, minimumHealthyCount)

	if desiredCount - minimumHealthyCount == 0 {
		errUsage("Not possible to deploy because of too high healthy percentage")
	}

	fmt.Println("OK")

	*taskDefinition.TaskDefinition.ContainerDefinitions[0].Image = imagePart + ":" + version

	fmt.Println(*taskDefinition.TaskDefinition.ContainerDefinitions[0].Image)

	newTaskDefinitionArn := _createTaskDefinition(taskDefinition)

	_updateTaskDefinitionForService(newTaskDefinitionArn, service)
}

func _describeTaskDefinition(taskDefinitionName string) *ecs.DescribeTaskDefinitionOutput {
	svc := ecs.New(session.New(), _getAwsConfig())

	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionName),
	}

	result, err := svc.DescribeTaskDefinition(params)

	if err != nil {
		errState(err.Error())
	}

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

func _waitForNewTask(cluster string, service string, tasks []*string, svc *ecs.ECS) error {
	newTask := ""

	attempts := 240
	sleepTime := 2
	newTaskAttempts := 5

	for i := 0; i < attempts; i++ {
		currentTasks, err := _listTasks(cluster, service, svc)
		if (err != nil) {
			return err
		}
		if len(currentTasks.TaskArns) < len(tasks) || len(currentTasks.TaskArns) == 0 {
		} else if newTask == "" {
			newTask = _findNewTask(tasks, currentTasks.TaskArns)
			newTaskAttempts--

			if newTask == "" && newTaskAttempts >= 0 {
			} else if newTask != "" {
			} else {
				return errors.New("No new task found among tasks")
			}

		}

		if newTask != "" {
			taskStates, err := _describeTasks(cluster, []*string{aws.String(newTask)}, svc)
			if (err != nil) {
				return err
			}
			if *taskStates.Tasks[0].LastStatus == "RUNNING" {
				return nil
			} else if *taskStates.Tasks[0].LastStatus == "STOPPED" {
				return errors.New("Could not start new task")
			}
		}

		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	return errors.New("Task " + newTask + " did not start in " + strconv.Itoa(attempts * sleepTime) + " seconds")
}

func _findNewTask(tasks, currentTasks []*string) string {
	for i := 0; i < len(tasks); i++ {
		for j := 0; j < len(currentTasks); j++ {
			if *tasks[i] != *currentTasks[j] {
				return *currentTasks[j]
			}
		}
	}

	return ""
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
