#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

apk add nginx

#apk -X `cat /etc/apk/repositories`  -U --allow-untrusted --root ${ROOTFS} --initdb add alpine-base nginx
