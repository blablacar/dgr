#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

cat > /etc/apt/sources.list.d/cassandra.list<<EOF
deb http://ftp.fr.debian.org/debian/ sid main non-free contrib # needed for java8
deb http://www.apache.org/dist/cassandra/debian 30x main
deb-src http://www.apache.org/dist/cassandra/debian 30x main
EOF

gpg --keyserver pgp.mit.edu --recv-keys F758CE318D77295D
gpg --export --armor F758CE318D77295D | apt-key add -

gpg --keyserver pgp.mit.edu --recv-keys 2B5C1B00
gpg --export --armor 2B5C1B00 | apt-key add -

gpg --keyserver pgp.mit.edu --recv-keys 0353B12C
gpg --export --armor 0353B12C | apt-key add -

apt-get update
apt-get install -y cassandra cassandra-tools

chown -R cassandra: /etc/cassandra
mkdir /data
chown cassandra: /data
