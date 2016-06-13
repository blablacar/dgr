#!/dgr/bin/busybox sh
package_name=${ACI_NAME#aci-centos-*}
[ x${package_name} == x"aci-centos" ] && exit 0
[ ! -d /dgr/builder/runlevels/build ] && [ ! -d /dgr/builder/runlevels/build-late ] && exit 0

mkdir -p /dgr/attributes/aci-centos

cat > /dgr/attributes/aci-centos/prestart.yml <<EOF
default:
  rootfs: ${ROOTFS}
EOF
echo $ROOTFS
