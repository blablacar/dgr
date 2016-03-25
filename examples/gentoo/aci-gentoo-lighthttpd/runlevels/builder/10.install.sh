#!/bin/bash
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

emerge -v glibc lighttpd

#TODO inherit cleanup of $ROOTFS/usr/share/{doc,man}
