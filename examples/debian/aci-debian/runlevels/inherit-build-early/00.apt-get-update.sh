#!/bin/bash
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

# Jessie is too old, so we add `-o Acquire::Check-Valid-Until=false` (for now)
apt-get -o Acquire::Check-Valid-Until=false update
