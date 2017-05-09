#!/bin/bash
set -x
set -e
dir=$( dirname $0 )
dist=${dir}/../dist
target=${dist}/aci-builder
rootfs=${target}/rootfs

echo -e "\033[0;32mBuilding aci-builder\033[0m\n"

${dir}/clean.sh

mkdir -p ${rootfs}/dgr ${rootfs}/usr/bin
[ -f ${dist}/templater ] || (echo "build templater first" && exit 1)

GOOS=linux GOARCH=amd64 go build --ldflags '-s -w -extldflags "-static"' -o ${rootfs}/dgr/builder/stage1/run ${dir}/bin-run
upx ${rootfs}/dgr/builder/stage1/run

: ${tar:="$(realpath "${dir}/files/dgr/usr/bin/tar")"}

sudo "${tar}" \
  --transform "s:usr/sbin/haveged:usr/bin/haveged:" \
  --exclude "./etc/udev" \
  --exclude "./usr/share/locale" \
  --exclude "./usr/libexec" \
  --exclude "./usr/lib/systemd" \
  --exclude "./usr/lib/udev" \
  --exclude "./usr/sbin" \
  -C ${rootfs}/dgr/ \
  -xf ${dir}/rootfs.tar.xz
sudo cp -R ${dir}/files/. ${rootfs}
sudo chown root: ${rootfs}
cp ${dir}/manifest.json ${target}/manifest
sudo cp --no-preserve=ownership ${dist}/templater ${rootfs}/dgr/usr/bin/

sudo bash -c "cd ${rootfs}/dgr/usr && ln -s bin sbin && cd -"

cd ${target}
sudo "${tar}" --sort=name --numeric-owner \
  -cpzf ../bindata/aci-builder.aci manifest rootfs
sudo chown ${USER}: ../bindata/aci-builder.aci
sudo rm -Rf rootfs/
cd -
