#!/bin/bash
set -x
set -e
dir=$( dirname $0 )

[ $(command -v glide-vc) ] || go get github.com/sgotti/glide-vc
cd ${dir}/..
glide-vc --only-code --no-tests --keep="**/*.json.in" --use-lock-file
