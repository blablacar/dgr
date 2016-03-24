#!/bin/bash
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

cd ${ROOTFS}
mkdir -p usr/lib
ln -s usr/lib lib
ln -s usr/lib lib64

cp -p /dgr/usr/lib/libdl.so.* ${ROOTFS}/usr/lib
cp -p /dgr/usr/lib/libc.so.* ${ROOTFS}/usr/lib
cp -p /dgr/usr/lib/libpthread.so.* ${ROOTFS}/usr/lib
cp -p /dgr/usr/lib/ld-linux-x86-64.so.* ${ROOTFS}/usr/lib
