#!/bin/bash
set -e
source /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

apt-get update
