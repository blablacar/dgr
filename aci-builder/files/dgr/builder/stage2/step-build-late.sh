#!/dgr/bin/busybox sh
set -e
. /dgr/builder/export
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

onError() {
    if [ "${TRAP_ON_ERROR}" == "true" ]; then
        echo_red "${1} failed. dropping to shell in build-late"
        sh
    fi
    exit 1
}

execute_files "/dgr/builder/runlevels/build-late" || onError "Build-late"
execute_files "/dgr/runlevels/inherit-build-late" || onError "Inherit-build-late"

if [ "${TRAP_ON_STEP}" == "true" ]; then
    echo_purple "Trap requested dropping to shell in build-late"
    sh
fi
