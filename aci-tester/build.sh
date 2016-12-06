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
    ${bats_src}/bats
curl --fail --silent --show-error --location --remote-time --compressed --create-dirs \
    {-z,-o}"${rootfs}/dgr/usr/bin/bats-exec-suite" \
    ${bats_src}/bats-exec-suite
curl --fail --silent --show-error --location --remote-time --compressed --create-dirs \
    {-z,-o}"${rootfs}/dgr/usr/bin/bats-exec-test" \
    ${bats_src}/bats-exec-test
curl --fail --silent --show-error --location --remote-time --compressed --create-dirs \
    {-z,-o}"${rootfs}/dgr/usr/bin/bats-format-tap-stream" \
    ${bats_src}/bats-format-tap-stream
curl --fail --silent --show-error --location --remote-time --compressed --create-dirs \
    {-z,-o}"${rootfs}/dgr/usr/bin/bats-preprocess" \
    ${bats_src}/bats-preprocess

chmod +x ${rootfs}/dgr/usr/bin/*

: ${tar:="$(realpath "${dir}/../aci-builder/files/dgr/usr/bin/tar")"}
cd ${target}
"${tar}" --sort=name --numeric-owner \
  --owner=0 --group=0 \
  -cpzf ../bindata/aci-tester.aci manifest rootfs
cd -
