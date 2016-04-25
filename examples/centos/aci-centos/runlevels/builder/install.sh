#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

CENTOS_VERSION=${CENTOS_VERSION:-7}

#========================
# Fetch Centos Squash Image
#========================
wget http://mirror.centos.org/centos-${CENTOS_VERSION}/${CENTOS_VERSION}/os/x86_64/LiveOS/squashfs.img

#========================
# Prepare Folder Tree
# and Ramdisk mount
#========================
mkdir -p /tmp/squashfs /tmp/rootfs /tmp/LiveOS/rootfs
mknod /dev/loop0 b 7 0
mount -t squashfs squashfs.img /tmp/squashfs
mount /tmp/squashfs/LiveOS/*.img /tmp/rootfs
cp -a /tmp/rootfs/. /tmp/LiveOS/.


#========================
# Mount ACI - ROOTFS
#========================
mount --bind $ROOTFS /tmp/LiveOS/rootfs
cp -a /etc/resolv.conf /tmp/LiveOS/etc/resolv.conf


#========================
# Fetch and install Centos
# release & Yum
#========================
BASEURL="http://mirror.centos.org/centos-7/7/os/x86_64/Packages/"
chroot /tmp/LiveOS wget -r -l1 --no-parent -A 'yum-3*.noarch.rpm' "${BASEURL}"
chroot /tmp/LiveOS wget -r -l1 --no-parent -A 'centos-release-7*.x86_64.rpm' "${BASEURL}"
chroot /tmp/LiveOS rpm -i --nodeps ${BASEURL//http:\/\//}yum*.rpm
echo 'group_package_types=mandatory' >> /tmp/LiveOS/etc/yum.conf
chroot /tmp/LiveOS rpm --root=/rootfs -i --nodeps ${BASEURL//http:\/\//}centos-release-7*.rpm
chroot /tmp/LiveOS rpm -i --nodeps ${BASEURL//http:\/\//}centos-release-7*.rpm
chroot /tmp/LiveOS sed -i '/distroverpkg=centos-release/a tsflags=nodocs' /etc/yum.conf



#========================
# Install Yum and rpm in
# ACI-ROOTFS
#========================
chroot /tmp/LiveOS yum --installroot=/rootfs update
cp -a /tmp/LiveOS/rootfs/var/lib/rpm /tmp/LiveOS/rootfs/var/lib/rpm.bkp
chroot /tmp/LiveOS yum --installroot=/rootfs fs filter nodocs
chroot /tmp/LiveOS yum --installroot=/rootfs fs filter langs en
chroot /tmp/LiveOS yum --installroot=/rootfs install -y rpm
chroot /tmp/LiveOS yum --installroot=/rootfs install -y yum
chroot /tmp/LiveOS yum --installroot=/rootfs install -y tar
chroot /tmp/LiveOS yum --installroot=/rootfs install -y gzip


#========================
# Clean up.
# ACI size reduction.
#========================
chroot /tmp/LiveOS yum --installroot=/rootfs -y remove  grub2 centos-logos hwdata os-prober gettext*
chroot /tmp/LiveOS yum --installroot=/rootfs -y remove firewalld dbus-glib dbus-python ebtables gobject-introspection libselinux-python pygobject3-base python-decorator python-slip python-slip-dbus
chroot /tmp/LiveOS yum --installroot=/rootfs -y clean all

# locales
# nuking the locales breaks things. Lets not do that anymore
# strip most of the languages from the archive.
chroot /tmp/LiveOS/rootfs  bash -c 'localedef --delete-from-archive $(localedef --list-archive | grep -v -i ^en | xargs )'
# prep the archive template
chroot /tmp/LiveOS/rootfs  mv /usr/lib/locale/locale-archive  /usr/lib/locale/locale-archive.tmpl
# rebuild archive
chroot /tmp/LiveOS/rootfs  /usr/sbin/build-locale-archive
#empty the template
chroot /tmp/LiveOS/rootfs  bash -c  ':>/usr/lib/locale/locale-archive.tmpl'



#  man pages and documentation
chroot /tmp/LiveOS/rootfs bash -c 'find /usr/share/{man,doc,info,gnome/help}  -type f | xargs /bin/rm'

#  sln
chroot /tmp/LiveOS/rootfs rm -f /sbin/sln

#  ldconfig
chroot /tmp/LiveOS/rootfs rm -rf /etc/ld.so.cache
chroot /tmp/LiveOS/rootfs rm -rf /var/cache/ldconfig/*
chroot /tmp/LiveOS/rootfs rm -rf /var/cache/yum/*
