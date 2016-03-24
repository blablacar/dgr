#!/bin/bash
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

url="https://github.com/PrFalken/prometheus/releases/download/0.16.1%2Bnerve2/prometheus-0.16.1.nerve2.linux-amd64.tar.gz"
PROGRAM_PATH="$ROOTFS/etc/prometheus"
mkdir -p ${PROGRAM_PATH}
curl ${url} -L | tar -C ${PROGRAM_PATH} -xzvf -
chown -R 0:0 ${PROGRAM_PATH}
