#!/dgr/bin/busybox sh
set -e
export TARGET=$( dirname $0 )
export ROOTFS=
export TERM=xterm

. ${ROOTFS}/dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

onError() {
    if [ "${TRAP_ON_ERROR}" == "true" ]; then
        echo_red "${1} failed. dropping to shell in build"
        sh
    fi
    exit 1
}

execute_files "$ROOTFS/dgr/runlevels/inherit-build-early" || onError "Inherit-build-early"
execute_files "$ROOTFS/dgr/builder/runlevels/build" || onError "Build"
