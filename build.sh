#!/usr/bin/env bash
if [[ -z ${GOPATH} ]]; then
   echo "There's no GOPATH. Need it to find project dependencies."
fi

rm -rf target

echo -n "Installing dependencies ..."
go get 
echo "done"

echo -n "Compiling OS X binary ..."
mkdir -p target/osx
go build -o target/osx/writer-tool
echo "done"

echo -n "Compiling amd64 Linux binary ..."
mkdir -p target/linux/amd64
env GOOS=linux GOARCH=amd64 go build -o target/linux/amd64/writer-tool
echo "done"

echo -n "Compiling 386 Linux binary ..."
mkdir -p target/linux/386
env GOOS=linux GOARCH=386 go build -o target/linux/386/writer-tool
echo "done"

echo "Binaries are found in ./target"