# writer-toolbox
Tool for managing writer installations

## How to build
    1. Set $GOPATH variable to your go src directory
    2. Issue: ./build.sh

## How to use

writer-tool -h

# Releases

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