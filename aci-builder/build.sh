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

sudo tar -C ${rootfs}/dgr/ -xf ${dir}/rootfs.tar.xz
sudo cp -R ${dir}/files/. ${rootfs}
sudo chown root: ${rootfs}
cp ${dir}/manifest.json ${target}/manifest
sudo cp --no-preserve=ownership ${dist}/templater ${rootfs}/dgr/usr/bin/

# some cleanup
sudo rm -Rf ${rootfs}/dgr/etc/udev
sudo rm -Rf ${rootfs}/dgr/usr/share/locale
sudo rm -Rf ${rootfs}/dgr/usr/libexec
sudo rm -Rf ${rootfs}/dgr/usr/lib/systemd
sudo rm -Rf ${rootfs}/dgr/usr/lib/udev


sudo mv ${rootfs}/dgr/usr/sbin/haveged ${rootfs}/dgr/usr/bin/haveged
sudo rm -Rf ${rootfs}/dgr/usr/sbin/
sudo bash -c "cd ${rootfs}/dgr/usr && ln -s bin sbin && cd -"

cd ${target}
sudo tar --sort=name --numeric-owner -cpzf ../bindata/aci-builder.aci manifest rootfs \
|| sudo tar -cpzf ../bindata/aci-builder.aci manifest rootfs
sudo chown ${USER}: ../bindata/aci-builder.aci
sudo rm -Rf rootfs/
cd -
