#!/dgr/bin/busybox sh

mkdir /go
export GOPATH=/go
export PATH=$GOPATH/bin:$PATH

cd /go/src/github.com/blablacar/dgr/
./build.sh

mkdir -p ${ROOTFS}/usr/bin
cp dist/linux-amd64/dgr ${ROOTFS}/usr/bin
