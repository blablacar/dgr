#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

export SYSTEMD_LOG_LEVEL=err
export ROOTFS="/opt/stage2/${ACI_NAME}/rootfs"
    if [ -z ${ACI_HOME} ]; then
    echo_red "'ACI_HOME' is not set in build aci. This is a builder issue"
    exit 1
fi

linkToDgrIfEmpty() {
    if [ ! "$(ls -A ${1}/ 2> /dev/null)" ]; then
        rm -Rf ${1}
        ln -s ${2} ${1}
    fi
}

export PATH=$PATH:/dgr/bin

linkToDgrIfEmpty /bin /usr/bin
linkToDgrIfEmpty /usr/bin /dgr/usr/bin
linkToDgrIfEmpty /lib64 /dgr/usr/lib
linkToDgrIfEmpty /etc/ssl /dgr/etc/ssl

echo "ce9d63a98a8b4438882fd795e294cd50" > /etc/machine-id

# builder
execute_files "${ACI_HOME}/runlevels/builder"

# copy internals
mkdir -p ${ROOTFS}/dgr/bin
cmp -s /dgr/bin/busybox ${ROOTFS}/dgr/bin/busybox || cp /dgr/bin/busybox ${ROOTFS}/dgr/bin/busybox
cmp -s /dgr/bin/functions.sh ${ROOTFS}/dgr/bin/functions.sh || cp /dgr/bin/functions.sh ${ROOTFS}/dgr/bin/functions.sh
cmp -s /dgr/bin/prestart ${ROOTFS}/dgr/bin/prestart || cp /dgr/bin/prestart ${ROOTFS}/dgr/bin/prestart
cmp -s /dgr/bin/templater ${ROOTFS}/dgr/bin/templater || cp /dgr/bin/templater ${ROOTFS}/dgr/bin/templater

mkdir -p ${ROOTFS}/usr/bin # this is required by the systemd-nspawn

if [ -d ${ACI_HOME}/runlevels/build ] || [ -d ${ACI_HOME}/runlevels/build-late ]; then
    # build runlevels
    mkdir -p /dgr/builder/runlevels
    if [ -d ${ACI_HOME}/runlevels/build ]; then
        cp -Rf ${ACI_HOME}/runlevels/build /dgr/builder/runlevels
    fi
    if [ -d ${ACI_HOME}/runlevels/build-late ]; then
        cp -Rf ${ACI_HOME}/runlevels/build-late /dgr/builder/runlevels
    fi

    # inherit
    if [ "$(${ACI_HOME}/runlevels/inherit-build-late 2> /dev/null)" ]; then
        mkdir -p ${ROOTFS}/dgr/runlevels/inherit-build-late
        cp -Rf ${ACI_HOME}/runlevels/inherit-build-late/. ${ROOTFS}/dgr/runlevels/inherit-build-late
    fi
    if [ "$(${ACI_HOME}/runlevels/inherit-build-early 2> /dev/null)" ]; then
        mkdir -p ${ROOTFS}/dgr/runlevels/inherit-build-early
        cp -Rf ${ACI_HOME}/runlevels/inherit-build-early/. ${ROOTFS}/dgr/runlevels/inherit-build-early
    fi

    LD_LIBRARY_PATH=/dgr/usr/lib /dgr/usr/lib/ld-linux-x86-64.so.2 /dgr/usr/bin/systemd-nspawn --setenv=LOG_LEVEL=${LOG_LEVEL} --register=no -q --directory=${ROOTFS} --capability=all \
        --bind=/dgr/builder:/dgr/builder dgr/builder/stage2/step-build.sh
fi

# prestart
if [ "$(ls -A ${ACI_HOME}/runlevels/prestart-early 2> /dev/null)" ]; then
    mkdir -p ${ROOTFS}/dgr/runlevels/prestart-early
    cp -Rf ${ACI_HOME}/runlevels/prestart-early/. ${ROOTFS}/dgr/runlevels/prestart-early
fi
if [ "$(ls -A ${ACI_HOME}/runlevels/prestart-late 2> /dev/null)" ]; then
    mkdir -p ${ROOTFS}/dgr/runlevels/prestart-late
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


if [ -d ${ACI_HOME}/runlevels/build ] || [ -d ${ACI_HOME}/runlevels/build-late ]; then
    # build-late
    LD_LIBRARY_PATH=/dgr/usr/lib /dgr/usr/lib/ld-linux-x86-64.so.2 /dgr/usr/bin/systemd-nspawn --setenv=LOG_LEVEL=${LOG_LEVEL} --register=no -q --directory=${ROOTFS} --capability=all \
        --bind=/dgr/builder:/dgr/builder dgr/builder/stage2/step-build-late.sh
fi

# builder-late
#execute_files "${ACI_HOME}/runlevels/builder-late"

rm -Rf /dgr/builder