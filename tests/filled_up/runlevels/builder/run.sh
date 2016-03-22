#!/dgr/bin/busybox sh
set -e

wget https://google.fr -O ${ROOTFS}/google.wget
curl https://google.fr -o ${ROOTFS}/google.curl

cd ${ROOTFS}/
tar cvzf res.tar.gz google.wget google.curl

