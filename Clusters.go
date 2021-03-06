package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ecs"
	"regexp"
)

func ClusterName(a *string) string {
	//noinspection RegExpRedundantEscape
	re := regexp.MustCompile("\\w+:\\w+:\\w+:[\\w-]+:\\d+:\\w+\\/(.+)")
	res := re.FindAllStringSubmatch(*a, -1)

	if res == nil {
		return "No match for " + *a
	}

	return res[0][1]
}

func ListClusters() {
	resp := listClusters(nil)

	for i := 0; i < len(resp.ClusterArns); i++ {
		name := ClusterName(resp.ClusterArns[i])
		fmt.Println(name)

		if verboseLevel > 0 {
			servicesResp := listServices(*resp.ClusterArns[i], nil)
			for j := 0; j < len(servicesResp.ServiceArns); j++ {
				fmt.Println("  " + ClusterName(servicesResp.ServiceArns[j]))
			}
		}
	}

}

func GetClusterArn(name string, svc *ecs.ECS) string {
	clusterArns := listClusters(svc)

	for i := 0; i < len(clusterArns.ClusterArns); i++ {
		arn := clusterArns.ClusterArns[i]

		if ClusterName(arn) == name {
			return *arn
		}
	}

	return ""
}

func listClusters(svc *ecs.ECS) *ecs.ListClustersOutput {
	if svc == nil {
		sess, cfg := getSessionAndConfig()
		svc = ecs.New(sess, cfg)
	}

	var marker = new(string)
	var result = new(ecs.ListClustersOutput)

	for marker != nil && len(result.ClusterArns) < int(maxResult) {
		if marker != nil && *marker == "" {
			marker = nil
		}

		params := &ecs.ListClustersInput{
			NextToken: marker,
			MaxResults: &maxResult,
		}

		resp, err := svc.ListClusters(params)
		assertError(err)

		result.ClusterArns = append(result.ClusterArns, resp.ClusterArns...)
		marker = result.NextToken
	}

	return result
}
