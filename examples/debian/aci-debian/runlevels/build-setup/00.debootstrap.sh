#!/bin/bash
set -e
. $TARGET/rootfs/dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

dir=$(dirname $0)
old_dir=$(pwd)
CHROOTFS=${TARGET}/rootfs

cd $TARGET

command -v /usr/sbin/debootstrap >/dev/null 2>&1 || { echo >&2 "debootstrap is required to build."; exit 1; }

echo 'Debootstraping new Jessie image'
LANG=C /usr/sbin/debootstrap --arch=amd64 --variant=minbase jessie ${CHROOTFS} 2>&1>/dev/null

echo 'Cleaning rootfs'

find ${CHROOTFS} -name '*.deb' -exec rm {} \;
for logfile in $(find ${CHROOTFS} -name '*.log')
do
  > $logfile
done
for rootdir in ${CHROOTFS}/var/lib/apt/lists/ \
               ${CHROOTFS}/usr/share/locale/ \
               ${CHROOTFS}/usr/share/doc \
               ${CHROOTFS}/usr/share/man; do
  find ${rootdir} -type f -exec rm {} \;
done
rm -f  ${CHROOTFS}/var/cache/apt/*.bin

ln -sf /proc/mounts ${CHROOTFS}/etc/mtab
