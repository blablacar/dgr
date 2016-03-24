#!/bin/bash
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

echo 'Cleaning rootfs from packages'
apt-get autoremove --purge -y
apt-get clean
echo 'Cleaning TMP dir'
rm -rf /tmp/*


echo 'Cleaning rootfs'
find / -prune -name '*.deb' -exec rm {} \;
for logfile in $(find / -prune -name '*.log')
 do
   > $logfile
 done
for rootdir in /var/lib/apt/lists/ \
               /usr/share/doc \
               /usr/share/man; do
  find ${rootdir} -prune -type f -exec rm {} \;
 done
rm -f  /var/cache/apt/*.bin
rm -Rf /var/lib/apt/lists/*
