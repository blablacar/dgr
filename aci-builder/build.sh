#!/bin/bash
set -x
set -e
dir=$( dirname $0 )
dist=${dir}/../dist
target=${dist}/aci-builder

echo -e "\033[0;32mBuilding aci-builder\033[0m\n"

mkdir -p ${target}/rootfs/dgr/bin
[ -f /bin/busybox ] || (echo "/bin/busybox is required to build dgr" && exit 1)
[ -f ${dist}/templater ] || (echo "build templater first" && exit 1)

GOOS=linux GOARCH=amd64 godep go build --ldflags '-s -w -extldflags "-static"' -o ${target}/rootfs/dgr/builder/run ${dir}/bin-run
upx ${target}/rootfs/dgr/builder/run
#GOOS=linux GOARCH=amd64 godep go build --ldflags '-extldflags "-static"' -o ${target}/rootfs/dgr/builder/enter ${dir}/bin-enter
#GOOS=linux GOARCH=amd64 godep go build --ldflags '-extldflags "-static"' -o ${target}/rootfs/dgr/builder/gc ${dir}/bin-gc

cp ${dir}/manifest.json ${target}/manifest
cp -R ${dir}/files/* ${target}/rootfs
cp /bin/busybox ${target}/rootfs/dgr/bin
cp ${dist}/templater ${target}/rootfs/dgr/bin

cd ${target}
tar cpfz ../bindata/aci-builder.aci rootfs manifest
cd -
