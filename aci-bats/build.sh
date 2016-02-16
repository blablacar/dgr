#!/bin/bash
set -x
set -e

DIR=$( dirname $0 )
ROOTFS=$DIR/rootfs

mkdir -p $ROOTFS/{bin,usr,usr/bin,usr/lib,lib64,dgr,dgr/bin}
cp -R $DIR/files/* $ROOTFS/

cp /bin/bash $ROOTFS/bin/
cp --preserve=links /usr/lib/libreadline.so.* $ROOTFS/usr/lib ||  cp --preserve=links /lib/x86_64-linux-gnu/libreadline.so.* $ROOTFS/usr/lib
cp --preserve=links /usr/lib/libncursesw.so.* $ROOTFS/usr/lib ||  cp --preserve=links /lib/x86_64-linux-gnu/libncursesw.so.* $ROOTFS/usr/lib
cp --preserve=links /usr/lib/libdl.so.* $ROOTFS/usr/lib  ||  cp --preserve=links /lib/x86_64-linux-gnu/libdl.so.* $ROOTFS/usr/lib
cp --preserve=links /usr/lib/libc.so.* $ROOTFS/usr/lib ||  cp --preserve=links /lib/x86_64-linux-gnu/libc.so.* $ROOTFS/usr/lib
cp --preserve=links /lib64/ld-linux-x86-64.so.* $ROOTFS/lib64


wget -O $ROOTFS/dgr/bin/bats https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats
wget -O $ROOTFS/dgr/bin/bats-exec-suite https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats-exec-suite
wget -O $ROOTFS/dgr/bin/bats-exec-test https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats-exec-test
wget -O $ROOTFS/dgr/bin/bats-format-tap-stream https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats-format-tap-stream
wget -O $ROOTFS/dgr/bin/bats-preprocess https://raw.githubusercontent.com/sstephenson/bats/master/libexec/bats-preprocess

chmod +x $ROOTFS/dgr/bin/*

cd $DIR
tar cpfz aci-bats.aci rootfs manifest
cd -
