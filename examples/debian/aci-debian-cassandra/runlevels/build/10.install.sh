#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

cat > /etc/apt/sources.list.d/cassandra.list<<EOF
deb http://ftp.fr.debian.org/debian/ jessie main non-free contrib # needed for java8
deb http://www.apache.org/dist/cassandra/debian 30x main
deb-src http://www.apache.org/dist/cassandra/debian 30x main
deb [check-valid-until=no] http://archive.debian.org/debian jessie-backports main
EOF

gpg --keyserver pool.sks-keyservers.net --recv-keys F758CE318D77295D || \
gpg --keyserver pgp.mit.edu --recv-keys F758CE318D77295D
gpg --export --armor F758CE318D77295D | apt-key add -

gpg --keyserver pool.sks-keyservers.net --recv-keys 2B5C1B00 || \
gpg --keyserver pgp.mit.edu --recv-keys 2B5C1B00
gpg --export --armor 2B5C1B00 | apt-key add -

gpg --keyserver pool.sks-keyservers.net --recv-keys 0353B12C || \
gpg --keyserver pgp.mit.edu --recv-keys 0353B12C
gpg --export --armor 0353B12C | apt-key add -

# For Cassandra
apt-key adv --keyserver pool.sks-keyservers.net --recv-key A278B781FE4B2BDA

# Jessie is too old, so we add `-o Acquire::Check-Valid-Until=false` (for now)
apt-get -o Acquire::Check-Valid-Until=false update
apt-get -o Acquire::Check-Valid-Until=false install -y --force-yes -t jessie-backports cassandra cassandra-tools

chown -R cassandra: /etc/cassandra
mkdir /data
chown cassandra: /data
