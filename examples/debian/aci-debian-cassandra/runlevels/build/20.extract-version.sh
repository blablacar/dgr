#!/dgr/bin/busybox sh

export CASSANDRA_VERSION=`dpkg -s cassandra | grep 'Version: ' | cut -d' ' -f 2`
echo -e "default:\n  version: ${CASSANDRA_VERSION}" > /dgr/builder/attributes/version.yml
