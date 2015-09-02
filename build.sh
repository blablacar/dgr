#!/bin/bash
set -e
set -x
start=`date +%s`
dir=$( dirname $0 )

rm -Rf $dir/target/
gofmt -w -s .
godep go test -cover $dir/...

#GOOS=darwin GOARCH=amd64 godep go build -o dist/darwin-amd64/cnt&
#GOOS=windows GOARCH=amd64 godep go build -o dist/windows-amd64/cnt.exe&
GOOS=linux GOARCH=amd64 godep go build -o dist/linux-amd64/cnt


end=`date +%s`
echo "Duration : $((end-start))s"
