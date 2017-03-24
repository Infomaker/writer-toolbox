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

### Prerequisites
The AWS environment must have a user with enough privileges to perform the operations available in the writer-tool.
Credentials for the user should be stored in a file, credentials.txt, in `~/.aws` directory

```bash
[default]
aws_access_key_id = {AKIA...}
aws_secret_access_key = {DV83453....}
```

Optionally, credentials may be specified using the `awsKey` and `awsSecretKey` options.

```bash
writer-tool -awsKey AKIA... -awsSecretKey DV83453...
```

## How to use

```bash
$ writer-tool -h
```

### Examples

#### List EC2 instances
```bash
$ writer-tool -credentials credentials.txt -command listEc2Instances
i-a6f3482a (52.51.119.155): internaloc-single
i-92e3ca1a (52.51.18.157): writer
i-7f1f8af3 (54.72.32.96): prod-writer
```

#### Describe a service
```bash
$ writer-tool -credentials credentials.txt -cluster dev-EditorService-cluster -service editor-service -command describeService
Name [editor-service], Running: 1, Pending: 0, Desired: 1
```

#### Update a running service
This command stop one task at the time, to force a new download of image from the docker registry. Useful for 
updating a service running _latest_.

```bash
$ writer-tool -credentials credentials.txt -cluster dev-EditorService-cluster -service editor-service -command updateService
Task: arn:aws:ecs:eu-west-1:187317280313:task/6073fd62-70d2-447e-ba9f-bdaf8eee1457
  Stopped
  Waiting for new task .... done
  New task: arn:aws:ecs:eu-west-1:187317280313:task/5bafe279-c8fb-4b8e-bd8b-c69c6ce23fa5 ............ done
Service [editor-service] is updated
```

#### Perform a thread dump on a Editor Service instance
```bash
$ writer-tool -credentials credentials.txt -command ssh -pemfile customer-pemfile.pem -instanceName editorservice 'docker exec $(docker ps -q | head -1) jstack 6' > target/dumps.txt 
$ head -5 target/dumps.txt
2016-05-25 09:17:39
Full thread dump Java HotSpot(TM) 64-Bit Server VM (25.92-b14 mixed mode):
"qtp2001294156-2715" #2715 prio=5 os_prio=0 tid=0x00007f30e85af800 nid=0xb13 waiting on condition [0x00007f30c92f3000]
   java.lang.Thread.State: TIMED_WAITING (parking)
```

#### Perform a curl operation to get HTTP status code from a service, executed on the remote host
```bash
$ writer-tool -command ssh -pemfile customer-pem.pem -instanceName editorservice 'curl --write-out %{http_code} --output /dev/null http://www.sunet.se'
301
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
