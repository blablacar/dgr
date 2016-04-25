#!/dgr/bin/busybox sh

mkdir -p /dgr/attributes/aci-centos

cat > /dgr/attributes/aci-centos/prestart.yml <<EOF
default:
  rootfs: ${ROOTFS}
EOF
echo $ROOTFS
