#!/bin/bash
set -x
set -e

cat > /cnt/attributes/aci-cassandra/prestart.yml <<EOF
default:
  cassandra:
    seeds: $IP
EOF
