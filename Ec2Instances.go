package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"
)

func _listEc2Instances() *ec2.DescribeInstancesOutput {
	svc := ec2.New(session.New(), _getAwsConfig())

	params := &ec2.DescribeInstancesInput{
	}

	resp, err := svc.DescribeInstances(params)
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return resp
}

func ListEc2Instances() {
	resp := _listEc2Instances()

	for i := 0; i < len(resp.Reservations); i++ {
		for j := 0; j < len(resp.Reservations[i].Instances); j++ {
			instance := resp.Reservations[i].Instances[j]
			if *instance.State.Name == "running" {
				if verboseLevel == 2 {
					// fmt.Printf("%s (%s): %s, %s \n", *instance.InstanceId, *instance.PublicIpAddress, _getName(instance.Tags), *instance.State.Name)

					if (instance.PublicIpAddress != nil) {
						fmt.Printf("%s %s %s: %s \n", tabs(18, *instance.PublicIpAddress), tabs(30, _getName(instance.Tags)), *instance.InstanceId, *instance.State.Name)
					} else if (instance.PrivateIpAddress != nil) {
						fmt.Printf("%s %s %s: %s \n", tabs(18, "(" + *instance.PrivateIpAddress + ")"), tabs(30, _getName(instance.Tags)), *instance.InstanceId, *instance.State.Name)
					} else {
						fmt.Printf("%s, %s: %s \n", _getName(instance.Tags), *instance.InstanceId, *instance.State.Name)
					}
				} else if verboseLevel == 1 {
					fmt.Println(_getName(instance.Tags))
				} else {
					fmt.Println(*instance.InstanceId)
				}
			}
		}
	}
}

func tabs(size int, output string) string {
	return output + strings.Repeat(" ", max(1, size - utf8.RuneCountInString(output)))
}

func max(a, b int) int {
	if (a < b) {
		return b
	}
	return a
}

func GetEntity(loadBalancerId, entityId string) {
	resp := _listLoadBalancers()

	host := ""

	for i := 0; i < len(resp.LoadBalancerDescriptions); i++ {
		if loadBalancerId == *resp.LoadBalancerDescriptions[i].LoadBalancerName {
			host = *resp.LoadBalancerDescriptions[i].DNSName
		}
	}

	if host != "" {
		url := "http://" + host + "/api/newsItem/" + entityId
		if verboseLevel > 0 {
			fmt.Printf("Fetching from url [%s]\n", url)
		}
		resp, err := http.Get(url)

		if err != nil {
			errState(err.Error())
		}

		if resp.StatusCode == 200 {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				errState(err.Error())
			}
			fmt.Println(string(body))
		} else {
			fmt.Println(resp.StatusCode)
		}

		return
	} else {
		errState("No host found for load balancer: " + loadBalancer)
	}

}

func _listLoadBalancers() *elb.DescribeLoadBalancersOutput {
	svc := elb.New(session.New(), _getAwsConfig())

	params := &elb.DescribeLoadBalancersInput{}

	resp, err := svc.DescribeLoadBalancers(params)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return resp
}

func ListLoadBalancers() {
	resp := _listLoadBalancers()

	for i := 0; i < len(resp.LoadBalancerDescriptions); i++ {
		loadBalancer := resp.LoadBalancerDescriptions[i]
		instances := _listEc2Instances()
		if verboseLevel == 0 {
			fmt.Println(*loadBalancer.LoadBalancerName)
		} else if verboseLevel == 1 {
			fmt.Println(*loadBalancer.DNSName)
		} else {
			fmt.Printf("%s (%s)\n", *loadBalancer.LoadBalancerName, *loadBalancer.DNSName)
			for j := 0; j < len(loadBalancer.Instances); j++ {
				instanceItem := loadBalancer.Instances[j]
				for k := 0; k < len(instances.Reservations); k++ {
					//fmt.Printf("Iterating over reservation %d/%d containing %d instances\n", k, len(instances.Reservations), len(instances.Reservations[k].Instances))
					instance := getInstanceForId(instances.Reservations[k].Instances, *instanceItem.InstanceId)
					if instance != nil {
						fmt.Printf("  * %s (%s): %s, %s\n", *instance.InstanceId, *instance.PublicIpAddress, _getName(instance.Tags), *instance.State.Name)
					}
				}
			}
		}
	}
}

func getInstanceForId(instances []*ec2.Instance, instanceId string) *ec2.Instance {
	for i := 0; i < len(instances); i++ {
		//fmt.Println(*instances[i].InstanceId + ", " + instanceId)
		if *instances[i].InstanceId == instanceId {
			return instances[i]
		}
		//fmt.Printf("What is i? %d, what is len? %d\n", i, len(instances))
	}

	return nil
}

func _getName(tags []*ec2.Tag) string {
	for i := 0; i < len(tags); i++ {
		tag := *tags[i]
		if *tag.Key == "Name" {
			return *tag.Value
		}
	}
	return "-"
}

func GetInstanceForId(instanceId string) *ec2.Instance {
	resp := _listEc2Instances()

	for i := 0; i < len(resp.Reservations); i++ {
		for j := 0; j < len(resp.Reservations[i].Instances); j++ {
			instance := resp.Reservations[i].Instances[j]
			if *instance.InstanceId == instanceId {
				return instance
			}
		}
	}

	return nil
}

func GetInstancesForName(name string) []*ec2.Instance {
	resp := _listEc2Instances()

	var result []*ec2.Instance

	for i := 0; i < len(resp.Reservations); i++ {
		for j := 0; j < len(resp.Reservations[i].Instances); j++ {
			instance := resp.Reservations[i].Instances[j]
			if _getName(instance.Tags) == name {
				result = append(result, instance)
			}
		}
	}

	return result
}
