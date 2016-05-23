package main

import (
	"fmt"
	"os/exec"
	"bytes"
)

func Ssh(ip string, pemFile string, commands []string) {
	if ip == "" {
		errUsage("No IP number was specified")
	}

	path, err := exec.LookPath("ssh")
	if (err != nil) {
		errUsage("Couldn not find binary 'ssh' in path")
	}

	arguments := make([]string, len(commands) + 3, cap(commands) + 3)
	for i := 0; i < len(commands); i++ {
		arguments[3+i] = commands[i];
	}
	arguments[0] = "-i"
	arguments[1] = pemFile
	arguments[2] = "ec2-user@" + ip

	cmd := exec.Command(path, arguments...)

	var out, stdErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stdErr
	errRun := cmd.Run()
	if (errRun != nil) {
		fmt.Println(stdErr.String())
		fmt.Println(arguments)
		errUsage("* Ssh command failed")
	}

	fmt.Println(out.String())
}