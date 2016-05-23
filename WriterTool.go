package main

import (
	"fmt"
	"flag"
	"os"
)

var cluster, command, instanceId, instanceName, service, sshPem string
var region = "eu-west-1"

func init() {
	flag.StringVar(&command, "command", "", "The command to use [listClusters, listServices, listEc2Services, ssh]")
	flag.StringVar(&cluster, "cluster", "", "Specify full arn of cluster to use")
	flag.StringVar(&instanceId, "instanceId", "", "Specify the EC2 instance")
	flag.StringVar(&instanceName, "instanceName", "", "Specify the EC2 instance(s) name")
	flag.StringVar(&service, "service", "", "Specify ECS service")
	flag.StringVar(&sshPem, "pemfile", "", "Specify PEM file for SSH acces")
}

func handleError(err error) {
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

}

func errUsage(message string) {
	fmt.Println(message)
	os.Exit(1);
}

func _getClusterArn() string {
	if (cluster == "") {
		errUsage("You must specify a cluster name with: -cluster");
	}

	arn := GetClusterArn(cluster)
	if (arn == "") {
		errUsage("Could not find cluster ARN for name: " + cluster);
	}

	return arn;
}

func _getServiceArn() string {
	if (service == "") {
		errUsage("You must specify a service name with: -service");
	}

	arn := GetServiceArn(service, cluster)
	if (arn == "") {
		errUsage("Could not find service ARN for name: " + service);
	}

	return arn;
}


func main() {
	flag.Parse()

	if (command == "") {
		flag.PrintDefaults();
		return
	}

	switch command {
	case "listClusters":
		ListClusters();
	case "listServices":
		clusterArn := _getClusterArn();
		ListServices(clusterArn)
	case "listEc2Instances":
		ListEc2Instances()
	case "ssh":
		if sshPem == "" {
			errUsage("A SSH PEM file must be specified")
		}
		if instanceId != "" {
			ip := GetIpForInstanceId(instanceId)
			Ssh(ip, sshPem, flag.Args())
		} else if instanceName != "" {
			ips := GetIpsForInstanceName(instanceName)
			if (len(ips) == 1) {
				Ssh(ips[0], sshPem, flag.Args())
			} else {
				for i := 0; i < len(ips); i++ {
					fmt.Printf("[%s]\n", ips[i])
					Ssh(ips[i], sshPem, flag.Args())
				}
			}
		} else {
			errUsage("Either instanceId or instanceName parameter has to be specified")
		}
	default:
		errUsage("Unknown command: " + command)
	}

}

