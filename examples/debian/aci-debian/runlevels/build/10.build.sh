#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

export LANG=C
export TERM=xterm
export DEBIAN_FRONTEND=noninteractive

echo 'APT::Install-Recommends "0";' > /etc/apt/apt.conf.d/01norecommend
echo 'APT::Install-Suggests "0";' >> /etc/apt/apt.conf.d/01norecommend

apt-get update
apt-get install -y ca-certificates

update-ca-certificates
