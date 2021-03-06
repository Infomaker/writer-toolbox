# writer-toolbox
This is a tool for performing various writer operations from command line. The command contains bash-completions 
for commands and values.

## How to build
    1. Set $GOPATH variable to your go src directory
    2. Issue: ./build.sh

## How to install
```bash
$ brew tap Infomaker/repo
$ brew install writer-toolbox
$ brew install bash-completion
```

## How to release
1. Follow principle of [gitflow](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow) when creating a new release
2. Don´t forget to update this README.md with new release information (bottom of file)
3. Build release using jenkins "build-writer-toolbox-release"
4. Build and push release to homebrew following the instructions found here: https://github.com/Infomaker/homebrew-repo

### Prerequisites
The AWS environment must have a user with enough privileges to perform the operations available in the writer-tool.
Credentials for the user should be stored in a file, credentials.txt, in `~/.aws` directory. By specifying which
"profile" (parameter `-p` or `-profile`), you will use the credentials connected to that profile in the credentials file.

In addition, it is possible to specify a AWS region as part of the profile (must be located on line below 
`aws_secret_access_key` or `pemfile` (if used)). Region will be overridden if writer-tool is executed with parameter 
`-region`. If no region is specified at all, writer-tool defaults to `eu-west-1`.

```bash
[default]
aws_access_key_id = {AKIA...}
aws_secret_access_key = {DV83453....}
region = eu-north-1
```

## How to use

### Display available arguments to writer tool
```bash
$ writer-tool help
```

### Display documentation for available commands and their arguments
```bash
$ writer-tool -command help
```

### Examples

#### Perform a thread dump on a Editor Service instance
```bash
$ writer-tool -p im -command ssh -pemfile customer-pemfile.pem -instanceName editorservice 'docker exec $(docker ps -q | head -1) jstack 6' > target/dumps.txt 
$ head -4 target/dumps.txt
2016-05-25 09:17:39
Full thread dump Java HotSpot(TM) 64-Bit Server VM (25.92-b14 mixed mode):
"qtp2001294156-2715" #2715 prio=5 os_prio=0 tid=0x00007f30e85af800 nid=0xb13 waiting on condition [0x00007f30c92f3000]
   java.lang.Thread.State: TIMED_WAITING (parking)
```

#### Perform a curl operation to get HTTP status code from a service, executed on the remote host
```bash
$ writer-tool -p im -command ssh -pemfile customer-pem.pem -instanceName editorservice 'curl --write-out %{http_code} --output /dev/null http://www.sunet.se'
301
```

#### Login, using ssh, on instance specifying external pem file
```bash
$ writer-tool -p im -command login -instanceId i-06bb6455c11517e54 -pemfile customer-pem.pem
```
Alias for parameter `-pemfile` is `-i`.

#### Generate release notes for a version
```bash
$ curl  -u user:password -X POST -H "Content-Type: application/json" --data '{"jql":"project = WRIT AND fixVersion = 3.0.3","fields":["id","key","issuetype", "summary"]}' https://jira.infomaker.se/rest/api/2/search > issues.json
$ writer-tool -command createReleaseNotes -version 3.0.3 -reportConfig issues.json -reportTemplate someTemplate.filetype -dependenciesFile dependencies.json
```

##### Report config file
The config should look like

```json
{
  "maxResults": 200,
  "issuesUrl": "https://jira.infomaker.se/rest/api/2/search",
  "jql": "project = WRIT AND fixVersion = 3.0",
  "fields": [
    "id",
    "key",
    "issuetype",
    "summary"
  ],
  "username": "...Jira username...",
  "password": "...Jira password..."
}
```

Optionally, the IssueTypesSortOrder array may be specified. Default is

```json
{
  "issuesTypeSortOrder": ["New Feature", "Task", "Bug", "Epic"]
}
```

Optionally, a release overview issue may be specified. This is done by specifying
the `releaseDescriptionLabel` property, defining the name of the label that specifies
the issue that contains the release overview description. Example:

```json
{
  "releaseDescriptionLabel": "release-overview"
}
```

in order to get information from `releaseDescriptionLabel`, the `description` and `labels`
fields must be added to the `fields` config.

##### Dependencies file

The dependencies file specifies the dependencies and their versions for a service for it to function

Syntax:

```json
[
    {
      "dependency" : "Editor Service",
      "versions" : [ "> 3.2", " <= 3.3.1"]
    },
    {
      "dependency" : "Concept backend",
      "versions" : [ ">= 1.0" , "< 2.0"]
    }
]
```

## Releases

    1.0      A service may be updated using the 'updateService' command
    1.1      Builds docker image of tool
    1.1.1    Removed target dir from git
    1.1.4    Bugfix for updating service, check for if there are no tasks
    1.2      Added list loadbalancers command
    1.2.1    Removed linux target, added help for options, added help command
    1.3      Added releaseService command
    1.4      Added describeService command, verbosity (-v and -vv flags). Added command completions
    1.4.1    Added licence
    1.4.2    README fixes
    1.4.3    Preparing for homebrew installation
    1.4.4    Deployment config is displayed for describeService when using -v. Preventing releaseService and updateService if service count > 1
    1.4.5    Updated -command help
    1.5      Added getEntity command. Added verbosity to listLoadBalancers
    1.5.1    Added load balancer to list of options for tab completion
    1.5.2    Added example for command getEntity
    1.5.3    Fixing bug where updateService fails with 'No new tasks'. This operation is retried a couple of times before failing now.
    1.6      Added profile parameter, which specifies a profile in the ~/.aws/credentials file to use as credentials
    1.6.1    Added -p and -profile to bash completion when tool is using itself to get values from AWS
    1.6.2    Extending timeout when waiting for new task
    1.7      Added command createReport
    1.7.1    Extracting image name from docker image in report
    1.8      Added support for AWS_CONFIG_FILE environment variable when using profiles as credentials
    1.9      Added lambda getInfo and deploy commands
    1.9.1    Fix for bash completion bug which did not list function names for -functionName parameter
    1.10     Added S3 operations. Fixed deploy lambda operation, which now puts the deploy version in lambda version description
    1.10.1   Fixed line-wrap issue in bash-completion
    1.10.2   Fixed typo in bash-completion for listFilesInS3Bucket operation
    1.10.3   Fixed typo in bash-completion for listFilesInS3Bucket operation
    1.11     Added lambdas to reporting
    1.11.1   Added description to report for lambda output item
    1.11.2   Fix for version reporting of lambda in report
    1.11.2.1 Bugfix, missing pointer in reporting
    1.11.2.2 Fix for label in lambda output in reporting
    1.11.2.3 Getting the correct lambda description for version
    1.11.3   Fixed docker image build problem
    1.12     Added public URL to verbosity level for listFilesInS3Bucket command
    1.13     Added other services and urls in services in reporting
    1.14     Added parallel update of services
    1.15     Removed cap of list size for ec2instances, clusters and services
    1.16     Fix for panic when listing ec2 instances where public IP is missing
    1.17     Added 'login' command for logging into ec2 instances using ssh
    1.18     Support for releasing multiple services at once and wait for them being completely updated before exising
    1.18.1   Added 'releaseServices' command to completion
    1.18.2   Updated README.md
    1.19     Added createReleaseNotes command
    2.0      Changed createReleaseNotes to read from URL, specified by config file. See README.md for config
    2.1      Added support for multiple container configurations inside a service, when generation report
    2.2      Added support for releasing service defined in task definition with multiple containers
    2.2.1    Release service using json config now supports containerName
    2.2.1.1  Fixed compilation issue for release 2.2.1
    2.3      Added suppport for {{version}} variable in -releaseConfig file
    2.4      Added -login and -password flags for release notes generation
    2.4.1    Encoded html entities are un-escaped when generating release notes
    2.5      Added flag -releaseDate which can be used in release note generation as variable .releaseDate
    2.6      Iterating over result set from AWS until all items are returned, in list functions Added -maxResults to limit the number of items included in response
    2.6.1    Bugfix: Invalid marker when listing load balancers
    2.7      Added describeContainerInstances command
    2.8      Added 'info' to report config
    2.9      Added releaseDescriptionLabel to release notes generation
    2.10     Added 'runtime' to lambda deploy command
    2.11     Bugfix: Added missing function name to call to runtime configration change for lambda function
    2.11.1   Bugfix: Fixed runtime panic when listing files on S3
    2.11.2   Added 'i' (pemfile alias) when executing 'login' command. Also, filtered listing of instanceId on instanceName enabled
    2.11.3   Altered describeService command, verbosity level -vv output format to json
    2.11.4   Added support for different aws regions specified either in the profile or as a command line parameter. Defaults to 'eu-west-1'.
    3.0.0    Breaking changes! Removed possibility to specify credentials as parameters, writer-tool now depends on either usage of a credential file or IAM roles (read more here [https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html). In addition, support for assuming roles has been added. 