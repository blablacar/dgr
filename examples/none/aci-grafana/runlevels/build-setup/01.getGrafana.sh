#!/bin/bash
set -e
. $TARGET/rootfs/dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

VERSION="2.6.0"
url_grafana="https://grafanarel.s3.amazonaws.com/builds/grafana-$VERSION.linux-x64.tar.gz"

mkdir -p $TARGET/rootfs/
echo ":: Downloading grafana $VERSION"
curl -L $url_grafana  | tar -C $TARGET/rootfs/ -xzf -
mv $TARGET/rootfs/grafana* $TARGET/rootfs/grafana

