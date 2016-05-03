#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

cat > /dgr/attributes/aci-debian-cassandra/prestart.yml <<EOF
default:
  cassandra:
    seeds: $IP
EOF
