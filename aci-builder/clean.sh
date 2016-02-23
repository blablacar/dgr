#!/bin/bash
set -x
set -e
dir=$( dirname $0 )
target=${dir}/../dist/aci-builder
rm -Rf ${target}
rm -f ${dir}/../dist/bindata/aci-builder.aci
