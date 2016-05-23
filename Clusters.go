package main

import (
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"fmt"
	"os"
)

func ClusterName(a *string) string {
	re := regexp.MustCompile("\\w+:\\w+:\\w+:[\\w-]+:\\d+:\\w+\\/(.+)")

	res := re.FindAllStringSubmatch(*a, -1);

	if res == nil {
		return "No match for " + *a;
	}

	return res[0][1]
}

func _listClusters() (*ecs.ListClustersOutput) {
	svc := ecs.New(session.New(), &aws.Config{Region: aws.String(region)})

	params := &ecs.ListClustersInput{
		MaxResults: aws.Int64(10),
	}

	resp, err := svc.ListClusters(params)
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

func ListClusters() {
	resp := _listClusters()
	for i := 0; i < len(resp.ClusterArns); i++ {
		name := ClusterName(resp.ClusterArns[i])
		fmt.Println(name)
	}

}

func GetClusterArn(name string) string {
	clusterArns := _listClusters();

	for i := 0; i < len(clusterArns.ClusterArns); i++ {
		arn := clusterArns.ClusterArns[i];
		if (ClusterName(arn) == name) {
			return *arn;
		}
	}

	return "";
}
