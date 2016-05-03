#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

app="cassandra"

version=`dpkg -s ${app} | grep 'Version: ' | cut -d' ' -f 2`
user=$(cat /etc/passwd | grep ${app} | cut -f3 -d:)
group=$(cat /etc/passwd | grep ${app} | cut -f4 -d:)

cat > /dgr/builder/attributes/version.yml <<EOF
default:
  user: ${user}
  group: ${group}
  version: ${version}
EOF

exit 0
