#!/bin/bash
set -e
set -x
dir=$( dirname $0 )

rm -Rf ${dir}/target/

GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o target/darwin-amd64/cnt&
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o target/windows-amd64/cnt.exe
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o target/linux-amd64/cnt
