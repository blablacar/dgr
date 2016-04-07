#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

dir=$(dirname $0)

#pacman -S --noconfirm `$dir/utils/get-pacman-dependencies.sh`
pacman -S --noconfirm \
  acl attr bzip2 curl e2fsprogs expat glibc gpgme keyutils krb5 libarchive libassuan \
  libgpg-error libidn libssh2 lzo openssl pacman xz zlib haveged \
  sed awk grep util-linux sudo \
  --assume-installed perl --assume-installed pcre
