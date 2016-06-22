#!/usr/bin/env bash

if [[ -z $1 ]]; then
  echo "Missing version"
  echo "Usage: $0 {version}"
  exit 1
fi

if [[ -z ${GOPATH} ]]; then
   echo "There's no GOPATH. Need it to find project dependencies."
   exit 1
fi

rm -rf target
buildroot=target

echo -n "Installing dependencies ... "
go get 
echo "done"

echo -n "Compiling amd64 Linux binary ... "
env GOOS=linux GOARCH=amd64 go build -o ${buildroot}/writer-tool
echo "done"

echo -n "Logging in to docker ... "
eval $(aws ecr --region eu-west-1 get-login --registry-ids 685070497634)
echo "done"

echo -n "Creating docker image ... "
cp Dockerfile target
docker build --no-cache=true -t  685070497634.dkr.ecr.eu-west-1.amazonaws.com/writer-tool:$1 target
docker push 685070497634.dkr.ecr.eu-west-1.amazonaws.com/writer-tool:$1
echo "done"

