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