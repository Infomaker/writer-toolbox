#!/usr/bin/env bash

if [[ -z $1 ]]; then
  echo "Missing version"
  echo "Usage: $0 {version}"
  exit 1
fi


rm -rf target
mkdir target

echo -n "Building writer toolbox ... "
docker run -it --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 golang:1.6 bash -c "go get -d ; go build -ldflags '-s' -o target/writer-tool"
echo "done"

cp Dockerfile target


