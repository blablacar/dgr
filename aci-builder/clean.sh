#!/bin/bash
set -x
set -e
dir=$( dirname $0 )
target=${dir}/../dist/aci-builder

sudo rm -Rf ${target}
sudo rm -f ${dir}/../dist/bindata/aci-builder.aci
