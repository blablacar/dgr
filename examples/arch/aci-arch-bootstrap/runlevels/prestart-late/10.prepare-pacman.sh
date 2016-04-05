#!/bin/bash

pacman-key --populate archlinux
mkdir -p ${ROOTFS}/var/lib/pacman
pacman -Sy
