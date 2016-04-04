#!/bin/bash
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x
version=${ACI_VERSION%-*}

url="https://github.com/prometheus/prometheus/releases/download/${version}/prometheus-${version}.linux-amd64.tar.gz"
curl ${url} -L | tar -C ${ROOTFS} -xzvf -
mv ${ROOTFS}/prometheus* ${ROOTFS}/prometheus
chown -R 0:0 ${ROOTFS}/prometheus
