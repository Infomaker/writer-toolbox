package main

import (
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"fmt"
	"os"
)

func ServiceName(a *string) string {
	re := regexp.MustCompile("\\w+:\\w+:\\w+:[\\w-]+:\\d+:\\w+\\/(.+)")

	res := re.FindAllStringSubmatch(*a, -1);

	if res == nil {
		return "No match for " + *a;
	}

	return res[0][1]
}

func _listServices(cluster string) (*ecs.ListServicesOutput) {
	svc := ecs.New(session.New(), _getAwsConfig())

	params := &ecs.ListServicesInput{
		Cluster: aws.String(cluster),
		MaxResults: aws.Int64(10),
	}

	resp, err := svc.ListServices(params)
	if (err != nil) {
		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
			os.Exit(1);
		}
	}

	return resp;
}

func ListServices(clusterArn string) {

	resp := _listServices(clusterArn)
	for i := 0; i < len(resp.ServiceArns); i++ {
		name := ServiceName(resp.ServiceArns[i])
		fmt.Println(name)
	}

}

func GetServiceArn(name, clusterArn string) string {
	clusterArns := _listServices(clusterArn);

	for i := 0; i < len(clusterArns.ServiceArns); i++ {
		arn := clusterArns.ServiceArns[i];
		if (ServiceName(arn) == name) {
			return *arn;
		}
	}

	return "";
}
