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

: ${bats_src:="https://raw.githubusercontent.com/sstephenson/bats/master/libexec"}
curl --fail --silent --show-error --location --remote-time --compressed --create-dirs \
    {-z,-o}"${rootfs}/dgr/usr/bin/bats" \
    ${bats_src}/bats \
  --next \
    {-z,-o}"${rootfs}/dgr/usr/bin/bats-exec-suite" \
    ${bats_src}/bats-exec-suite \
  --next \
    {-z,-o}"${rootfs}/dgr/usr/bin/bats-exec-test" \
    ${bats_src}/bats-exec-test \
  --next \
    {-z,-o}"${rootfs}/dgr/usr/bin/bats-format-tap-stream" \
    ${bats_src}/bats-format-tap-stream \
  --next \
    {-z,-o}"${rootfs}/dgr/usr/bin/bats-preprocess" \
    ${bats_src}/bats-preprocess

chmod +x ${rootfs}/dgr/usr/bin/*

cd ${target}
tar cpfz ../bindata/aci-tester.aci rootfs manifest
cd -
