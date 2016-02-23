#!/dgr/bin/busybox sh
set -e
export TARGET=$( dirname $0 )
export ROOTFS=
export TERM=xterm

. ${ROOTFS}/dgr/bin/functions.sh
isLevelEnabled "debug" && set -x


execute_files "$ROOTFS/dgr/runlevels/inherit-build-early"
execute_files "$ROOTFS/dgr/builder/runlevels/build"
