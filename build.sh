#!/bin/bash
set -e
set -x
dir=$( dirname $0 )

rm -Rf ${dir}/target/
gofmt -w -s .

#GOOS=darwin GOARCH=amd64 godep go build -o dist/darwin-amd64/cnt&
#GOOS=windows GOARCH=amd64 godep go build -o dist/windows-amd64/cnt.exe&
GOOS=linux GOARCH=amd64 godep go build -o dist/linux-amd64/cnt
