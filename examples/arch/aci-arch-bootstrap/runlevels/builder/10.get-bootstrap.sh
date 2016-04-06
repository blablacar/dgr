#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

URL=https://archlinux.mirror.pkern.at/iso/latest/
RELEASE_DATE=$(curl -Ls ${URL}|grep -o "[0-9]\{4\}\.[0-9]\{2\}\.[0-9]\{2\}"|head -n1)
FILENAME=archlinux-bootstrap-${RELEASE_DATE}-x86_64.tar.gz
FILE_URL=${URL}/${FILENAME}

wget ${FILE_URL}
tar --strip-components 1 -C ${ROOTFS} -xzf ${FILENAME}
rm ${FILENAME}

# entropy
cd ${ROOTFS}
curl http://archlinux.mirrors.ovh.net/archlinux/extra/os/x86_64/haveged-1.9.1-2-x86_64.pkg.tar.xz > haveged.tar.xz
tar xf haveged.tar.xz
rm haveged.tar.xz