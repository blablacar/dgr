#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

mkdir -p /dgr/attributes/aci-arch-bootstrap
cat > /dgr/attributes/aci-arch-bootstrap/prestart.yml <<EOF
default:
  rootfs: ${ROOTFS}
EOF
