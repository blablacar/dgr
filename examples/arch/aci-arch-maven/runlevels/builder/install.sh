#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x


su -c 'yaourt -S jdk --noconfirm' yaourt

pacman -S maven which --noconfirm
