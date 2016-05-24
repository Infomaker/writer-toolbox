package main

import (
	"os"
	"os/user"
	"strings"
	"time"
	"fmt"
)

const toolpath = ".writer-tool"


func createPath(pathElement ...string) string {
	return strings.Join(pathElement[:], "" + string(os.PathSeparator))
}


func createFromToolkitPath(elements ...string) string {
	user, err := user.Current()
	if err != nil {
		errUsage("Error looking up current user")
	}

	var targetPath []string

	targetPath = append(targetPath, user.HomeDir, toolpath)

	targetPath = append(targetPath, elements...)

	path := createPath(targetPath...)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		errUsage(fmt.Sprintf("Problem creating dir [%s]", targetPath))
	}
	return path
}

func CreateServerPathWithDate(server string) string {
	timestamp := time.Now().Format("20060102-150405");

	return createFromToolkitPath(server, timestamp);
}


func GetFileMode(path string) os.FileMode{
	f, err := os.Open(path)
	if err != nil {
		errUsage(err.Error())
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		errUsage(err.Error())
	}
	return fi.Mode()
}