#!/bin/bash
set -e
set -x
start=`date +%s`
dir=$( dirname $0 )

[ -f ${GOPATH}/bin/godep ] || go get github.com/tools/godep
[ -f /usr/bin/upx ] || (echo "upx is required to build dgr" && exit 1)

# clean
rm -Rf ${dir}/dist/*-amd64
mkdir -p ${dir}/dist

#save dep
godep save ${dir}/bin-dgr ${dir}/bin-templater ${dir}/aci-builder/bin-run || true

# format
gofmt -w -s ${dir}/bin-dgr ${dir}/bin-templater ${dir}/aci-builder/bin-run

# bin
mkdir -p ${dir}/dist/bindata/aci/dgrtool.com
[ -f ${dir}/dist/templater ] || ${dir}/bin-templater/build.sh
[ -f ${dir}/dist/bindata/aci-tester.aci ] || ${dir}/aci-tester/build.sh
[ -f ${dir}/dist/bindata/aci-builder.aci ] || ${dir}/aci-builder/build.sh

# binary
[ -f ${GOPATH}/bin/go-bindata ] || go get -u github.com/jteeuwen/go-bindata/...
go-bindata -nomemcopy -pkg dist -o ${dir}/dist/bindata.go ${dir}/dist/bindata/...

echo -e "\033[0;32mBuilding dgr\033[0m\n"

if [ -z ${VERSION} ]; then
    VERSION=0
fi

# build
GOOS=linux GOARCH=amd64 godep go build --ldflags "-s -w -X main.BuildDate=`date -u '+%Y-%m-%d_%H:%M'` \
 -X main.DgrVersion=${VERSION} \
 -X main.CommitHash=`git rev-parse HEAD`" \
 -o ${dir}/dist/linux-amd64/dgr ${dir}/bin-dgr
upx ${dir}/dist/linux-amd64/dgr

# test #TODO move to test
godep go test -cover ${dir}/bin-dgr/... ${dir}/bin-templater/... ${dir}/aci-builder/... ${dir}/aci-tester/...

# install
cp ${dir}/dist/linux-amd64/dgr ${GOPATH}/bin/dgr

end=`date +%s`
echo "Duration : $((end-start))s"
