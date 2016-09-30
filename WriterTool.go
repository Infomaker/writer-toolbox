package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var cluster, command, instanceId, instanceName, service, sshPem, output, credentialsFile, profile, awsKey, awsSecretKey, version, loadBalancer, reportJson, reportTemplate, functionName, alias string
var recursive, verbose, moreVerbose bool
var region = "eu-west-1"
var auth *Auth
var verboseLevel = 0

type Auth struct {
	key    string `toml:"aws_access_key_id"`
	secret string `toml:"aws_secret_access_key"`
}

func init() {
	flag.StringVar(&command, "command", "", "The command to use. The command 'help' displays available commands.")
	flag.StringVar(&cluster, "cluster", "", "Specify cluster to use")
	flag.StringVar(&instanceId, "instanceId", "", "Specify the EC2 instance")
	flag.StringVar(&instanceName, "instanceName", "", "Specify the EC2 instance(s) name")
	flag.StringVar(&service, "service", "", "Specify ECS service")
	flag.StringVar(&sshPem, "pemfile", "", "Specify PEM file for SSH access")
	flag.BoolVar(&recursive, "recursive", false, "Specify recursive operation")
	flag.StringVar(&output, "output", "", "Specify output directory")
	flag.StringVar(&credentialsFile, "credentials", "", "Specify credentials used for accessing AWS. Should be of format: .aws/credentials")
	flag.StringVar(&profile, "profile", "", "Specify profile for ./aws/credentials file used for accessing AWS.")
	flag.StringVar(&profile, "p", "", "Specify profile for ./aws/credentials file used for accessing AWS.")
	flag.StringVar(&awsKey, "awsKey", "", "AWS key used for authentication. Overrides credentials file")
	flag.StringVar(&awsSecretKey, "awsSecretKey", "", "AWS secret key used for authentication, used in conjunction with 'awsKey'")
	flag.StringVar(&version, "version", "", "The version to use for docker image in the task definition")
	flag.StringVar(&loadBalancer, "loadBalancer", "", "Specifies the load balancer name to use")
	flag.StringVar(&reportJson, "reportConfig", "", "Filename for the JSON file containing report configuration")
	flag.StringVar(&reportTemplate, "reportTemplate", "", "Filename for the template that produces the report")
	flag.StringVar(&functionName, "functionName", "", "Lambda function name")
	flag.StringVar(&alias, "alias", "", "Lambda alias")
	flag.BoolVar(&verbose, "v", false, "Making output more verbose, where applicable")
	flag.BoolVar(&moreVerbose, "vv", false, "Making output more verbose, where applicable")
}

func printCommandHelp() {
	var m = map[string]string{
		"help":              "Prints this help.",
		"createReport":     "Generates a report of running services. Needs -reportConfig and -reportTemplate",
		"listClusters":      "List available clusters. -v will also list services for all clusters",
		"listServices":      "List available services. Needs -cluster flag.",
		"listTasks":         "List tasks for a service. Needs -cluster, -service flags.",
		"describeService":   "Describes the service. Needs -cluster, -service flags. Optionaly -v and -vv may be used.",
		"updateService":     "Stop/start all running tasks for the specified service. Needs -cluster, -service flags.",
		"releaseService":    "Creates a new release for the service. Neews -cluster, -service, -version flags.",
		"listEc2Instances":  "List available EC2 instances.",
		"listLoadBalancers": "List available Load Balancers and their contained EC2 instances.",
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
			"                         -pemfile      : The SSH pem file used for authentication    (required)\n" +
			"                         -output       : the target directory   (required)\n" +
			"                         -recursive    : copies from source recursively\n" +
			"                         Example: -command scp -instanceName writer -pemfile ~/.ssh/pem-files/im-dev -output Documents -recursive /var/log/writer",
	}

	k := sortKeys(m)

	for _, v := range k {
		fmt.Print(v)
		for j := 0; j < 20 - (len([]rune(v))); j++ {
			fmt.Print(" ")
		}
		fmt.Println(m[v])
	}

}

func sortKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func errUsage(message string) {
	fmt.Println(message)
	os.Exit(1)
}

func errState(message string) {
	fmt.Println(message)
	os.Exit(2)
}

func errStatef(message string, a ...interface{}) {
	fmt.Printf(message, a)
	os.Exit(2)
}

func _readConfigFromFile() []byte {
	if reportJson == "" {
		errUsage("You must specify a report config file with: -reportConfig");
	}

	content, err := ioutil.ReadFile(reportJson)

	if err != nil {
		errState(err.Error())
	}

	return content;
}

func _readTemplateFromFile() string {
	if reportTemplate == "" {
		errUsage("You must specify a report template file with: -reportTemplate");
	}

	content, err := ioutil.ReadFile(reportTemplate)

	if err != nil {
		errState(err.Error())
	}

	return string(content);
}

func _getClusterArn() string {
	if cluster == "" {
		errUsage("You must specify a cluster name with: -cluster")
	}

	arn := GetClusterArn(cluster)
	if arn == "" {
		errUsage("Could not find cluster ARN for name: " + cluster)
	}

	return arn
}

func _getServiceArn() string {
	if cluster == "" {
		errUsage("You must specify a cluster name with: -cluster")
	}

	if service == "" {
		errUsage("You must specify a service name with: -service")
	}

	clusterArn := _getClusterArn()

	serviceArn := GetServiceArn(clusterArn, service)

	return serviceArn
}

func _getVersion() string {
	if version == "" {
		errUsage("You must specify a version with: -version")
	}

	return version
}

func getAwsCredentials(filepath string) (awsAccessKeyId, awsSecretKey string) {
	var awsAccessKeyIdRegexp = regexp.MustCompile("aws_access_key_id\\s*=\\s*(.*)")
	var awsSecretKeyRegexp = regexp.MustCompile("aws_secret_access_key\\s*=\\s*(.*)")

	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		errUsage(err.Error())
	}

	matches := awsAccessKeyIdRegexp.FindStringSubmatch(string(file))
	if len(matches) > 1 {
		awsAccessKeyId = strings.TrimSpace(matches[1])
	} else {
		errUsage("Could not find aws access key id")
	}

	matches = awsSecretKeyRegexp.FindStringSubmatch(string(file))

	if len(matches) > 1 {
		awsSecretKey = strings.TrimSpace(matches[1])
	} else {
		errUsage("Could not find aws secret key")
	}

	return awsAccessKeyId, awsSecretKey
}

func getAwsCredentialsFromProfile(profile string) (awsAccessKeyId, awsSecretKey string) {
	var awsAccessKeyIdRegexp = regexp.MustCompile("\\[" + profile + "\\]\\s*\n\\s*aws_access_key_id\\s*=\\s*(.*)")
	var awsSecretKeyRegexp = regexp.MustCompile("\\[" + profile + "\\]\\s*\n\\s*aws_access_key_id.*\n\\s*aws_secret_access_key\\s*=\\s*(.*)")

	user, err := user.Current()
	if err != nil {
		errState(err.Error())
	}

	file, err := ioutil.ReadFile(filepath.Join(user.HomeDir, ".aws", "credentials"))
	if err != nil {
		errUsage(err.Error())
	}

	matches := awsAccessKeyIdRegexp.FindStringSubmatch(string(file))
	if len(matches) > 1 {
		awsAccessKeyId = strings.TrimSpace(matches[1])
	} else {
		errUsage("Could not find aws access key id")
	}

	matches = awsSecretKeyRegexp.FindStringSubmatch(string(file))

	if len(matches) > 1 {
		awsSecretKey = strings.TrimSpace(matches[1])
	} else {
		errUsage("Could not find aws secret key")
	}

	return awsAccessKeyId, awsSecretKey
}

func main() {
	flag.Parse()

	if verbose {
		verboseLevel = 1
	}
	if moreVerbose {
		verboseLevel = 2
	}

	if command == "" {
		flag.PrintDefaults()
		return
	}

	if credentialsFile != "" {
		key, secret := getAwsCredentials(credentialsFile)
		auth = new(Auth)
		auth.key = key
		auth.secret = secret
	}

	if profile != "" {
		key, secret := getAwsCredentialsFromProfile(profile)
		auth = new(Auth)
		auth.key = key
		auth.secret = secret
	}

	if awsKey != "" && awsSecretKey == "" {
		errUsage("Missing secretKey")
	}

	if awsKey != "" && awsSecretKey != "" {
		auth = new(Auth)
		auth.key = awsKey
		auth.secret = awsSecretKey
	}

	switch command {
	case "createReport":
		bytes := _readConfigFromFile();
		template := _readTemplateFromFile();
		GenerateReport(bytes, template)
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
	case "updateService":
		clusterArn := _getClusterArn()
		serviceArn := _getServiceArn()
		UpdateService(clusterArn, serviceArn)
	case "releaseService":
		clusterArn := _getClusterArn()
		serviceArn := _getServiceArn()
		version := _getVersion()
		ReleaseService(clusterArn, serviceArn, version)
	case "listEc2Instances":
		ListEc2Instances()
	case "listLoadBalancers":
		ListLoadBalancers()
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
	case "help":
		printCommandHelp()
	default:
		errUsage("Unknown command: " + command)
	}

}
