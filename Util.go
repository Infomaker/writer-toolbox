package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
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

func _getSession() *session.Session {
	result, err := session.NewSession()
	assertError(err)

	return result
}

func getSessionAndConfig() (*session.Session, *aws.Config) {
	var sess *session.Session
	var cfg *aws.Config

	if profile != "" {
		// A profile and a credential (and possibly a config file) is expected
		// in order to create a valid session
		if verbose {
			fmt.Printf(
				"Get session and config using profile \"%s\" and region \"%s\"\n",
				profile, region,
			)
		}

		// Note that in this case, no config is necessary
		if region != "" {
			// Use region supplied as parameter
			sess = session.Must(session.NewSessionWithOptions(session.Options{
				Config:            aws.Config{Region: aws.String(region)},
				SharedConfigState: session.SharedConfigEnable,
				Profile:           profile,
			}))
		} else {
			// Use region specified in .aws/config file
			sess = session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
				Profile:           profile,
			}))
		}
	} else if roleArn != "" {
		// Defaults to resolve permission and account by configuration of
		// instance on which writer-tool runs in combination with role_arn.
		if verbose {
			fmt.Printf(
				"Get session and config using roleArn \"%s\"\n",
				roleArn,
			)
		}

		sess = session.Must(session.NewSession())
		cfg = &aws.Config{
			Credentials:                   stscreds.NewCredentials(sess, roleArn),
			LogLevel:                      aws.LogLevel(aws.LogDebugWithHTTPBody),
			CredentialsChainVerboseErrors: aws.Bool(true),
			Region:                        aws.String(region),
		}
	} else {
		// Invalid use of writer-tool
		errUsage("Invalid use of writer-tool. Parameter \"profile\" OR \"roleArn\" must be supplied")
	}

	return sess, cfg
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
	assertError(err)
	defer f.Close()
	fi, err := f.Stat()
	assertError(err)

	return fi.Mode()
}

func getAwsConfig() *aws.Config {
	fmt.Println("We are about to test new config solution!")
	//return &aws.Config{Region: aws.String(region)}
	return &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", "writer-imit"),
	}
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
