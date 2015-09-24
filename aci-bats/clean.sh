#!/bin/bash
set -x
set -e

DIR=$( dirname $0 )

rm -Rf $DIR/rootfs/ $DIR/aci-bats.aci
