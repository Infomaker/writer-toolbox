package main

import (
	"fmt"
	"flag"
	"os"
	"io/ioutil"
	"strings"
	"regexp"
)

var cluster, command, instanceId, instanceName, service, sshPem, output, credentialsFile, awsKey, awsSecretKey string
var recursive bool
var region = "eu-west-1"
var auth *Auth

type Auth struct {
	key string `toml:"aws_access_key_id"`
	secret string `toml:"aws_secret_access_key"`
}

func init() {
	flag.StringVar(&command, "command", "", "The command to use [listClusters, listServices, listTasks, updateService, listEc2Instances, listLoadBalancers, ssh, scp]")
	flag.StringVar(&cluster, "cluster", "", "Specify cluster to use")
	flag.StringVar(&instanceId, "instanceId", "", "Specify the EC2 instance")
	flag.StringVar(&instanceName, "instanceName", "", "Specify the EC2 instance(s) name")
	flag.StringVar(&service, "service", "", "Specify ECS service")
	flag.StringVar(&sshPem, "pemfile", "", "Specify PEM file for SSH access")
	flag.BoolVar(&recursive, "recursive", false, "Specify recursive operation")
	flag.StringVar(&output, "output", "", "Specify output directory")
	flag.StringVar(&credentialsFile, "credentials", "", "Specify credentials used for accessing AWS. Should be of format: .aws/credentials")
	flag.StringVar(&awsKey, "awsKey", "", "AWS key used for authentication. Overrides credentials file")
	flag.StringVar(&awsSecretKey, "awsSecretKey", "", "AWS secret key used for authentication, used in conjunction with 'awsKey'")
}

func errUsage(message string) {
	fmt.Println(message)
	os.Exit(1);
}

func errState(message string) {
	fmt.Println(message)
	os.Exit(2)
}

func errStatef(message string, a ...interface{}) {
	fmt.Printf(message, a)
	os.Exit(2)
}


func _getClusterArn() string {
	if (cluster == "") {
		errUsage("You must specify a cluster name with: -cluster");
	}

	arn := GetClusterArn(cluster)
	if (arn == "") {
		errUsage("Could not find cluster ARN for name: " + cluster);
	}

	return arn
}

func _getServiceArn() string {
	if (cluster == "") {
		errUsage("You must specify a cluster name with: -cluster");
	}

	if (service == "") {
		errUsage("You must specify a service name with: -service");
	}

	clusterArn := _getClusterArn();

	serviceArn := GetServiceArn(clusterArn, service)

	return serviceArn
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

func main() {
	flag.Parse()

	if (command == "") {
		flag.PrintDefaults();
		return
	}

	if (credentialsFile != "") {
		key, secret := getAwsCredentials(credentialsFile)
		auth = new(Auth)
		auth.key = key
		auth.secret = secret
	}

	if (awsKey != "" && awsSecretKey == "" ) {
		errUsage("Missing secretKey")
	}

	if (awsKey != "" && awsSecretKey != "" ) {
		auth = new(Auth)
		auth.key = awsKey
		auth.secret = awsSecretKey
	}

	switch command {
	case "listClusters":
		ListClusters();
	case "listServices":
		clusterArn := _getClusterArn()
		ListServices(clusterArn)
	case "listTasks":
		clusterArn := _getClusterArn()
		serviceArn := _getServiceArn()
		ListTasks(clusterArn, serviceArn)
	case "updateService":
		clusterArn := _getClusterArn()
		serviceArn := _getServiceArn()
		UpdateService(clusterArn, serviceArn)
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
			if (len(instances) == 1) {
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
			if (len(ips) == 1) {
				Scp(ips[0], sshPem, flag.Args())
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
	default:
		errUsage("Unknown command: " + command)
	}

}

