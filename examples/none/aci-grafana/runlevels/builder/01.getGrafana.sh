#!/bin/bash
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

VERSION=${ACI_VERSION%-*}
url_grafana="https://grafanarel.s3.amazonaws.com/builds/grafana-$VERSION.linux-x64.tar.gz"

echo ":: Downloading grafana $VERSION"
curl -L $url_grafana  | tar -C ${ROOTFS} -xzf -
mv ${ROOTFS}/grafana* ${ROOTFS}/grafana
