package main

import (
	"regexp"

	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func ClusterName(a *string) string {
	re := regexp.MustCompile("\\w+:\\w+:\\w+:[\\w-]+:\\d+:\\w+\\/(.+)")

	res := re.FindAllStringSubmatch(*a, -1)

	if res == nil {
		return "No match for " + *a
	}

	return res[0][1]
}

func _listClusters(svc *ecs.ECS) *ecs.ListClustersOutput {
	if svc == nil {
		svc = ecs.New(session.New(), _getAwsConfig())
	}

	params := &ecs.ListClustersInput{
	}

	resp, err := svc.ListClusters(params)
	assertError(err);

	return resp
}

func ListClusters() {
	resp := _listClusters(nil)
	for i := 0; i < len(resp.ClusterArns); i++ {
		name := ClusterName(resp.ClusterArns[i])
		fmt.Println(name)
		if (verboseLevel > 0) {
			servicesResp := _listServices(*resp.ClusterArns[i], nil)
			for j := 0; j < len(servicesResp.ServiceArns); j++ {
				fmt.Println("  " + ClusterName(servicesResp.ServiceArns[j]))
			}
		}
	}

}

func GetClusterArn(name string, svc *ecs.ECS) string {
	clusterArns := _listClusters(svc)

	for i := 0; i < len(clusterArns.ClusterArns); i++ {
		arn := clusterArns.ClusterArns[i]
		if ClusterName(arn) == name {
			return *arn
		}
	}

	return ""
}
