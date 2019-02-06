package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
	"os/user"
	"strings"
	"time"
)

const toolpath = ".writer-tool"

func buildPath(pathElement ...string) string {
	return strings.Join(pathElement[:], ""+string(os.PathSeparator))
}

func getSessionAndConfigForParams(paramProfile string, paramRegion string) (*session.Session, *aws.Config) {
	var sess *session.Session
	var cfg *aws.Config

	if verbose {
		fmt.Printf(
			"Get session and config using profile \"%s\" and region \"%s\"\n",
			paramProfile, paramRegion,
		)
	}

	if paramProfile == "" {
		errUsage("Invalid request. Missing explicit parameter for profile")
	}

	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           paramProfile,
	}))

	if paramRegion != "" {
		cfg = &aws.Config{Region: aws.String(paramRegion)}
	}

	return sess, cfg
}

func getSessionAndConfig() (*session.Session, *aws.Config) {
	var sess *session.Session
	var cfg *aws.Config

	if verbose {
		fmt.Printf(
			"Get session and config using profile \"%s\" and region \"%s\"\n",
			profile, region,
		)
	}

	if profile != "" {
		sess = session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Profile:           profile,
		}))
	} else {
		sess = session.Must(session.NewSession())
	}

	if region != "" {
		cfg = &aws.Config{Region: aws.String(region)}
	}

	return sess, cfg
}

func createDirFromToolkitPath(elements ...string) string {
	currUser, err := user.Current()

	if err != nil {
		errUsage("Error looking up current user")
	}

	var targetPath []string
	targetPath = append(targetPath, currUser.HomeDir, toolpath)
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
	assertError(err)

	//noinspection GoUnhandledErrorResult
	defer f.Close()

	fi, err := f.Stat()
	assertError(err)

	return fi.Mode()
}

func getPemFile() string {
	if profile != "" {
		return getPemfileFromProfile(profile)
	}

	return ""
}
