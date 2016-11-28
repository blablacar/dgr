#!/bin/bash
set -x
set -e
dir=$( dirname $0 )

[ $(command -v glide) ] || go get github.com/Masterminds/glide
cd ${dir}/..
glide install --strip-vendor --skip-test