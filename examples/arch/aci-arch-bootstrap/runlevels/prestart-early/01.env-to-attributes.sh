#!/bin/bash
set -x
set -e

mkdir -p /dgr/attributes/aci-arch-bootstrap
cat > /dgr/attributes/aci-arch-bootstrap/prestart.yml <<EOF
default:
  rootfs: ${ROOTFS}
EOF
