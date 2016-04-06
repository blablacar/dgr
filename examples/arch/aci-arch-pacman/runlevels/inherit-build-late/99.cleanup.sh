#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

rm -Rf /var/lib/pacman/sync/*

#find /usr/share ! -name pacman -exec rm -Rf {} \; || true
rm -Rf /usr/include
rm -Rf /usr/lib/libgo.so*
