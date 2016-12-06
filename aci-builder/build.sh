#!/bin/bash
set -x
set -e
dir=$( dirname $0 )
dist=${dir}/../dist
target=${dist}/aci-builder
rootfs=${target}/rootfs

echo -e "\033[0;32mBuilding aci-builder\033[0m\n"

${dir}/clean.sh

mkdir -p ${rootfs}/dgr ${rootfs}/usr/bin
[ -f ${dist}/templater ] || (echo "build templater first" && exit 1)

GOOS=linux GOARCH=amd64 go build --ldflags '-s -w -extldflags "-static"' -o ${rootfs}/dgr/builder/stage1/run ${dir}/bin-run
upx ${rootfs}/dgr/builder/stage1/run

: ${tar:="$(realpath "${dir}/files/dgr/usr/bin/tar")"}
if [[ ! -x "${tar}" ]] || ! "${tar}" --help | grep -q -F sort; then
  # 'core2' is very old, but without setting an option some GCC versions
  # will pick built-in defaults – which could be as "recent" as 'silvermont',
  # thus excluding CPUs without SSE4 (like some pre-2007 AMDs still in the wild).
  : ${CPU_SETTING:="-march=core2 -mtune=intel"}
  if [[ "$(uname -m)" != "x86_64" ]]; then
    CPU_SETTING=""
  fi

  WORKDIR="$(mktemp -d -t aci-builder-tar.XXXXXX)"
  pushd .
  cd ${WORKDIR}
  curl -fLROsS http://ftp.gnu.org/gnu/tar/tar-1.29.tar.xz
  tar --strip-components=1 -xaf tar-1.29.tar.xz

  ./configure --prefix=/usr --libexecdir=/libexec --disable-rpath \
    CFLAGS="-Os ${CPU_SETTING} -ffunction-sections -fdata-sections -fstack-protector-strong -fpie -fpic" \
    LDFLAGS="-Wl,-O1 -Wl,-z,relro -Wl,-znow -Wl,--as-needed -Wl,--strip-all -Wl,--gc-sections" >/dev/null
  make -j$(nproc) >/dev/null

  popd
  mkdir -p ${dir}/files/dgr/usr/bin
  mv ${WORKDIR}/src/tar "${tar}"
  rm -r ${WORKDIR}
fi

sudo "${tar}" \
  --transform "s:usr/sbin/haveged:usr/bin/haveged:" \
  --exclude "./etc/udev" \
  --exclude "./usr/share/locale" \
  --exclude "./usr/libexec" \
  --exclude "./usr/lib/systemd" \
  --exclude "./usr/lib/udev" \
  --exclude "./usr/sbin" \
  -C ${rootfs}/dgr/ \
  -xf ${dir}/rootfs.tar.xz
sudo cp -R ${dir}/files/. ${rootfs}
sudo chown root: ${rootfs}
cp ${dir}/manifest.json ${target}/manifest
sudo cp --no-preserve=ownership ${dist}/templater ${rootfs}/dgr/usr/bin/

sudo bash -c "cd ${rootfs}/dgr/usr && ln -s bin sbin && cd -"

cd ${target}
sudo "${tar}" --sort=name --numeric-owner \
  -cpzf ../bindata/aci-builder.aci manifest rootfs
sudo chown ${USER}: ../bindata/aci-builder.aci
sudo rm -Rf rootfs/
cd -
