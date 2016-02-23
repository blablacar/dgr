#!/bin/bash
set -x
set -e
dir=$( dirname $0 )
target=${dir}/../dist/aci-bats
rootfs=${target}/rootfs

echo -e "\033[0;32mBuilding aci-bats\033[0m\n"

mkdir -p ${rootfs}/{bin,usr,usr/bin,usr/lib,lib64,dgr,dgr/bin}
cp -R ${dir}/files/* ${rootfs}/
cp ${dir}/manifest.json ${target}/manifest

cp /bin/bash ${rootfs}/bin/
cp --preserve=links /usr/lib/libreadline.so.* ${rootfs}/usr/lib ||  cp --preserve=links /lib/x86_64-linux-gnu/libreadline.so.* ${rootfs}/usr/lib
cp --preserve=links /usr/lib/libncursesw.so.* ${rootfs}/usr/lib ||  cp --preserve=links /lib/x86_64-linux-gnu/libncursesw.so.* ${rootfs}/usr/lib
cp --preserve=links /usr/lib/libdl.so.* ${rootfs}/usr/lib  ||  cp --preserve=links /lib/x86_64-linux-gnu/libdl.so.* ${rootfs}/usr/lib
cp --preserve=links /usr/lib/libc.so.* ${rootfs}/usr/lib ||  cp --preserve=links /lib/x86_64-linux-gnu/libc.so.* ${rootfs}/usr/lib
cp --preserve=links /lib64/ld-linux-x86-64.so.* ${rootfs}/lib64


wget -O ${rootfs}/dgr/bin/bats https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats
wget -O ${rootfs}/dgr/bin/bats-exec-suite https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats-exec-suite
wget -O ${rootfs}/dgr/bin/bats-exec-test https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats-exec-test
wget -O ${rootfs}/dgr/bin/bats-format-tap-stream https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats-format-tap-stream
wget -O ${rootfs}/dgr/bin/bats-preprocess https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats-preprocess

chmod +x ${rootfs}/dgr/bin/*

cd ${target}
tar cpfz ../bindata/aci-bats.aci rootfs manifest
cd -
