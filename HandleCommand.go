package main

import (
	"fmt"
	"flag"
	"os"
)

func printCommandHelp() {
	var m = map[string]string{
		"help": "Prints this help.",
		"copyFileFromS3Bucket": "Copies file from S3 to local system" +
			"                         -s3bucket     : The source bucket   (required)\n" +
			"                         -s3filename   : The filename to copy   (required)\n" +
			"                         -output       : The target directory   (required)\n" +
			"                         Example: -command copyFileFromS3Bucket -s3bucket images -s3filename cat.gif -output ~/Downloads",
		"createReleaseNotes": "Create release notes document from Jira issues. Needs -reportConfig and -reportTemplate. Optional to use -dependenciesFile\n" +
			"                             -reportConfig      : The configuration file used to fetch issues from Jira   (required)\n" +
			"                         -reportTemplate    : Transforms jira issues into release notes file   (required)\n" +
			"                         -dependenciesFile  : Specifies dependencies for service, used in template\n" +
			"                         -version           : Specifies the version that is released. Can be used in template as .Version\n" +
			"                         -releaseDate       : Specifies the release date. Can be used in template as .ReleaseDate\n" +
			"                         -login             : Jira login\n" +
			"                         -password          : Jira password\n",
		"createReport":        "Generates a report of running services. Needs -reportConfig and -reportTemplate",
		"listS3Buckets":       "List available S3 buckets",
		"listFilesInS3Bucket": "List available objects in an S3 bucket. Requires -s3bucket. -s3filename could be used as prefix for filtering",
		"listClusters":        "List available clusters. -v will also list services for all clusters",
		"listServices":        "List available services. Needs -cluster flag.",
		"listTasks":           "List tasks for a service. Needs -cluster, -service flags.",
		"deployLambdaFunction": "Deploys new code for a lambda function" +
			"                         -s3bucket     : The bucket where the new code is placed in   (required)\n" +
			"                         -s3filename   : The filename of the code zip   (required)\n" +
			"                         -functionName : The name of the function to update   (required)\n" +
			"                         -publish      : 'True' to publish a new version   (optional) default 'false'\n" +
			"                               -alias      : The alias to update   (required)\n" +
			"                               -version    : The version number to publish   (required)\n" +
			"                         Example: -command deployLambdaFunction -s3bucket newCode -s3filename myCode.zip -functionName addNumbers -publish true -alias PRIMARY -version 1.2.1",
		"describeService": "Describes the service. Needs -cluster, -service flags. Optionaly -v and -vv may be used.",
		"updateService":   "Stop/start all running tasks for the specified service. Needs -cluster, -service flags.",
		"updateServices": "Stop/start all running tasks for specified services. Needs -updateFiles flag.\n" +
			"                           -updatesFile : Path to a file containing services to update. Format of file is: \n" +
			"                             [\n" +
			"                               {\n" +
			"                                  \"awsKey\": \"(aws key)\"\n" +
			"                                  \"awsSecret\": \"(aws secret key)\"\n" +
			"                                  \"cluster\": \"(cluster as reported using -listClusters)\"\n" +
			"                                  \"service\": \"(service as reported using -listServices)\"\n" +
			"                                  \"label\": \"(Label that should be used in output for service)\"\n" +
			"                               }\n" +
			"                             ]\n",
		"releaseService": "Creates a new release for the service. Neews -cluster, -service, -version flags.",
		"releaseServices": "Release all services specified. Needs -version, -updateFiles flag. \n" +
			"                           -updatesFile : Path to a file containing services to update. Format of file is: \n" +
			"                             [\n" +
			"                               {\n" +
			"                                  \"awsKey\": \"(aws key)\"\n" +
			"                                  \"awsSecret\": \"(aws secret key)\"\n" +
			"                                  \"cluster\": \"(cluster as reported using -listClusters)\"\n" +
			"                                  \"service\": \"(service as reported using -listServices)\"\n" +
			"                                  \"label\": \"(Label that should be used in output for service)\"\n" +
			"                               }\n" +
			"                             ]\n",
		"listEc2Instances":           "List available EC2 instances.",
		"listLoadBalancers":          "List available Load Balancers and their contained EC2 instances.",
		"listLambdaFunctions":        "List available lambda functions.",
		"getLambdaFunctionInfo":      "Get lambda function information. Requires -functionName",
		"getLambdaFunctionAliasInfo": "Get lambda function information. Requires -functionName, -alias",
		"runtime":                    "Specifies the runtime to use for the lambda function. See https://docs.aws.amazon.com/cli/latest/reference/lambda/update-function-configuration.html for rumtimes. Uses current runtime for function if unset.",
		"getEntity": "Gets an entity from the writer load balancer\n" +
			"                         -loadBalancer : The load balancer fronting the writer instances    (required)\n" +
			"                         {entityId}    : The ID of the entity to fetch    (required)\n" +
			"                         Example: -command getEntity -loadBalancer writer-loadbalancer a9fbd742-ea87-425d-ae86-045ab3ac91c1",
		"ssh": "Executes a command over SSH for the specified service.\n" +
			"                         -instanceName : The aws instance(s) to use as source(s). Operation will occur on all instances with the specific name   (required if instanceId is not specified)\n" +
			"                         -instanceId   : The specific aws instance to use as source.   (required if instanceName is not specified)\n" +
			"                         -pemfile      : The SSH pem file used for authentication    (required)\n" +
			"                         {command}     : The command to execute (e.g. 'ls -l')   (required)\n" +
			"                         Example: -command ssh -instanceName writer -pemfile ~/.ssh/pem-files/im-dev tail -20 /var/log/writer/writer.log",
		"scp": "Copies files from the specified instance(s). Needs -instanceName or -instanceId, -output and optionally -recursive flags.\n" +
			"                         -instanceName : The aws instance(s) to use as source(s). Operation will occur on all instances with the specific name   (required if instanceId is not specified)\n" +
			"                         -instanceId   : The specific aws instance to use as source.   (required if instanceName is not specified)\n" +
			"                         -pemfile      : The SSH pem file used for authentication    (optional)\n" +
			"                         -output       : the target directory   (required)\n" +
			"                         -recursive    : copies from source recursively\n" +
			"                         Example: -command scp -instanceName writer -pemfile ~/.ssh/pem-files/im-dev -output Documents -recursive /var/log/writer",
		"login": "Log in to instance using SSH\n" +
			"                         -instanceName : The aws instance(s) to use as source(s). Operation will occur on all instances with the specific name   (required if instanceId is not specified)\n" +
			"                         -instanceId   : The specific aws instance to use as source.   (required if instanceName is not specified)\n" +
			"                         -pemfile      : The SSH pem file used for authentication    (optional)",
		"version": "Display writer-tool version.",
	}

	k := sortKeys(m)

	for _, v := range k {
		fmt.Print(v)
		for j := 0; j < 20-(len([]rune(v))); j++ {
			fmt.Print(" ")
		}
		fmt.Println(m[v])
	}

}

func executeCommand() {
	switch command {
	case "copyFileFromS3Bucket":
		if bucket == "" {
			errUsage("s3bucket must be specified")
		}
		if filename == "" {
			errUsage("s3filename must be specified")
		}
		if output == "" {
			errUsage("output must be speficied")
		}
		CopyFileFromS3Bucket(bucket, filename, output)
	case "createReleaseNotes":
		bytes := _readConfigFromFile();
		template := _readTemplateFromFile();
		version := _getVersion()
		dependencies := _readDependenciesFromFile();
		GenerateReleaseNotes(bytes, template, version, releaseDate, dependencies)
	case "createReport":
		bytes := _readConfigFromFile();
		template := _readTemplateFromFile();
		GenerateReport(bytes, template)
	case "deployLambdaFunction":
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
		DeployLambdaFunction(functionName, bucket, filename, alias, version, runtime, publish)
	case "listS3Buckets":
		ListS3Buckets()
	case "listFilesInS3Bucket":
		if bucket == "" {
			errUsage("s3bucket must be specified")
		}
		ListFilesInS3Bucket(bucket, filename)
	case "listClusters":
		ListClusters()
	case "listServices":
		clusterArn := _getClusterArn()
		ListServices(clusterArn)
	case "listTasks":
		clusterArn := _getClusterArn()
		serviceArn := _getServiceArn()
		ListTasks(clusterArn, serviceArn)
	case "describeService":
		clusterArn := _getClusterArn()
		serviceArn := _getServiceArn()
		DescribeService(clusterArn, serviceArn)
	case "describeContainerInstances":
		clusterArn := _getClusterArn()
		DescribeContainerInstances(clusterArn)
	case "updateService":
		clusterArn := _getClusterArn()
		serviceArn := _getServiceArn()
		UpdateService(clusterArn, serviceArn)
	case "updateServices":
		updatesFile := _getUpdatesFile()
		UpdateServices(updatesFile)
	case "releaseService":
		clusterArn := _getClusterArn()
		serviceArn := _getServiceArn()
		version := _getVersion()
		ReleaseService(clusterArn, serviceArn, version)
	case "releaseServices":
		updatesFile := _getUpdatesFile();
		version := _getVersion()
		ReleaseServices(version, updatesFile)
	case "listEc2Instances":
		ListEc2Instances()
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
		if (functionName == "") {
			errUsage("functionName needs to be specified")
		}
		GetLambdaFunctionInfo(functionName)
	case "getLambdaFunctionAliasInfo":
		if (functionName == "") {
			errUsage("functionName needs to be specified")
		}
		if (alias == "") {
			errUsage("alias needs to be specified")
		}
		GetLambdaFunctionAliasInfo(functionName, alias)
	case "version":
		fmt.Println(appVersion);
	case "help":
		printCommandHelp()
	default:
		errUsage("Unknown command: " + command)
	}

}
