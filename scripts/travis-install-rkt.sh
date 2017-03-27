#!/bin/bash
set -x
set -e

version="v1.25.0"
filename="rkt-${version}.tar.gz"
url="https://github.com/coreos/rkt/releases/download/${version}/${filename}"

$(rkt version | grep "${version}") || {
    mkdir -p "/tmp/rkt"
    cd "/tmp/rkt"
	wget $url
	tar xvzf "${filename}"
	cp rkt*/rkt rkt*/stage1* /bin/
	chmod +x /bin/rkt
	mkdir -p /var/lib/rkt
	cd -
}

rkt version
