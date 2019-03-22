package main

import (
	"flag"
	"fmt"
	"os"
)

type CommandHelp struct {
	CommandName        string
	CommandDescription string
	Parameters         []Parameter
}

type Parameter struct {
	ParameterName        string
	ParameterDescription string
	Required             bool
}

func newCommandHelp(name, description string) *CommandHelp {
	return &CommandHelp{CommandName: name, CommandDescription: description}
}

func newParameter(name, description string, required bool) *Parameter {
	return &Parameter{ParameterName: name, ParameterDescription: description, Required: required}
}

func printCommandHelp() {
	var commands []CommandHelp

	// Help
	help := newCommandHelp("help", "Prints this help")
	commands = append(commands, *help)
	//commands := [...]CommandHelp{*help}

	// copyFileFromS3Bucket
	copyFileFromS3Bucket := newCommandHelp("copyFileFromS3Bucket", "Copies file from S3 to local system")
	copyFileFromS3Bucket.Parameters = append(copyFileFromS3Bucket.Parameters,
		*newParameter("s3bucket", "The source bucket", true),
		*newParameter("s3filename", "The filename to copy", true),
		*newParameter("output", "The target directory", true),
	)
	commands = append(commands, *copyFileFromS3Bucket)

	createReleaseNotes := newCommandHelp("createReleaseNotes", "Create release notes document from Jira issues")
	createReleaseNotes.Parameters = append(createReleaseNotes.Parameters,
		*newParameter("reportConfig", "The configuration file used to fetch issues from Jira", true),
		*newParameter("reportTemplate", "Transforms jira issues into release notes file", true),
		*newParameter("dependenciesFile", "Specifies dependencies for service, used in template", false),
		*newParameter("version", "Specifies the version that is released. Can be used in template as .Version", false),
		*newParameter("releaseDate", "Specifies the release date. Can be used in template as .ReleaseDate", false),
		*newParameter("login", "Jira login", false),
		*newParameter("password", "Jira password", false),
	)
	commands = append(commands, *createReleaseNotes)

	createReport := newCommandHelp("createReport", "Generates a report of running services")
	createReport.Parameters = append(createReport.Parameters,
		*newParameter("reportConfig", "The configuration file used to fetch issues from Jira", true),
		*newParameter("reportTemplate", "Transform jira issues into release notes file", true),
	)
	commands = append(commands, *createReport)

	listS3Buckets := newCommandHelp("listS3Buckets", "List available S3 buckets")
	commands = append(commands, *listS3Buckets)

	listFilesInS3Bucket := newCommandHelp("listFilesInS3Bucket", "List available objects in an S3 bucket")
	listFilesInS3Bucket.Parameters = append(listFilesInS3Bucket.Parameters,
		*newParameter("s3Bucket", "The source bucket", true),
		*newParameter("s3Filename", "The filename to copy could be used as prefix for filtering", false),
	)
	commands = append(commands, *listFilesInS3Bucket)

	listClusters := newCommandHelp("listClusters", "List available clusters. -v will also list services for all clusters")
	commands = append(commands, *listClusters)

	listServices := newCommandHelp("listServices", "List available services")
	listServices.Parameters = append(listServices.Parameters,
		*newParameter("cluster", "Cluster for which to list services", true),
	)
	commands = append(commands, *listServices)

	listTasks := newCommandHelp("listTasks", "List tasks for a service")
	listTasks.Parameters = append(listTasks.Parameters,
		*newParameter("cluster", "Cluster for which to list tasks", true),
		*newParameter("service", "Service for which to list tasks", true),
	)
	commands = append(commands, *listTasks)

	deployLambdaFunction := newCommandHelp("deployLambdaFunction", "Deploys new code for a lambda function")
	deployLambdaFunction.Parameters = append(deployLambdaFunction.Parameters,
		*newParameter("s3bucket", "The bucket where the new code is placed in", true),
		*newParameter("s3filename", "The filename of the code zip", true),
		*newParameter("functionName", "The name of the function to update", true),
		*newParameter("publish", "'True' to publish a new version (optional) default 'false'", false),
		*newParameter("alias", "The alias to update", true),
		*newParameter("version", "The version number to publish", true),
		*newParameter("runtime", "Specifies the runtime to use for the lambda function. See https://docs.aws.amazon.com/cli/latest/reference/lambda/update-function-configuration.html for runtimes. Uses current runtime for function if unset", true),
	)
	commands = append(commands, *deployLambdaFunction)

	describeService := newCommandHelp("describeService", "Describes the service. Optionally -v and -vv may be used")
	describeService.Parameters = append(describeService.Parameters,
		*newParameter("cluster", "Cluster for which the service belongs", true),
		*newParameter("service", "Service to describe", true),
	)
	commands = append(commands, *describeService)

	updateService := newCommandHelp("updateService", "Stop/start all running tasks for the specified service")
	updateService.Parameters = append(updateService.Parameters,
		*newParameter("cluster", "Cluster for which the service to update belongs", true),
		*newParameter("service", "Service to update", true),
	)
	commands = append(commands, *updateService)

	updateServices := newCommandHelp("updateServices", "Stop/start all running tasks for specified services")
	updateServices.Parameters = append(updateServices.Parameters,
		*newParameter("updatesFile", "[{\"profile\": \"(profile in credential file)\", \"region\": \"(region to use (if not specified, writer-tool will use region specified in credential file))\", \"cluster\": \"(cluster as reported using -listClusters)\", \"service\": \"(service as reported using -listServices)\", \"containerName\": \"(name of container to update (if multiple containers in same service))\", \"label\": \"(Label that should be used in output for service)\"}]", true),
	)
	commands = append(commands, *updateServices)

	releaseService := newCommandHelp("releaseService", "Creates a new release for the service")
	releaseService.Parameters = append(releaseService.Parameters,
		*newParameter("cluster", "Cluster for which the service to release belongs", true),
		*newParameter("service", "Service to release", true),
		*newParameter("version", "Version to release", true),
	)
	commands = append(commands, *releaseService)

	releaseServices := newCommandHelp("releaseServices", "Release all services specified")
	releaseServices.Parameters = append(releaseServices.Parameters,
		*newParameter("updatesFile", "[{\"profile\": \"(profile in credential file)\", \"region\": \"(region to use (if not specified, writer-tool will use region specified in credential file))\", \"cluster\": \"(cluster as reported using -listClusters)\", \"service\": \"(service as reported using -listServices)\", \"containerName\": \"(name of container to update (if multiple containers in same service))\", \"label\": \"(Label that should be used in output for service)\"}]", true),
	)
	commands = append(commands, *releaseServices)

	listEc2Instances := newCommandHelp("listEc2Instances", "List available EC2 instances")
	commands = append(commands, *listEc2Instances)

	listLoadBalancers := newCommandHelp("listLoadBalancers", "List available Load Balancers and their contained EC2 instances")
	commands = append(commands, *listLoadBalancers)

	listLambdaFunctions := newCommandHelp("listLambdaFunctions", "List available lambda functions")
	commands = append(commands, *listLambdaFunctions)

	getLambdaFunctionInfo := newCommandHelp("getLambdaFunctionInfo", "Get lambda function information")
	getLambdaFunctionInfo.Parameters = append(getLambdaFunctionInfo.Parameters,
		*newParameter("functionName", "Name of lambda for which to fetch info", true),
	)
	commands = append(commands, *getLambdaFunctionInfo)

	getLambdaFunctionAliasInfo := newCommandHelp("getLambdaFunctionAliasInfo", "Get lambda function information")
	getLambdaFunctionAliasInfo.Parameters = append(getLambdaFunctionAliasInfo.Parameters,
		*newParameter("functionName", "Name of lambda for which to fetch info", true),
		*newParameter("alias", "Alias of lambda for which to fetch info", true),
	)
	commands = append(commands, *getLambdaFunctionAliasInfo)

	getEntity := newCommandHelp("getEntity", "Gets an entity from the writer load balancer")
	getEntity.Parameters = append(getEntity.Parameters,
		*newParameter("loadBalancer", "The load balancer fronting the writer instances", true),
		*newParameter("{entityId}", "The ID of the entity to fetch", true),
	)
	commands = append(commands, *getEntity)

	ssh := newCommandHelp("ssh", "Executes a command over SSH for the specified service")
	ssh.Parameters = append(ssh.Parameters,
		*newParameter("instanceName", "The aws instance(s) to use as source(s). Operation will occur on all instances with the specific name (required if instanceId is not specified)", false),
		*newParameter("instanceId", "The specific aws instance to use as source. (required if instanceName is not specified)", false),
		*newParameter("pemfile", "The SSH pem file used for authentication", true),
		*newParameter("{command}", "The command to execute (e.g. 'ls -l')", true),
	)
	commands = append(commands, *ssh)

	scp := newCommandHelp("scp", "Copies files from the specified instance(s)")
	scp.Parameters = append(scp.Parameters,
		*newParameter("instanceName", "The aws instance(s) to use as source(s). Operation will occur on all instances with the specific name (required if instanceId is not specified)", false),
		*newParameter("instanceId", "The specific aws instance to use as source. (required if instanceName is not specified)", false),
		*newParameter("pemfile", "The SSH pem file used for authentication", false),
		*newParameter("output", "The target directory", true),
		*newParameter("recursive", "Copies from source recursively", false),
	)
	commands = append(commands, *scp)

	login := newCommandHelp("login", "Log in to instance using SSH")
	login.Parameters = append(login.Parameters,
		*newParameter("instanceName", "The aws instance(s) to use as source(s). Operation will occur on all instances with the specific name (required if instanceId is not specified)", false),
		*newParameter("instanceId", "The specific aws instance to use as source. (required if instanceName is not specified)", false),
		*newParameter("pemfile", "The SSH pem file used for authentication", false),
	)
	commands = append(commands, *login)

	version := newCommandHelp("version", "Display writer-tool version")
	commands = append(commands, *version)

	// Print commands and parameters
	maxCmdNameLen := 0
	maxParamNameLen := 0

	// Calculate indentations
	for i := 0; i < len(commands); i++ {
		command := commands[i]
		if len(command.CommandName) > maxCmdNameLen {
			maxCmdNameLen = len(command.CommandName)
		}

		for j := 0; j < len(command.Parameters); j++ {
			parameter := command.Parameters[j]
			if len(parameter.ParameterName) > maxParamNameLen {
				maxParamNameLen = len(parameter.ParameterName)
			}
		}
	}

	// Do print...
	for k := 0; k < len(commands); k++ {
		command := commands[k]
		cmdCurlen := len(command.CommandName)
		cmdOffset := maxCmdNameLen - cmdCurlen

		cmdLine := command.CommandName + "  "
		for l := 0; l < cmdOffset; l++ {
			cmdLine = cmdLine + " "
		}

		cmdLine = "  " + cmdLine + command.CommandDescription + "\n"
		fmt.Print(cmdLine)

		for j := 0; j < len(command.Parameters); j++ {
			parameter := command.Parameters[j]

			req := "optional"
			if parameter.Required {
				req = "required"
			}

			paramCurlen := len(parameter.ParameterName)
			paramOffset := maxParamNameLen - paramCurlen

			if maxCmdNameLen > maxParamNameLen {
				paramOffset = maxCmdNameLen - paramCurlen
			}

			paramLine := "-" + parameter.ParameterName + "  "
			for n := 0; n < paramOffset; n++ {
				paramLine = paramLine + " "
			}

			paramLine = "      " + paramLine + parameter.ParameterDescription + " (" + req + ")\n"
			fmt.Print(paramLine)
		}

		fmt.Print("\n")
	}
}

func validateListFilesInS3Bucket() {
	if bucket == "" {
		errUsage("s3bucket must be specified")
	}
}

func validateCopyFileFromS3Bucket() {
	if bucket == "" {
		errUsage("s3bucket must be specified")
	}

	if filename == "" {
		errUsage("s3filename must be specified")
	}

	if output == "" {
		errUsage("output must be specified")
	}
}

func validateDeployLambdaFunction() {
	if bucket == "" {
		errUsage("s3bucket must be speficied")
	}

	if filename == "" {
		errUsage("s3filename must be specified")
	}

	if functionName == "" {
		errUsage("functionName must be specified")
	}

	if publish == "true" && alias == "" {
		errUsage("alias must be specified when publishing")
	}

	if publish == "true" && version == "" {
		errUsage("version must be specified when publishing")
	}
}

func executeCommand() {
	switch command {
	case "copyFileFromS3Bucket":
		validateCopyFileFromS3Bucket()
		CopyFileFromS3Bucket(bucket, filename, output)
	case "createReleaseNotes":
		bytes := readConfigFromFile()
		template := readTemplateFromFile()
		version := getVersion()
		dependencies := readDependenciesFromFile()
		GenerateReleaseNotes(bytes, template, version, releaseDate, dependencies)
	case "createReport":
		bytes := readConfigFromFile()
		template := readTemplateFromFile()
		GenerateReport(bytes, template)
	case "deployLambdaFunction":
		validateDeployLambdaFunction()
		DeployLambdaFunction(functionName, bucket, filename, alias, version, runtime, publish)
	case "listS3Buckets":
		ListS3Buckets()
	case "listFilesInS3Bucket":
		validateListFilesInS3Bucket()
		ListFilesInS3Bucket(bucket, filename)
	case "listClusters":
		ListClusters()
	case "listServices":
		clusterArn := getClusterArn()
		ListServices(clusterArn)
	case "listTasks":
		clusterArn := getClusterArn()
		serviceArn := getServiceArn()
		ListTasks(clusterArn, serviceArn)
	case "describeService":
		clusterArn := getClusterArn()
		serviceArn := getServiceArn()
		DescribeService(clusterArn, serviceArn)
	case "describeContainerInstances":
		clusterArn := getClusterArn()
		DescribeContainerInstances(clusterArn)
	case "updateService":
		clusterArn := getClusterArn()
		serviceArn := getServiceArn()
		UpdateService(clusterArn, serviceArn)
	case "updateServices":
		updatesFile := getUpdatesFile()
		UpdateServices(updatesFile)
	case "releaseService":
		clusterArn := getClusterArn()
		serviceArn := getServiceArn()
		version := getVersion()
		ReleaseService(clusterArn, serviceArn, version)
	case "releaseServices":
		updatesFile := getUpdatesFile()
		version := getVersion()
		ReleaseServices(version, updatesFile)
	case "listEc2Instances":
		ListEc2Instances(instanceName)
	case "listLoadBalancers":
		ListLoadBalancers()
	case "listLambdaFunctions":
		ListLambdaFunctions()
	case "ssh":
		if sshPem == "" {
			errUsage("A SSH PEM file must be specified")
		}
		if instanceId != "" {
			instance := GetInstanceForId(instanceId)
			Ssh(instance, sshPem, flag.Args())
		} else if instanceName != "" {
			instances := GetInstancesForName(instanceName)
			if len(instances) == 1 {
				Ssh(instances[0], sshPem, flag.Args())
			} else {
				for i := 0; i < len(instances); i++ {
					fmt.Printf("[%s]\n", *instances[i].InstanceId)
					Ssh(instances[i], sshPem, flag.Args())
				}
			}
		} else {
			errUsage("Either instanceId or instanceName parameter has to be specified")
		}
	case "login":
		if instanceId != "" {
			instance := GetInstanceForId(instanceId)
			SshLogin(instance, sshPem)
		} else if instanceName != "" {
			instances := GetInstancesForName(instanceName)
			if len(instances) == 1 {
				SshLogin(instances[0], sshPem)
			} else {
				fmt.Printf("Found %d instances with name %s\n", len(instances), instanceName)
				fmt.Println("Please specify instance ID with -instanceId flag for:")
				for i := 0; i < len(instances); i++ {
					fmt.Printf("   %s\n", *instances[i].InstanceId)
				}
				os.Exit(1)
			}
		} else {
			errUsage("Either instanceId or instanceName parameter has to be specified")
		}
	case "scp":
		if sshPem == "" {
			errUsage("A SSH PEM file must be specified")
		}
		if instanceId != "" {
			instance := GetInstanceForId(instanceId)
			Scp(instance, sshPem, flag.Args())
		} else if instanceName != "" {
			ips := GetInstancesForName(instanceName)
			if len(ips) == 1 {
				fmt.Printf("[%s] ... ", *ips[0].InstanceId)
				Scp(ips[0], sshPem, flag.Args())
				fmt.Println("done")
			} else {
				for i := 0; i < len(ips); i++ {
					fmt.Printf("[%s] ... ", *ips[i].InstanceId)
					Scp(ips[i], sshPem, flag.Args())
					fmt.Println("done")
				}
			}
		} else {
			errUsage("Either instanceId or instanceName parameter has to be specified")
		}
	case "getEntity":
		if loadBalancer == "" {
			errUsage("loadBalancer must be specified")
		}
		if len(flag.Args()) != 1 {
			errUsage("Entity ID must be provided")
		}
		GetEntity(loadBalancer, flag.Args()[0])
	case "getLambdaFunctionInfo":
		if functionName == "" {
			errUsage("functionName needs to be specified")
		}
		GetLambdaFunctionInfo(functionName)
	case "getLambdaFunctionAliasInfo":
		if functionName == "" {
			errUsage("functionName needs to be specified")
		}
		if alias == "" {
			errUsage("alias needs to be specified")
		}
		GetLambdaFunctionAliasInfo(functionName, alias)
	case "version":
		fmt.Println(appVersion)
	case "help":
		printCommandHelp()
	default:
		errUsage("Unknown command: " + command)
	}
}
