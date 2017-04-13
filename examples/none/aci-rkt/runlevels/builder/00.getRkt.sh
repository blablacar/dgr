#!/bin/bash
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x
version=${ACI_VERSION%-*}

url="https://github.com/rkt/rkt/releases/download/v${version}/rkt-v${version}.tar.gz"
curl ${url} -L | tar -C / -xzvf -
mv /rkt*/rkt ${ROOTFS}/usr/bin
mv /rkt*/stage1* ${ROOTFS}/usr/bin
