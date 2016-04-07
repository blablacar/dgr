#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

/dgr/bin/prestart #TODO

pacman -Sy yaourt base-devel --noconfirm
echo "yaourt:x:20002:20002:yaourt:/home/yaourt:/usr/bin/sh" >> /etc/passwd
echo "yaourt:x:20002:" >> /etc/group
echo "yaourt:x:20002::::::" >> /etc/shadow

echo "Cmnd_Alias  PACMAN = /usr/bin/pacman, /usr/bin/yaourt
%yaourt ALL=(ALL) NOPASSWD: PACMAN" > /etc/sudoers

#ln -s /proc/self/fd /dev/fd

# su -c 'yaourt -Sy jdk' yaourt

#ln -s /proc/self/fd /dev/fd
#ln -s /proc/self/fd/0 /dev/stdin
#ln -s /proc/self/fd/1 /dev/stdout
#ln -s /proc/self/fd/2 /dev/stderr
