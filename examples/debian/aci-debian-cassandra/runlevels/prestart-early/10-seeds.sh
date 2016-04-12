#!/bin/bash
set -x
set -e

cat > /dgr/attributes/aci-cassandra/prestart.yml <<EOF
default:
  cassandra:
    seeds: $IP
EOF
