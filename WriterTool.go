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

// Build version variables
var appVersion string

var cluster, command, containerName, instanceId, instanceName, service, sshPem,
output, profile, version, loadBalancer, reportJson, releaseDate, reportTemplate,
runtime, functionName, alias, bucket, filename, publish, updatesFile,
dependenciesFile, login, region, password, roleArn string

var recursive, verbose, moreVerbose bool
var verboseLevel = 0
var maxResult int64

func init() {
	flag.Int64Var(&maxResult, "maxResults", 100, "Max items to return in list operations")
	flag.StringVar(&alias, "alias", "", "Lambda alias")
	flag.StringVar(&bucket, "s3bucket", "", "The S3 bucket name.")
	flag.StringVar(&containerName, "containerName", "", "The name of the container inside a task definition.")
	flag.StringVar(&filename, "s3filename", "", "The S3 filename.")
	flag.StringVar(&publish, "publish", "false", "Specifies if lambda function deployment should be published.")
	flag.StringVar(&command, "command", "", "The command to use. The command 'help' displays available commands.")
	flag.StringVar(&cluster, "cluster", "", "Specify cluster to use")
	flag.StringVar(&functionName, "functionName", "", "Lambda function name")
	flag.StringVar(&runtime, "runtime", "", "Runtime for lambda function, see: https://docs.aws.amazon.com/cli/latest/reference/lambda/update-function-configuration.html. Example: 'nodejs8.10'")
	flag.StringVar(&instanceId, "instanceId", "", "Specify the EC2 instance")
	flag.StringVar(&instanceName, "instanceName", "", "Specify the EC2 instance(s) name")
	flag.StringVar(&service, "service", "", "Specify ECS service")
	flag.StringVar(&sshPem, "pemfile", "", "Specify PEM file for SSH access")
	flag.StringVar(&sshPem, "i", "", "Specify PEM file for SSH access")
	flag.BoolVar(&recursive, "recursive", false, "Specify recursive operation")
	flag.StringVar(&output, "output", "", "Specify output directory")
	flag.StringVar(&login, "login", "", "Specify login for external service")
	flag.StringVar(&password, "password", "", "Specify password for external service")
	flag.StringVar(&profile, "profile", "", "Specify profile for ./aws/credentials file used for accessing AWS.")
	flag.StringVar(&profile, "p", "", "Specify profile for ./aws/credentials file used for accessing AWS.")
	flag.StringVar(&version, "version", "", "The version to use for docker image in the task definition")
	flag.StringVar(&releaseDate, "releaseDate", "", "The date for a release, used in release notes generation")
	flag.StringVar(&loadBalancer, "loadBalancer", "", "Specifies the load balancer name to use")
	flag.StringVar(&reportJson, "reportConfig", "", "Filename for the JSON file containing report configuration")
	flag.StringVar(&reportTemplate, "reportTemplate", "", "Filename for the template that produces the report")
	flag.StringVar(&updatesFile, "updatesFile", "", "File containing services to update. JSON formatted")
	flag.StringVar(&dependenciesFile, "dependenciesFile", "", "File containing service dependencies. JSON formatted")
	flag.BoolVar(&verbose, "v", false, "Making output more verbose, where applicable")
	flag.BoolVar(&moreVerbose, "vv", false, "Making output more verbose, where applicable")
	flag.StringVar(&region, "region", "", "The region to use")
	flag.StringVar(&roleArn, "roleArn", "", "ARN of the role to assume when executing AWS command")
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

func readConfigFromFile() []byte {
	if reportJson == "" {
		errUsage("You must specify a report config file with: -reportConfig")
	}

	content, err := ioutil.ReadFile(reportJson)
	assertError(err)

	return content
}

func readTemplateFromFile() string {
	if reportTemplate == "" {
		errUsage("You must specify a report template file with: -reportTemplate")
	}

	content, err := ioutil.ReadFile(reportTemplate)
	assertError(err)

	return string(content)
}

func readDependenciesFromFile() string {
	if dependenciesFile == "" {
		return ""
	}

	content, err := ioutil.ReadFile(dependenciesFile)
	assertError(err)

	return string(content)
}

func getClusterArn() string {
	if cluster == "" {
		errUsage("You must specify a cluster name with: -cluster")
	}

	arn := GetClusterArn(cluster, nil)
	if arn == "" {
		errUsage("Could not find cluster ARN for name: " + cluster)
	}

	return arn
}

func getServiceArn() string {
	if cluster == "" {
		errUsage("You must specify a cluster name with: -cluster")
	}

	if service == "" {
		errUsage("You must specify a service name with: -service")
	}

	clusterArn := getClusterArn()
	serviceArn := GetServiceArn(clusterArn, service, nil)

	return serviceArn
}

func getVersion() string {
	if version == "" {
		errUsage("You must specify a version with: -version")
	}

	return version
}

func getPemfileFromProfile(profile string) string {
	var pemregex = regexp.MustCompile("\\[\\s*" + profile + "\\s*\\]\\s*\n\\s*aws_access_key_id.*\n\\s*aws_secret_access_key.*\n\\s*pemfile\\s*=\\s*(.*)")

	currUser, err := user.Current()
	assertError(err)

	var path string
	if os.Getenv("AWS_CONFIG_FILE") != "" {
		path = os.Getenv("AWS_CONFIG_FILE")
	} else {
		path = filepath.Join(currUser.HomeDir, ".aws", "credentials")
	}

	file, err := ioutil.ReadFile(path)
	assertError(err)

	matches := pemregex.FindStringSubmatch(string(file))
	if len(matches) < 2 {
		errUsage("Could not find pemfile in config file, please specify with -pemfile")
	}

	return strings.TrimSpace(matches[1])
}

func getUpdatesFile() []byte {
	if updatesFile == "" {
		errUsage("An updates file needs to be provided with -updatesFile")
	}

	file, err := ioutil.ReadFile(updatesFile)
	assertError(err)

	return file
}

func assertError(err error) {
	if err != nil {
		errState(err.Error())
	}
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

	executeCommand()
}
