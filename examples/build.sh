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
    dgr -W ${1} -L debug clean install
}

# base
buildAci ${dir}/aci-base

# none
buildAci ${dir}/none/aci-libc
buildAci ${dir}/none/aci-grafana
buildAci ${dir}/none/aci-prometheus
buildAci ${dir}/none/aci-rkt

# debian
buildAci ${dir}/debian/aci-debian
buildAci ${dir}/debian/aci-debian-cassandra

# alpine
buildAci ${dir}/alpine/aci-alpine-base
buildAci ${dir}/alpine/aci-alpine-nginx

# gentoo
buildAci ${dir}/gentoo/aci-gentoo-stage4
#buildAci ${dir}/gentoo/aci-gentoo-lighttpd # TODO build is too long for travis

# archlinux
buildAci ${dir}/arch/aci-arch-bootstrap
buildAci ${dir}/arch/aci-arch-pacman
buildAci ${dir}/arch/aci-arch-yaourt
buildAci ${dir}/arch/aci-arch-git
buildAci ${dir}/arch/aci-arch-go
buildAci ${dir}/arch/aci-arch-upx
buildAci ${dir}/arch/aci-arch-gcc
buildAci ${dir}/arch/aci-arch-jdk
buildAci ${dir}/arch/aci-arch-maven
buildAci ${dir}/arch/aci-arch-prometheus-jmx-exporter

# pod
buildAci ${dir}/pod/pod-cassandra

echo -e "\n\033[0;32mEverything looks good !\033[0m\n"
