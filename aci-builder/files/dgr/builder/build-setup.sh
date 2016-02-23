#!/bin/sh
set -e
dir=$( dirname "$0" )
. "${dir}/../bin/functions.sh"
isLevelEnabled "debug" && set -x

execute_files "${BASEDIR}/runlevels/build-setup"
