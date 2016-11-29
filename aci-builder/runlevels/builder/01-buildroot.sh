#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

if [ -f /cache/rootfs.tar.gz ]; then
    echo_purple "Not rebuilding buildroot, rootfs.tar.gz found in cache folder"
    exit 0
fi

version=${ACI_VERSION%-*}
url="https://buildroot.org/downloads/buildroot-${version}.tar.gz"

echo_green "Download buildroot"
curl -Ls ${url} -o /buildroot.tar.gz

echo_green "Extract buildroot"
mkdir /buildroot && cd /buildroot
tar --strip 1 -xzf /buildroot.tar.gz

echo_green "Building buildroot"
cd /buildroot
sh

