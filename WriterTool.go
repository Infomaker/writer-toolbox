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

var cluster, command, containerName, instanceId, instanceName, service, sshPem, output, credentialsFile, profile,
awsKey, awsSecretKey, version, loadBalancer, reportJson, releaseDate, reportTemplate, functionName, alias,
bucket, filename, publish, updatesFile, dependenciesFile, login, password string
var recursive, verbose, moreVerbose bool
var region = "eu-west-1"
var auth *Auth
var verboseLevel = 0
var maxResult int64

type Auth struct {
	key    string `toml:"aws_access_key_id"`
	secret string `toml:"aws_secret_access_key"`
}

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
	flag.StringVar(&instanceId, "instanceId", "", "Specify the EC2 instance")
	flag.StringVar(&instanceName, "instanceName", "", "Specify the EC2 instance(s) name")
	flag.StringVar(&service, "service", "", "Specify ECS service")
	flag.StringVar(&sshPem, "pemfile", "", "Specify PEM file for SSH access")
	flag.BoolVar(&recursive, "recursive", false, "Specify recursive operation")
	flag.StringVar(&output, "output", "", "Specify output directory")
	flag.StringVar(&login, "login", "", "Specify login for external service")
	flag.StringVar(&password, "password", "", "Specify password for external service")
	flag.StringVar(&credentialsFile, "credentials", "", "Specify credentials used for accessing AWS. Should be of format: .aws/credentials")
	flag.StringVar(&profile, "profile", "", "Specify profile for ./aws/credentials file used for accessing AWS.")
	flag.StringVar(&profile, "p", "", "Specify profile for ./aws/credentials file used for accessing AWS.")
	flag.StringVar(&awsKey, "awsKey", "", "AWS key used for authentication. Overrides credentials file")
	flag.StringVar(&awsSecretKey, "awsSecretKey", "", "AWS secret key used for authentication, used in conjunction with 'awsKey'")
	flag.StringVar(&version, "version", "", "The version to use for docker image in the task definition")
	flag.StringVar(&releaseDate, "releaseDate", "", "The date for a release, used in release notes generation")
	flag.StringVar(&loadBalancer, "loadBalancer", "", "Specifies the load balancer name to use")
	flag.StringVar(&reportJson, "reportConfig", "", "Filename for the JSON file containing report configuration")
	flag.StringVar(&reportTemplate, "reportTemplate", "", "Filename for the template that produces the report")
	flag.StringVar(&updatesFile, "updatesFile", "", "File containing services to update. JSON formatted")
	flag.StringVar(&dependenciesFile, "dependenciesFile", "", "File containing service dependencies. JSON formatted")
	flag.BoolVar(&verbose, "v", false, "Making output more verbose, where applicable")
	flag.BoolVar(&moreVerbose, "vv", false, "Making output more verbose, where applicable")
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

func _readConfigFromFile() []byte {
	if reportJson == "" {
		errUsage("You must specify a report config file with: -reportConfig");
	}

	content, err := ioutil.ReadFile(reportJson)

	assertError(err);

	return content;
}

func _readTemplateFromFile() string {
	if reportTemplate == "" {
		errUsage("You must specify a report template file with: -reportTemplate");
	}

	content, err := ioutil.ReadFile(reportTemplate)

	assertError(err);

	return string(content);
}

func _readDependenciesFromFile() string {
	if dependenciesFile == "" {
		return "";
	}

	content, err := ioutil.ReadFile(dependenciesFile)

	assertError(err);

	return string(content);
}

func _getClusterArn() string {
	if cluster == "" {
		errUsage("You must specify a cluster name with: -cluster")
	}

	arn := GetClusterArn(cluster, nil)
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

	serviceArn := GetServiceArn(clusterArn, service, nil)

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
	assertError(err);

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

	currUser, err := user.Current()
	assertError(err);

	var path string
	if (os.Getenv("AWS_CONFIG_FILE") != "") {
		path = os.Getenv("AWS_CONFIG_FILE");
	} else {
		path = filepath.Join(currUser.HomeDir, ".aws", "credentials")
	}
	file, err := ioutil.ReadFile(path)
	assertError(err);

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

func getPemfileFromProfile(profile string) string {
	var pemregex = regexp.MustCompile("\\[\\s*" + profile + "\\s*\\]\\s*\n\\s*aws_access_key_id.*\n\\s*aws_secret_access_key.*\n\\s*pemfile\\s*=\\s*(.*)")

	currUser, err := user.Current()
	assertError(err);

	var path string
	if (os.Getenv("AWS_CONFIG_FILE") != "") {
		path = os.Getenv("AWS_CONFIG_FILE");
	} else {
		path = filepath.Join(currUser.HomeDir, ".aws", "credentials")
	}
	file, err := ioutil.ReadFile(path)
	assertError(err);

	matches := pemregex.FindStringSubmatch(string(file))
	if len(matches) < 2 {
		errUsage("Could not find pemfile in config file, please specify with -pemfile")
	}

	return strings.TrimSpace(matches[1])
}

func _getUpdatesFile() []byte {
	if updatesFile == "" {
		errUsage("An updates file needs to be provided with -updatesFile")
	}

	file, err := ioutil.ReadFile(updatesFile)

	assertError(err);

	return file;
}

func UpdateCredentials() {

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

	UpdateCredentials()

	executeCommand()
}
