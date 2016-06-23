#!/usr/bin/env bash

if [[ -z $1 ]]; then
  echo "Missing version"
  echo "Usage: $0 {version}"
  exit 1
fi


rm -rf target
mkdir target

echo -n "Building writer toolbox ... "
docker run -it --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 golang:1.6 bash -c "go get -d -v; go build -ldflags '-s' -v -o target/writer-tool"
echo "done"

cp Dockerfile target

echo -n "Creating docker image ... "
docker build --no-cache=true -t  685070497634.dkr.ecr.eu-west-1.amazonaws.com/writer-tool:$1 target

echo -n "Logging in to docker ... "
eval $(aws ecr --profile im-docker-push --region eu-west-1 get-login --registry-ids 685070497634)
echo "done"

echo -n "Pusing to registry"
docker push 685070497634.dkr.ecr.eu-west-1.amazonaws.com/writer-tool:$1
echo "done"

