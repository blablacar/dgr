#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x


apt-get install -qqy build-essential bison \
                                     flex \
                                     gettext \
                                     curl \
                                     wget \
                                     sed \
                                     make \
                                     binutils \
                                     build-essential \
                                     gcc \
                                     g++ \
                                     bash \
                                     patch \
                                     gzip \
                                     bzip2 \
                                     perl \
                                     tar \
                                     cpio \
                                     python \
                                     unzip \
                                     rsync \
                                     file \
                                     bc \
                                     cvs \
                                     git \
                                     mercurial \
                                     rsync \
                                     subversion
BROOT_VERSION=${ACI_VERSION%-*}
BROOT_URL=https://buildroot.org/downloads/buildroot-${BROOT_VERSION}.tar.gz
BROOT_FILE=${BROOT_URL##*/}
BROOT_DIR=${BROOT_FILE//.tar.gz/}
curl ${BROOT_URL} | tar -C / -xzv
mv /${BROOT_DIR} /buildroot
