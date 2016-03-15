#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

export SYSTEMD_LOG_LEVEL=err
rootfs="/opt/stage2/${ACI_NAME}/rootfs"
if [ -z ${aci_home} ]; then
    echo_red "'aci_home' is not set in build aci. This is a builder issue"
    exit 1
fi

echo "ce9d63a98a8b4438882fd795e294cd50" > /etc/machine-id

# builder-early
execute_files "${aci_home}/runlevels/builder-early"

# copy internals
mkdir -p ${rootfs}/dgr/b    in
cmp -s /dgr/bin/busybox ${rootfs}/dgr/bin/busybox || cp /dgr/bin/busybox ${rootfs}/dgr/bin/busybox
cmp -s /dgr/bin/functions.sh ${rootfs}/dgr/bin/functions.sh || cp /dgr/bin/functions.sh ${rootfs}/dgr/bin/functions.sh
cmp -s /dgr/bin/prestart ${rootfs}/dgr/bin/prestart || cp /dgr/bin/prestart ${rootfs}/dgr/bin/prestart
cmp -s /dgr/bin/templater ${rootfs}/dgr/bin/templater || cp /dgr/bin/templater ${rootfs}/dgr/bin/templater

# build runlevels
mkdir -p /dgr/builder/runlevels
if [ -d ${aci_home}/runlevels/build ]; then
    cp -Rf ${aci_home}/runlevels/build /dgr/builder/runlevels
fi
if [ -d ${aci_home}/runlevels/build-late ]; then
    cp -Rf ${aci_home}/runlevels/build-late /dgr/builder/runlevels
fi

# inherit
if [ "$(${aci_home}/runlevels/inherit-build-late 2> /dev/null)" ]; then
    mkdir -p ${rootfs}/dgr/runlevels/inherit-build-late
    cp -Rf ${aci_home}/runlevels/inherit-build-late/. ${rootfs}/dgr/runlevels/inherit-build-late
fi
if [ "$(${aci_home}/runlevels/inherit-build-early 2> /dev/null)" ]; then
    mkdir -p ${rootfs}/dgr/runlevels/inherit-build-early
    cp -Rf ${aci_home}/runlevels/inherit-build-early/. ${rootfs}/dgr/runlevels/inherit-build-early
fi

mkdir -p ${rootfs}/usr/bin # this is required by the systemd-nspawn

systemd-nspawn --setenv=LOG_LEVEL=${LOG_LEVEL} --register=no -q --directory=${rootfs} --capability=all \
    --bind=/dgr/builder:/dgr/builder dgr/builder/stage2/step-build.sh

# prestart
if [ "$(ls -A ${aci_home}/runlevels/prestart-early 2> /dev/null)" ]; then
    mkdir -p ${rootfs}/dgr/runlevels/prestart-early
    cp -Rf ${aci_home}/runlevels/prestart-early/. ${rootfs}/dgr/runlevels/prestart-early
fi
if [ "$(ls -A ${aci_home}/runlevels/prestart-late 2> /dev/null)" ]; then
    mkdir -p ${rootfs}/dgr/runlevels/prestart-late
    cp -Rf ${aci_home}/runlevels/prestart-late/. ${rootfs}/dgr/runlevels/prestart-late
fi

# attributes
if [ "$(ls -A ${aci_home}/attributes 2> /dev/null)" ]; then
    mkdir -p ${rootfs}/dgr/attributes/${ACI_NAME}
    find ${aci_home}/attributes \( -name "*.yml" -o -name "*.yaml" \) -exec cp {} ${rootfs}/dgr/attributes/${ACI_NAME} \;
fi

# files
if [ -d ${aci_home}/files ]; then
    cp -Rf ${aci_home}/files/. ${rootfs}
fi

# templates
if [ "$(ls -A ${aci_home}/templates 2> /dev/null)"  ]; then
    mkdir -p ${rootfs}/dgr/templates
    cp -Rf ${aci_home}/templates/. ${rootfs}/dgr/templates
fi

systemd-nspawn --setenv=LOG_LEVEL=${LOG_LEVEL} --register=no -q --directory=${rootfs} --capability=all \
    --bind=/dgr/builder:/dgr/builder dgr/builder/stage2/step-build-late.sh


# builder-late
execute_files "${aci_home}/runlevels/builder-late"


rm -Rf /dgr/builder