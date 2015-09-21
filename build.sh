#!/bin/bash
set -e
set -x
start=`date +%s`
dir=$( dirname $0 )

[ -f /bin/busybox ] || (echo "/bin/busybox is required to build cnt" && exit 1)

[ -f $GOPATH/bin/godep ] || go get github.com/tools/godep

# clean
rm -Rf $dir/dist/*-amd64

# binary
[ -f $GOPATH/bin/go-bindata ] || go get -u github.com/jteeuwen/go-bindata/...
mkdir -p $dir/dist/bindata
[ -f $dir/dist/bindata/busybox ] || cp /bin/busybox $dir/dist/bindata/busybox
[ -f $dir/dist/bindata/attributes-merger ] || wget "https://github.com/blablacar/attributes-merger/releases/download/0.1/attributes-merger" -O $dir/dist/bindata/attributes-merger
[ -f $dir/dist/bindata/confd ] || wget "https://github.com/kelseyhightower/confd/releases/download/v0.10.0/confd-0.10.0-linux-amd64" -O $dir/dist/bindata/confd
go-bindata -nomemcopy -pkg dist -o $dir/dist/bindata.go $dir/dist/bindata/...

# format && test
gofmt -w -s .
godep go test -cover $dir/...


# build
#GOOS=darwin GOARCH=amd64 godep go build -o dist/darwin-amd64/cnt&
#GOOS=windows GOARCH=amd64 godep go build -o dist/windows-amd64/cnt.exe&
GOOS=linux GOARCH=amd64 godep go build -o $dir/dist/linux-amd64/cnt

# install
cp $dir/dist/linux-amd64/cnt $GOPATH/bin/cnt

end=`date +%s`
echo "Duration : $((end-start))s"
