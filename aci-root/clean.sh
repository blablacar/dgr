#!/bin/bash
if [ "$(id -u)" != "0" ]; then
	echo "Sorry, you are not root."
	exit 1
fi

set -x
set -e
dir=$( dirname $0 )
target=${dir}/../dist/aci-root
rm -Rf ${target}
rm -f ${dir}/../dist/bindata/aci-root.aci
