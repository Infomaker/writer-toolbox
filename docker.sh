#!/usr/bin/env bash

if [[ -z $1 ]]; then
  echo "Missing version"
  echo "Usage: $0 {version}"
  exit 1
fi


rm -rf target
mkdir target

echo -n "Building writer toolbox ... "
docker run -i --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 golang:1.7 bash -c 'go get -d -v ; go build -ldflags -s -v -o target/writer-tool'
echo "done"

cp Dockerfile target


