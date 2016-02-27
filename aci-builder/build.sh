#!/bin/bash
set -x
set -e
dir=$( dirname $0 )
dist=${dir}/../dist
target=${dist}/aci-builder
rootfs=${target}/rootfs

echo -e "\033[0;32mBuilding aci-builder\033[0m\n"

mkdir -p ${rootfs}/dgr/bin
[ -f /bin/busybox ] || (echo "/bin/busybox is required to build dgr" && exit 1)
[ -f ${dist}/templater ] || (echo "build templater first" && exit 1)

GOOS=linux GOARCH=amd64 godep go build --ldflags '-s -w -extldflags "-static"' -o ${rootfs}/dgr/builder/stage1/run ${dir}/bin-run
upx ${rootfs}/dgr/builder/stage1/run
#GOOS=linux GOARCH=amd64 godep go build --ldflags '-extldflags "-static"' -o ${rootfs}/dgr/builder/enter ${dir}/bin-enter
#GOOS=linux GOARCH=amd64 godep go build --ldflags '-extldflags "-static"' -o ${rootfs}/dgr/builder/gc ${dir}/bin-gc

cp ${dir}/manifest.json ${target}/manifest
cp -R ${dir}/files/. ${rootfs}
cp /bin/busybox ${rootfs}/dgr/bin
cp ${dist}/templater ${rootfs}/dgr/bin

cd ${target}
tar cpfz ../bindata/aci-builder.aci rootfs manifest
cd -
