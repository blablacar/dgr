#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

mirror=http://nl.alpinelinux.org/alpine/
version=2.6.8-r1

wget ${mirror}/latest-stable/main/x86_64/apk-tools-static-${version}.apk
tar -xzf apk-tools-static-*.apk
./sbin/apk.static -X ${mirror}/latest-stable/main -U --allow-untrusted --root ${ROOTFS} --initdb add alpine-base
