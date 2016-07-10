#!/bin/bash
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

package_name=${ACI_NAME#aci-centos-*}
[ x${package_name} == x"aci-centos" ] && exit 0


#========================
# Prevent issue with lib64
#========================
mkdir -p ${ROOTFS:-"/"}/var/lib ${ROOTFS:-"/"}/usr
mkdir -p ${ROOTFS:-"/"}/usr/lib64
ln -s ${ROOTFS:-"/"}/usr/lib64 /lib64


#========================
# Yum Update.
# With a clean rpm.
#========================
cp -a /var/lib/rpm.bkp ${ROOTFS:-"/"}/var/lib/rpm
yum update
yum --installroot=${ROOTFS:-"/"} update
rpm --root=${ROOTFS:-"/"} --rebuilddb
