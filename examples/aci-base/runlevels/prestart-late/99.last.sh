#!/dgr/bin/busybox sh
set -e
source /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x


isLevelEnabled "debug" && {
    echo ""
    echo "Pre-victory !"
    echo ""
}
