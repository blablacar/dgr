#!/bin/bash

if [ "$(id -u)" != "0" ]; then
	echo "Sorry, you are not root."
	exit 1
fi

set -x
set -e
dir=$( dirname $0 )
target=${dir}/../dist/aci-root
rootfs=${target}/rootfs

rm -Rf ${rootfs}/
mkdir -p ${rootfs}/

cp ${dir}/manifest.json ${target}/manifest

# buildroot
if [ ! -d ${target}/buildroot ]; then
    cd ${target}
    BR_VERSION=2016.02
    wget -nv http://buildroot.uclibc.org/downloads/buildroot-${BR_VERSION}.tar.gz
    wget -nv http://buildroot.uclibc.org/downloads/buildroot-${BR_VERSION}.tar.gz.sign
    wget -nv http://uclibc.org/~jacmet/pubkey.gpg
    gpg --import pubkey.gpg
    gpg --verify buildroot-${BR_VERSION}.tar.gz.sign
    tar -zxf buildroot-${BR_VERSION}.tar.gz
    rm buildroot-${BR_VERSION}.tar.gz buildroot-${BR_VERSION}.tar.gz.sign pubkey.gpg
    mv buildroot-${BR_VERSION} buildroot
    cd -
fi

#curl
mkdir -p ${target}/buildroot/package/curl
cp -r ${dir}/curl/* ${target}/buildroot/package/curl

# make
cd ${target}/buildroot


#export CXXCPP=/usr/bin/cpp

# config
make defconfig
sed -i 's/BR2_i386=y/BR2_x86_64=y/' .config
#echo BR2_TOOLCHAIN_BUILDROOT_LARGEFILE=y >>.config
#echo BR2_TOOLCHAIN_BUILDROOT_INET_IPV6=y >>.config
#echo BR2_TOOLCHAIN_BUILDROOT_WCHAR=y >>.config
#echo BR2_PACKAGE_LIBCURL=y >>.config
#echo BR2_PACKAGE_CURL=y >>.config
#echo BR2_PACKAGE_OPENSSL=y >>.config
#echo BR2_PACKAGE_LIBSSH2=y >>.config
#echo BR2_PACKAGE_CA_CERTIFICATES=y >>.config
#echo BR2_PACKAGE_ZLIB=y >> .config
#
#echo BR2_PACKAGE_WGET=y >>.config
#
#echo BR2_PACKAGE_LIBCAP=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_LIBBLKID=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_LIBMOUNT=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_LIBUUID=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_BINARIES=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_AGETTY=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_FSCK=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_MOUNT=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_RENAME=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_WALL=y >> .config
#echo BR2_PACKAGE_KMOD=y >> .config
#echo BR2_PACKAGE_SYSTEMD=y >> .config
#
#echo BR2_PACKAGE_HAS_UDEV=y >> .config
#echo BR2_PACKAGE_DBUS=y >> .config
#echo BR2_PACKAGE_LIBCAP=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_LIBBLKID=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_LIBMOUNT=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_BINARIES=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_MOUNT=y >> .config
#echo BR2_PACKAGE_UTIL_LINUX_NOLOGIN=y >> .config
#echo BR2_PACKAGE_KMOD=y >> .config
#echo BR2_PACKAGE_BUSYBOX_SHOW_OTHERS=y >> .config
#echo BR2_PACKAGE_KMOD_TOOLS=y >> .config
#
#echo BR2_PACKAGE_EXPAT=y >> .config
#echo BR2_INSTALL_LIBSTDCPP=y >> .config

#make olddefconfig

#echo BR2_PACKAGE_SYSTEMD=y >> .config
#echo BR2_PACKAGE_PROVIDES_UDEV=systemd >> .config

#cd -
#cp ${dir}/config ${target}/buildroot/.config
#cd ${target}/buildroot


#make olddefconfig
make allnoconfig


# make
make --quiet
tar xf output/images/rootfs.tar -C ../rootfs

cd ..
tar cpfz ../bindata/aci-root.aci rootfs manifest
