#!/bin/bash
set -euxo pipefail
start=$( date +%s )
: ${dir:="$( dirname $0 )"}

if ! command -v godep >/dev/null; then
  go get github.com/tools/godep
fi
if ! command -v upx >/dev/null; then
  >&2 echo "upx is required to build dgr"
  >&2 echo "You can get it from: https://github.com/upx/upx/releases"
  exit 1
fi
if ! command -v go-bindata >/dev/null; then
  go get github.com/jteeuwen/go-bindata
  (cd "$(find ../../../.. -name 'go-bindata' -type d | head -n 1)" \
  && sed -i -e '/^check/s: vet::' testdata/Makefile \
  && make)
fi

# clean
rm -Rf ${dir}/dist/*-amd64
mkdir -p ${dir}/dist

#save dep
godep save ./${dir}/aci-builder/... ./${dir}/bin-templater/... ./${dir}/bin-dgr/... || true

# format
gofmt -w -s ${dir}/bin-dgr ${dir}/bin-templater ${dir}/aci-builder/bin-run

# bin
mkdir -p ${dir}/dist/bindata/aci/blablacar.github.io/dgr
[ -f ${dir}/dist/templater ] || ${dir}/bin-templater/build.sh
[ -f ${dir}/dist/bindata/aci-tester.aci ] || ${dir}/aci-tester/build.sh
[ -f ${dir}/dist/bindata/aci-builder.aci ] || ${dir}/aci-builder/build.sh

# binary
go-bindata -nomemcopy -pkg dist -o ${dir}/dist/bindata.go ${dir}/dist/bindata/...

echo -e "\033[0;32mBuilding dgr\033[0m\n"

: ${VERSION:=0}

# build
GOOS=linux GOARCH=amd64 godep go build --ldflags "-s -w -X main.buildDate=`date -u '+%Y-%m-%d_%H:%M'` \
 -X main.dgrVersion=${VERSION} \
 -X main.commitHash=`git rev-parse HEAD`" \
 -o ${dir}/dist/linux-amd64/dgr ${dir}/bin-dgr
upx ${dir}/dist/linux-amd64/dgr

# test #TODO move to test
godep go test -cover ${dir}/bin-dgr/... ${dir}/bin-templater/... ${dir}/aci-builder/... ${dir}/aci-tester/...

# install
cp ${dir}/dist/linux-amd64/dgr ${dir}/../../../../bin/dgr

end=`date +%s`
echo "Duration : $((end-start))s"
