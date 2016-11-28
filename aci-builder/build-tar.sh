#!/bin/bash

set -euo pipefail
if [[ ! -z ${debug+x} ]]; then
  set -x
fi

start=$(date +%s)
dir="$(dirname $0)"
mkdir -p "${dir}/files/dgr/usr/bin"
: ${tar:="$(realpath "${dir}/files/dgr/usr/bin/tar")"}

echo -e "\033[0;32mBuilding tar\033[0m\n"

# 'core2' is very old, but without setting an option some GCC versions
# will pick built-in defaults – which could be as "recent" as 'silvermont',
# thus excluding CPUs without SSE4 (like some pre-2007 AMDs still in the wild).
: ${CPU_SETTING:="-march=core2 -mtune=intel"}
if [[ "$(uname -m)" != "x86_64" ]]; then
  CPU_SETTING=""
fi

WORKDIR="$(mktemp -d -t aci-builder-tar.XXXXXX)"
pushd . &>/dev/null
cd ${WORKDIR}
curl --silent --show-error --fail --location \
  -H "accept: application/x-xz, application/x-tar, application/tar+xz" \
  https://ftp.gnu.org/gnu/tar/tar-1.29.tar.xz \
| tar --strip-components=1 --xz -x

if ! ./configure --prefix=/usr --libexecdir=/libexec --disable-rpath \
  CFLAGS="-Os ${CPU_SETTING} -ffunction-sections -fdata-sections -fstack-protector-strong -fpie -fpic" \
  LDFLAGS="-Wl,-O1 -Wl,-z,relro,-z,now -Wl,--as-needed -Wl,--strip-all -Wl,--gc-sections" >/dev/null; then

  # Old compiler. Most probably in a Travis-CI VM. Resort to conservative settings.
  echo "== First run of configure failed. Trying fallback options:"
  ./configure --prefix=/usr --libexecdir=/libexec --disable-rpath \
    CFLAGS="-Os ${CPU_SETTING%% *} -ffunction-sections -fdata-sections" \
    LDFLAGS="-Wl,-O1 -Wl,-z,relro,-z,now -Wl,--strip-all -Wl,--gc-sections" >/dev/null
fi
make -j$(nproc) >/dev/null

popd &>/dev/null
mv ${WORKDIR}/src/tar "${dir}/files/dgr/usr/bin/tar"
rm -r ${WORKDIR}

echo -e "\033[0;32mBuilding tar took: $[ $(date +%s) - ${start} ]s\033[0m\n"

exit 0
