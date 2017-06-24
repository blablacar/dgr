#!/bin/bash
set -x
set -e
dir=$( dirname $0 )
dist=${dir}/../dist

echo -e "\033[0;32mBuilding bin-templater\033[0m\n"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags '-extldflags "-static"'  -tags static  -o ${dist}/templater ${dir}
upx ${dist}/templater
