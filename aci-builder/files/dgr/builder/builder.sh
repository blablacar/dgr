#!/dgr/bin/busybox sh
set -e
. ${ROOTFS}/dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

export SYSTEMD_LOG_LEVEL=err
rootfs="/opt/stage2/${ACI_NAME}/rootfs"
acihome=/opt/aci-home

echo "ce9d63a98a8b4438882fd795e294cd50" > /etc/machine-id

#TODO copy internals

# build runlevels
mkdir -p /dgr/builder/runlevels
if [ -d /opt/aci-home/runlevels/build ]; then
    cp -Rf /opt/aci-home/runlevels/build /dgr/builder/runlevels
fi
if [ -d /opt/aci-home/runlevels/build-late ]; then
    cp -Rf /opt/aci-home/runlevels/build-late /dgr/builder/runlevels
fi

systemd-nspawn --setenv=LOG_LEVEL=${LOG_LEVEL} --register=no -q --directory=${rootfs} --capability=all \
    --bind=/dgr/builder:/dgr/builder dgr/builder/step-build.sh


# prestart
mkdir -p ${rootfs}/dgr/prestart/{prestart-early,prestart-late}
if [ -d /opt/aci-home/runlevels/prestart-early ]; then
    cp -Rf /opt/aci-home/runlevels/prestart-early/. ${rootfs}/dgr/prestart/prestart-early
fi
if [ -d /opt/aci-home/runlevels/prestart-late ]; then
    cp -Rf /opt/aci-home/runlevels/prestart-late/. ${rootfs}/dgr/prestart/prestart-late
fi
# attributes
mkdir -p ${rootfs}/dgr/attributes/${ACI_NAME}
if [ -d ${acihome}/attributes ]; then
    cp -Rf ${acihome}/attributes/. ${rootfs}/dgr/attributes/${ACI_NAME} # TODO do not copy no .yml .yaml
fi

# files
if [ -d ${acihome}/files ]; then
    cp -Rf ${acihome}/files/. ${rootfs}
fi

# templates
mkdir -p ${rootfs}/dgr/templates
if [ -d ${acihome}/templates ]; then
    cp -Rf ${acihome}/templates/. ${rootfs}/dgr/templates
fi

systemd-nspawn --setenv=LOG_LEVEL=${LOG_LEVEL} --register=no -q --directory=${rootfs} --capability=all \
    --bind=/dgr/builder:/dgr/builder dgr/builder/step-build-late.sh


systemd-nspawn --setenv=LOG_LEVEL=${LOG_LEVEL} --register=no -q --directory=${rootfs} --capability=all \
    --bind=/dgr/builder:/dgr/builder /dgr/bin/busybox sh #TODO remove

sh # TODO remove
