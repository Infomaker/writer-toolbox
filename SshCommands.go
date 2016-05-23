package main

import (
	"fmt"
	"os/exec"
	"bytes"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func Ssh(instance *ec2.Instance, pemFile string, commands []string) {
	path, err := exec.LookPath("ssh")
	if (err != nil) {
		errUsage("Couldn not find binary 'ssh' in path")
	}

	var arguments []string

	arguments = append(arguments, "-i", pemFile, "ec2-user@" + *instance.PublicIpAddress)
	arguments = append(arguments, commands...)

	doExec(path, arguments)
}

func Scp(instance *ec2.Instance, pemFile string, commands []string) {
	path, err := exec.LookPath("scp")
	if (err != nil) {
		errUsage("Couldn not find binary 'scp' in path")
	}

	var arguments []string

	arguments = append(arguments, "-i", pemFile, "ec2-user@" + *instance.PublicIpAddress)
	arguments = append(arguments, commands...)
	arguments = append(arguments, CreateServerPathWithDate(*instance.InstanceId))

	doExec(path, arguments)
}

func doExec(path string, arguments []string) {
	cmd := exec.Command(path, arguments...)

	var out, stdErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stdErr
	errRun := cmd.Run()
	if (errRun != nil) {
		fmt.Println(stdErr.String())
	}

	fmt.Println(out.String())
}