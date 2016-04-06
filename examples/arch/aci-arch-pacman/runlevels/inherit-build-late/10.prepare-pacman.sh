#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

pacman-key --populate archlinux
mkdir -p ${ROOTFS}/var/lib/pacman
pacman -Sy
