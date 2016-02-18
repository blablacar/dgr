#!/bin/bash
set -e
. $TARGET/rootfs/dgr/bin/functions.sh
isLevelEnabled "debug" && set -x


cd $TARGET/rootfs/
mkdir -p usr/lib
ln -s usr/lib lib
ln -s usr/lib lib64

cp --preserve=links /usr/lib/libdl.so.* $TARGET/rootfs/usr/lib  ||  cp --preserve=links /lib/x86_64-linux-gnu/libdl.so.* $TARGET/rootfs/usr/lib
cp --preserve=links /usr/lib/libc.so.* $TARGET/rootfs/usr/lib ||  cp --preserve=links /lib/x86_64-linux-gnu/libc.so.* $TARGET/rootfs/usr/lib
cp --preserve=links /usr/lib/libpthread.so.* $TARGET/rootfs/usr/lib ||  cp --preserve=links /lib/x86_64-linux-gnu/libpthread.so.* $TARGET/rootfs/usr/lib
cp --preserve=links /lib64/ld-linux-x86-64.so.* $TARGET/rootfs/lib64
