package main

import (
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"fmt"
	"os"
	"time"
)

func ExtractName(a *string) string {
	re := regexp.MustCompile("\\w+:\\w+:\\w+:[\\w-]+:\\d+:\\w+\\/(.+)")

	res := re.FindAllStringSubmatch(*a, -1);

	if res == nil {
		return "No match for " + *a;
	}

	return res[0][1]
}

func ExtractVersion(a string) (string, string) {
	re := regexp.MustCompile("(.*?)\\:(.+)")

	res := re.FindAllStringSubmatch(a, -1);

	if (res == nil) {
		return "", ""
	}

	return res[0][2], res[0][1]
}

func _listServices(cluster string) *ecs.ListServicesOutput {
	svc := ecs.New(session.New(), _getAwsConfig())

	params := &ecs.ListServicesInput{
		Cluster: aws.String(cluster),
		MaxResults: aws.Int64(10),
	}

	resp, err := svc.ListServices(params)
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		os.Exit(1);
	}

	return resp;
}

func _listTasks(cluster, service string) *ecs.ListTasksOutput {
	svc := ecs.New(session.New(), _getAwsConfig())

	params := &ecs.ListTasksInput{
		Cluster: aws.String(cluster),
		ServiceName: aws.String(service),
	}

	resp, err := svc.ListTasks(params)
	if (err != nil) {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return resp
}

func _describeService(clusterArn, serviceArn string) (*ecs.DescribeServicesOutput) {
	svc := ecs.New(session.New(), _getAwsConfig())

	params := &ecs.DescribeServicesInput{
		Cluster: aws.String(clusterArn),
		Services: []*string{aws.String(serviceArn)},
	}

	result, err := svc.DescribeServices(params)
	if (err != nil) {
		errState(err.Error())
	}

	return result;
}

func _describeTasks(clusterArn string, tasks []*string) *ecs.DescribeTasksOutput {
	svc := ecs.New(session.New(), _getAwsConfig())

	params := &ecs.DescribeTasksInput{
		Cluster: aws.String(clusterArn),
		Tasks: tasks,
	}

	result, err := svc.DescribeTasks(params)

	if (err != nil) {
		errState(err.Error())
	}

	return result
}

func _createTaskDefinition(taskDefinition *ecs.DescribeTaskDefinitionOutput) string {
	svc := ecs.New(session.New(), _getAwsConfig())

	params := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions : taskDefinition.TaskDefinition.ContainerDefinitions,
		Family: taskDefinition.TaskDefinition.Family,
		Volumes: taskDefinition.TaskDefinition.Volumes,
	}

	registrationResult, err := svc.RegisterTaskDefinition(params)

	if (err != nil) {
		errState(err.Error())
	}

	taskDefinitionArn := registrationResult.TaskDefinition.TaskDefinitionArn;

	return *taskDefinitionArn
}

func _updateService(newTaskDefinitionArn string, service *ecs.DescribeServicesOutput) {
	svc := ecs.New(session.New(), _getAwsConfig())

	params := &ecs.UpdateServiceInput{
		Cluster: aws.String(*service.Services[0].ClusterArn),
		DesiredCount: aws.Int64(*service.Services[0].DesiredCount),
		Service: aws.String(*service.Services[0].ServiceArn),
		TaskDefinition: aws.String(newTaskDefinitionArn),
	}

	_, err := svc.UpdateService(params)

	if (err != nil) {
		errState(err.Error())
	}
}

func ListServices(clusterArn string) {

	resp := _listServices(clusterArn)
	for i := 0; i < len(resp.ServiceArns); i++ {
		name := ExtractName(resp.ServiceArns[i])
		fmt.Println(name)
	}
}

func ListTasks(clusterArn, serviceArn string) {
	resp := _listTasks(clusterArn, serviceArn)
	for i := 0; i < len(resp.TaskArns); i++ {
		name := ExtractName(resp.TaskArns[i])
		fmt.Println(name)
	}
}

func UpdateService(clusterArn, serviceArn string) {
	tasks := _listTasks(clusterArn, serviceArn)
	service := _describeService(clusterArn, serviceArn)

	desiredCount := *service.Services[0].DesiredCount

	if len(tasks.TaskArns) < int(desiredCount) {
		errStatef("The number of actual tasks (%d) differs from desired tasks (%d)\n", len(tasks.TaskArns), desiredCount)
	}

	for i := 0; i < len(tasks.TaskArns); i++ {
		currentTasks := _listTasks(clusterArn, serviceArn)
		fmt.Println("Task: " + *tasks.TaskArns[i])
		_stopTask(clusterArn, *tasks.TaskArns[i])
		fmt.Println("  Stopped")
		_waitForNewTask(clusterArn, serviceArn, currentTasks.TaskArns)
	}

	fmt.Printf("Service [%s] is updated\n", *service.Services[0].ServiceName)
}

func ReleaseService(clusterArn, serviceArn, version string) {
	service := _describeService(clusterArn, serviceArn);

	taskDefinitionName := *service.Services[0].TaskDefinition
	taskDefinition := _describeTaskDefinition(taskDefinitionName)
	dockerImage := *taskDefinition.TaskDefinition.ContainerDefinitions[0].Image
	currentVersion, imagePart := ExtractVersion(dockerImage)

	fmt.Printf("Service            [%s]\n", *service.Services[0].ServiceName);
	fmt.Printf("Task definition    [%s]\n", taskDefinitionName);
	fmt.Printf("Docker image       [%s]\n", imagePart)
	fmt.Printf("Version            [%s]\n", currentVersion);

	if (version == currentVersion) {
		errUsage("Specified version is already deployed!");
	}

	minimumHealthyPercentage := *service.Services[0].DeploymentConfiguration.MinimumHealthyPercent
	desiredCount := *service.Services[0].DesiredCount
	minimumHealthyCount := desiredCount * minimumHealthyPercentage / 100

	fmt.Printf("HealthyPercentage [%d], DesiredCount [%d] -> Number of running instances during deployment [%d] ... ", minimumHealthyPercentage, desiredCount, minimumHealthyCount)

	if (desiredCount - minimumHealthyCount == 0) {
		errUsage("Not possible to deploy because of too high healthy percentage")
	}

	fmt.Println("OK")

	*taskDefinition.TaskDefinition.ContainerDefinitions[0].Image = imagePart + ":" + version

	fmt.Println(*taskDefinition.TaskDefinition.ContainerDefinitions[0].Image)

	newTaskDefinitionArn := _createTaskDefinition(taskDefinition)

	_updateService(newTaskDefinitionArn, service)
}

func _describeTaskDefinition(taskDefinitionName string) *ecs.DescribeTaskDefinitionOutput {
	svc := ecs.New(session.New(), _getAwsConfig())

	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionName),
	}

	result, err := svc.DescribeTaskDefinition(params)

	if (err != nil) {
		errState(err.Error())
	}

	return result
}

func _stopTask(cluster, taskArn string) {
	svc := ecs.New(session.New(), _getAwsConfig())

	params := &ecs.StopTaskInput{
		Cluster: aws.String(cluster),
		Reason: aws.String("Stopped by WriterTool"),
		Task: aws.String(taskArn),
	}

	_, err := svc.StopTask(params)

	if (err != nil) {
		errState(err.Error())
	}
}

func _waitForNewTask(cluster string, service string, tasks []*string) {
	fmt.Print("  Waiting for new task ")
	newTask := ""

	attempts := 60
	sleepTime := 2

	for i := 0; i < attempts; i++ {
		currentTasks := _listTasks(cluster, service)
		if (len(currentTasks.TaskArns) < len(tasks) || len(currentTasks.TaskArns) == 0) {
			fmt.Print(".")
		} else if (newTask == "") {
			newTask = _findNewTask(tasks, currentTasks.TaskArns)
			if (newTask == "") {
				errState("No new task found among tasks")
			}
			fmt.Println(" done")
			fmt.Print("  New task: " + newTask + " ")

		}

		if (newTask != "") {
			taskStates := _describeTasks(cluster, []*string{aws.String(newTask)})
			if (*taskStates.Tasks[0].LastStatus == "RUNNING") {
				fmt.Println(" done")
				return
			} else if (*taskStates.Tasks[0].LastStatus == "STOPPED") {
				fmt.Print("X\n")
				errState("Could not start new task")
			} else {
				fmt.Print(".")
			}
		}

		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	errStatef("Task [%s] did not start in %d attempts", newTask, attempts)

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

func GetServiceArn(clusterArn, name string) string {
	clusterArns := _listServices(clusterArn);

	for i := 0; i < len(clusterArns.ServiceArns); i++ {
		arn := clusterArns.ServiceArns[i];
		if (ExtractName(arn) == name) {
			return *arn;
		}
	}

	return "";
}
