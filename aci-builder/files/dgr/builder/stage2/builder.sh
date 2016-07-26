#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

onError() {
    if [ "${CATCH_ON_ERROR}" == "true" ]; then
        echo_red "${1} failed. dropping to shell in builder"
        sh
    fi
    exit 1
}

export SYSTEMD_LOG_LEVEL=err
export ROOTFS="/opt/stage2/${ACI_NAME}/rootfs"
chmod 755 /opt/stage2 && chmod 755 /opt/stage2/${ACI_NAME} # this is required as soon as you run builder action as non root

if [ -z ${ACI_HOME} ]; then
    echo_red "'ACI_HOME' is not set in build aci. This is a builder issue"
    exit 1
fi

if [ ! "$(ls -A /usr/bin/ 2> /dev/null)" ]; then
    isLevelEnabled "debug" && echo_purple "Nothing as builder dependency, mapping / to /dgr"
    cp /etc/resolv.conf /dgr/etc/resolv.conf
    rm -Rf /usr /etc /lib64 /lib /bin
    ln -s /usr/bin /bin
    ln -s /dgr/usr /usr
    ln -s /dgr/etc /etc
    ln -s /dgr/lib64 /lib64
    ln -s /dgr/lib /lib
fi

echo "ce9d63a98a8b4438882fd795e294cd50" > /etc/machine-id

mkdir -p /dgr/builder/runlevels
mkdir -p /dgr/builder/attributes

# allow adding attributes and runlevels in prestart without being root
mkdir -p ${ROOTFS}/dgr/runlevels
mkdir -p ${ROOTFS}/dgr/attributes
chmod 777 ${ROOTFS}/dgr/runlevels
chmod 777 ${ROOTFS}/dgr/attributes

# save envs
export | grep -v -E " SHLV=| ROOTFS=| TARGET= | ACI_PATH= | ACI_HOME= | ACI_EXEC=" > /dgr/builder/export

# inherit build
execute_files "/dgr/runlevels/inherit-build-early" || onError "Inherit-build-early"

# builder runlevel
execute_files "${ACI_HOME}/runlevels/builder" || onError "Builder"
if [ "$(ls -A "${ACI_HOME}/runlevels/builder" 2> /dev/null)" ] && [ "${CATCH_ON_STEP}" == "true" ]; then
    echo_purple "Catch requested dropping to shell after builder"
    sh
fi

# copy internals
mkdir -p ${ROOTFS}/dgr/bin
cmp -s /dgr/bin/busybox ${ROOTFS}/dgr/bin/busybox || cp /dgr/bin/busybox ${ROOTFS}/dgr/bin/busybox
cmp -s /dgr/bin/functions.sh ${ROOTFS}/dgr/bin/functions.sh || cp /dgr/bin/functions.sh ${ROOTFS}/dgr/bin/functions.sh
cmp -s /dgr/bin/prestart ${ROOTFS}/dgr/bin/prestart || cp /dgr/bin/prestart ${ROOTFS}/dgr/bin/prestart
cmp -s /dgr/bin/templater ${ROOTFS}/dgr/bin/templater || cp /dgr/bin/templater ${ROOTFS}/dgr/bin/templater

mkdir -p ${ROOTFS}/usr/bin # this is required by the systemd-nspawn

# inherit
if [ -d ${ACI_HOME}/runlevels/inherit-build-early ]; then
    mkdir -p ${ROOTFS}/dgr/runlevels/inherit-build-early
    chmod 777 ${ROOTFS}/dgr/runlevels/inherit-build-early
    cp -Rf ${ACI_HOME}/runlevels/inherit-build-early/. ${ROOTFS}/dgr/runlevels/inherit-build-early
fi
if [ -d ${ACI_HOME}/runlevels/inherit-build-late ]; then
    mkdir -p ${ROOTFS}/dgr/runlevels/inherit-build-late
    chmod 777 ${ROOTFS}/dgr/runlevels/inherit-build-late
    cp -Rf ${ACI_HOME}/runlevels/inherit-build-late/. ${ROOTFS}/dgr/runlevels/inherit-build-late
fi

if [ -d ${ACI_HOME}/runlevels/build ]; then
    cp -Rf ${ACI_HOME}/runlevels/build /dgr/builder/runlevels
fi
if [ -d ${ACI_HOME}/runlevels/build-late ]; then
    cp -Rf ${ACI_HOME}/runlevels/build-late /dgr/builder/runlevels
fi

# build runlevel
if [ -d ${ACI_HOME}/runlevels/build ] || [ -d ${ACI_HOME}/runlevels/build-late ] || [ -d ${ROOTFS}/dgr/runlevels/inherit-build-early ]; then

    LD_LIBRARY_PATH=/dgr/usr/lib /dgr/usr/lib/ld-linux-x86-64.so.2 /dgr/usr/bin/systemd-nspawn \
        --register=no -q --directory=${ROOTFS} --capability=all \
        --bind=/dgr/builder:/dgr/builder dgr/builder/stage2/step-build.sh || onError "Build"
fi

# prestart
if [ "$(ls -A ${ACI_HOME}/runlevels/prestart-early 2> /dev/null)" ]; then
    mkdir -p ${ROOTFS}/dgr/runlevels/prestart-early
    chmod 777 ${ROOTFS}/dgr/runlevels/prestart-early
    cp -Rf ${ACI_HOME}/runlevels/prestart-early/. ${ROOTFS}/dgr/runlevels/prestart-early
fi
if [ "$(ls -A ${ACI_HOME}/runlevels/prestart-late 2> /dev/null)" ]; then
    mkdir -p ${ROOTFS}/dgr/runlevels/prestart-late
    chmod 777 ${ROOTFS}/dgr/runlevels/prestart-late
    cp -Rf ${ACI_HOME}/runlevels/prestart-late/. ${ROOTFS}/dgr/runlevels/prestart-late
fi

# attributes
if [ "$(ls -A ${ACI_HOME}/attributes 2> /dev/null)" ]; then
    mkdir -p ${ROOTFS}/dgr/attributes/${ACI_NAME}
    find ${ACI_HOME}/attributes \( -name "*.yml" -o -name "*.yaml" \) -exec cp {} ${ROOTFS}/dgr/attributes/${ACI_NAME} \;
fi

# files
if [ -d ${ACI_HOME}/files ]; then
    cp -Rf ${ACI_HOME}/files/. ${ROOTFS}
fi

# templates
if [ "$(ls -A ${ACI_HOME}/templates 2> /dev/null)"  ]; then
    mkdir -p ${ROOTFS}/dgr/templates
    cp -Rf ${ACI_HOME}/templates/. ${ROOTFS}/dgr/templates
fi

# build-late runlevel
if [ -d ${ACI_HOME}/runlevels/build ] || [ -d ${ACI_HOME}/runlevels/build-late ] || [ -d ${ROOTFS}/dgr/runlevels/inherit-build-late ]; then
    LD_LIBRARY_PATH=/dgr/usr/lib /dgr/usr/lib/ld-linux-x86-64.so.2 /dgr/usr/bin/systemd-nspawn \
        --register=no -q --directory=${ROOTFS} --capability=all \
        --bind=/dgr/builder:/dgr/builder dgr/builder/stage2/step-build-late.sh || onError "Build-late"
fi

# inherit build
execute_files "/dgr/runlevels/inherit-build-late" || onError "Inherit-build-late"

rmdir ${ROOTFS}/dgr/builder >/dev/null 2>&1 || true

if [ "${CATCH_ON_STEP}" == "true" ]; then
    echo_purple "Catch requested dropping to shell at end of build"
    sh
fi
