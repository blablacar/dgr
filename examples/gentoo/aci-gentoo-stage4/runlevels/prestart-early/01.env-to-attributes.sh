#!/bin/bash
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

mkdir -p /dgr/attributes/aci-gentoo-stage4
cat > /dgr/attributes/aci-gentoo-stage4/prestart.yml <<EOF
default:
  rootfs: ${ROOTFS}
EOF
