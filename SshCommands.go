package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"os"
	"os/exec"
	"syscall"
)

func SshLogin(instance *ec2.Instance, pemFile string) {
	path, err := exec.LookPath("ssh")

	if err != nil {
		errUsage("Could not find binary 'ssh' in path")
	}

	if instance.PublicIpAddress == nil {
		errUsage("No public IP number on instance: " + _getName(instance.Tags))
	}

	if pemFile == "" {
		pemFile = _getPemFile()
	}

	args := []string{"ssh", "-i", pemFile, "ec2-user@" + *instance.PublicIpAddress}
	env := os.Environ()
	execErr := syscall.Exec(path, args, env)

	if execErr != nil {
		errState(execErr.Error())
	}
}

func Ssh(instance *ec2.Instance, pemFile string, commands []string) {
	path, err := exec.LookPath("ssh")

	if err != nil {
		errUsage("Could not find binary 'ssh' in path")
	}

	var arguments []string
	arguments = append(arguments, "-i", pemFile, "ec2-user@"+*instance.PublicIpAddress)
	arguments = append(arguments, commands...)

	doExec(path, arguments, true)
}

func Scp(instance *ec2.Instance, pemFile string, commands []string) {
	path, err := exec.LookPath("scp")

	if err != nil {
		errUsage("Could not find binary 'scp' in path")
	}

	if len(commands) != 1 {
		errUsage("The Scp command requires 1 parameter")
	}

	var arguments []string
	rflag := ""

	if recursive {
		rflag = "-r"
	}

	name := _getName(instance.Tags) + "-" + *instance.InstanceId
	arguments = append(arguments, "-i", pemFile, rflag, "-p", "ec2-user@"+*instance.PublicIpAddress+":"+commands[0])

	if output == "" {
		arguments = append(arguments, CreateDirUsingServerPathWithDate(name))
	} else {
		mode := GetFileMode(output)
		if !mode.IsDir() {
			errUsage("Output '" + output + "' must be directory")
		}

		dir := CreateDir(output, name)
		arguments = append(arguments, dir)
	}

	doExec(path, arguments, false)
}

func doExec(path string, arguments []string, doOutput bool) {
	cmd := exec.Command(path, arguments...)

	var out, stdErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stdErr
	errRun := cmd.Run()

	if errRun != nil {
		fmt.Println(stdErr.String())
		os.Exit(1)
	}

	if doOutput {
		fmt.Println(out.String())
	}
}
