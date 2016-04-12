#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

cat > /dgr/attributes/aci-cassandra/prestart.yml <<EOF
default:
  cassandra:
    seeds: $IP
EOF
