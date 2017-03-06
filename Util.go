package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"os"
	"os/user"
	"strings"
	"time"
)

const toolpath = ".writer-tool"

func buildPath(pathElement ...string) string {
	return strings.Join(pathElement[:], ""+string(os.PathSeparator))
}

func createDirFromToolkitPath(elements ...string) string {
	user, err := user.Current()
	if err != nil {
		errUsage("Error looking up current user")
	}

	var targetPath []string

	targetPath = append(targetPath, user.HomeDir, toolpath)

	targetPath = append(targetPath, elements...)

	path := buildPath(targetPath...)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		errUsage(fmt.Sprintf("Problem creating dir [%s]", targetPath))
	}
	return path
}

func CreateDirUsingServerPathWithDate(server string) string {
	timestamp := time.Now().Format("20060102-150405")

	return createDirFromToolkitPath(server, timestamp)
}

func CreateDir(source, target string) string {

	path := buildPath(source, target)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		errUsage(fmt.Sprintf("Problem creating dir [%s]", path))
	}

	return path
}

func GetFileMode(path string) os.FileMode {
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

func _getAwsConfig() *aws.Config {
	if auth != nil {
		return &aws.Config{Region: aws.String(region), Credentials: credentials.NewStaticCredentials(auth.key, auth.secret, "")}
	}
	return &aws.Config{Region: aws.String(region)}
}


func _getPemFile() string {
	if profile != "" {
		return getPemfileFromProfile(profile)
	}
	return ""
}