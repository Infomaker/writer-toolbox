#!/usr/bin/env bash
if [[ -z ${GOPATH} ]]; then
   echo "There's no GOPATH. Need it to find project dependencies."
fi

rm -rf target
buildroot=target/writer-toolbox

echo -n "Installing dependencies ... "
go get 
echo "done"

echo -n "Compiling OS X binary ... "
mkdir -p ${buildroot}/osx
go build -o ${buildroot}/osx/writer-tool
echo "done"

#echo -n "Compiling amd64 Linux binary ... "
#mkdir -p ${buildroot}/linux/amd64
#env GOOS=linux GOARCH=amd64 go build -o ${buildroot}/linux/amd64/writer-tool
#echo "done"

#echo -n "Compiling 386 Linux binary ... "
#mkdir -p ${buildroot}/linux/386
#env GOOS=linux GOARCH=386 go build -o ${buildroot}/linux/386/writer-tool
#echo "done"

echo -n "Creating tarball ... "
tar zcf writer-toolbox.tgz target
echo "done"

mv writer-toolbox.tgz target

echo "Tarball is placed in ./target"
