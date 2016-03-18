#!/bin/bash
set -x
set -e
dir=$( dirname $0 )
target=${dir}/../dist/aci-tester
rootfs=${target}/rootfs

echo -e "\033[0;32mBuilding aci-tester\033[0m\n"

mkdir -p ${rootfs}/dgr/usr/bin ${rootfs}/dgr/usr/lib
cp -R ${dir}/files/. ${rootfs}/
cp ${dir}/manifest.json ${target}/manifest

cp /bin/bash ${rootfs}/dgr/usr/bin
cp --preserve=links /usr/lib/libreadline.so.* ${rootfs}/dgr/usr/lib ||  cp --preserve=links /lib/x86_64-linux-gnu/libreadline.so.* ${rootfs}/dgr/usr/lib
cp --preserve=links /usr/lib/libncursesw.so.* ${rootfs}/dgr/usr/lib ||  cp --preserve=links /lib/x86_64-linux-gnu/libncursesw.so.* ${rootfs}/dgr/usr/lib
wget -O ${rootfs}/dgr/usr/bin/bats https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats
wget -O ${rootfs}/dgr/usr/bin/bats-exec-suite https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats-exec-suite
wget -O ${rootfs}/dgr/usr/bin/bats-exec-test https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats-exec-test
wget -O ${rootfs}/dgr/usr/bin/bats-format-tap-stream https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats-format-tap-stream
wget -O ${rootfs}/dgr/usr/bin/bats-preprocess https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats-preprocess

chmod +x ${rootfs}/dgr/usr/bin/*

cd ${target}
tar cpfz ../bindata/aci-tester.aci rootfs manifest
cd -
