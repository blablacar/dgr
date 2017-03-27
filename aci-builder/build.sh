#!/bin/bash
set -e
dir=$(dirname $0)
dir_absolute=$(dirname $(readlink -nf $0))
dist=${dir}/../dist
target=${dist}/aci-builder
rootfs=${target}/rootfs
echo -e "\033[0;32mBuilding aci-builder\033[0m\n"
rootfs_hash=7454bbe622619b85668da082ad961f7cd3c49d70
${dir}/clean.sh

mkdir -p ${rootfs}/dgr ${rootfs}/usr/bin
[ -f ${dist}/templater ] || (echo "build templater first" && exit 1)

GOOS=linux GOARCH=amd64 go build --ldflags '-s -w -extldflags "-static"' -o ${rootfs}/dgr/builder/stage1/run ${dir}/bin-run
upx ${rootfs}/dgr/builder/stage1/run
set +e
cat ${dir_absolute}/rootfs.tar.xz | openssl dgst -sha1 | grep ${rootfs_hash}
retcode=$?
set -e
if [ ${retcode} -eq 1 ];then
  sudo dgr clean install -W ${dir_absolute}/../examples/aci-base
  sudo dgr clean install -W ${dir_absolute}/../examples/debian/aci-debian
  sudo dgr clean install -W ${dir_absolute}/../examples/aci-buildroot
  sudo dgr build -W ${dir}/buildroot
  cp ${dir_absolute}/buildroot/rootfs.tar.xz ${dir_absolute}/rootfs.tar.xz
  rm -f ${dir_absolute}/buildroot/rootfs.tar.xz
  sudo dgr clean -W ${dir}/buildroot
fi

sudo tar xf ${dir_absolute}/rootfs.tar.xz -C ${dir_absolute}/../dist --strip 3 ./usr/bin/tar
sudo ${dir_absolute}/../dist/tar --transform "s:usr/sbin/haveged:usr/bin/haveged:" \
  --transform "s:usr/lib/systemd/:usr/lib/:" \
  --exclude "./etc/udev" \
  --exclude "./usr/share/locale" \
  --exclude "./usr/libexec" \
  --exclude "./usr/lib/udev" \
  --exclude "./usr/sbin" \
  -C ${rootfs}/dgr/ \
  -xf ${dir}/rootfs.tar.xz

sudo cp -R ${dir}/files/. ${rootfs}
sudo chown root: ${rootfs}
cp ${dir_absolute}/manifest.json ${target}/manifest
sudo cp --no-preserve=ownership ${dist}/templater ${rootfs}/dgr/usr/bin/

# sudo bash -c "cd ${rootfs}/dgr/usr && ln -s bin sbin && cd -"
sudo ln -s ./sbin ${rootfs}/dgr/usr/bin
cd ${target}
sudo ${dir_absolute}/../dist/tar --sort=name --numeric-owner \
      -cpzf ../bindata/aci-builder.aci rootfs manifest
sudo chown ${USER}: ../bindata/aci-builder.aci
sudo rm -Rf rootfs/
cd -
