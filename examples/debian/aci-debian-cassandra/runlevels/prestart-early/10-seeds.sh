#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

mkdir -p /dgr/attributes/aci-debian-cassandra-prestart
cat > /dgr/attributes/aci-debian-cassandra-prestart/prestart.yml <<EOF
default:
  cassandra:
    seeds: $IP
EOF
