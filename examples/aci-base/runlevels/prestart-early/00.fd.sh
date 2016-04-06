#!/dgr/bin/busybox sh
set -e
source /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

[ -d /dev/fd ] || ln -s /proc/self/fd /dev/fd
[ -L /dev/stdin ] || ln -s /proc/self/fd/0 /dev/stdin
[ -L /dev/stdout ] || ln -s /proc/self/fd/1 /dev/stdout
[ -L /dev/stderr ] || ln -s /proc/self/fd/2 /dev/stderr

