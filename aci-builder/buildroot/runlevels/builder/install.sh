#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x
BROOT_DIR=/buildroot

cd ${BROOT_DIR}
cp ${ACI_HOME}/config configs/dgr_defconfig
cp -a ${ACI_HOME}/busybox ./package
cp -a ${ACI_HOME}/tar ./package

cp -a ${ACI_HOME}/busybox/busybox.config ./package/busybox/.config
mkdir -p /dgr-cache/ccache /dgr-cache/dl
make dgr_defconfig
export BR2_CCACHE_DIR=/dgr-cache/ccache
export BR2_DL_DIR=/dgr-cache/dl
make BR2_CCACHE_DIR=/dgr-cache/ccache BR2_DL_DIR=/dgr-cache/dl
cp output/images/rootfs.tar.xz ${ACI_HOME}

ls
exit
