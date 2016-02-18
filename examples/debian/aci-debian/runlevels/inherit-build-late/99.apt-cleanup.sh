#!/bin/bash
set -e
source /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

echo 'Cleaning rootfs from packages'
apt-get autoremove --purge -y
apt-get clean
echo 'Cleaning TMP dir'
rm -rf /tmp/*


echo 'Cleaning rootfs'
find / -path '/target*' -prune -name '*.deb' -exec rm {} \;
for logfile in $(find / -path '/target*' -prune -name '*.log')
 do
   > $logfile
 done
for rootdir in /var/lib/apt/lists/ \
               /usr/share/doc \
               /usr/share/man; do
  find ${rootdir} -path '/target*' -prune -type f -exec rm {} \;
 done
rm -f  /var/cache/apt/*.bin
