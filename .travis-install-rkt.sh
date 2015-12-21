#!/bin/bash
set -x
set -e

version="v0.14.0"
filename="rkt-${version}.tar.gz"
url="https://github.com/coreos/rkt/releases/download/${version}/${filename}"

$(rkt version | grep "${version}") || {
    mkdir -p "/tmp/rkt"
    cd "/tmp/rkt"
	wget $url
	tar xvzf "${filename}" --strip=1
	cp rkt stage1* /bin/
	cd -
}
