#!/bin/bash
set -e

if [ "$(id -u)" != "0" ]; then
	echo "Sorry, you are not root."
	exit 1
fi

dir=$( dirname $0 )

buildAci() {
    echo -e "\n\n\033[0;32mBuilding aci : ${1}\033[0m\n\n"
    sleep 1
    dgr -W ${1} clean install
}
# base
buildAci ${dir}/aci-base

# none
buildAci ${dir}/none/aci-libc
buildAci ${dir}/none/aci-grafana
buildAci ${dir}/none/aci-prometheus

# debian
buildAci ${dir}/debian/aci-debian
buildAci ${dir}/debian/aci-cassandra

# gentoo
buildAci ${dir}/debian/aci-gentoo-stage4
buildAci ${dir}/debian/aci-gentoo-nginx


echo -e "\n\033[0;32mEverything looks good !\033[0m\n"
