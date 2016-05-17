#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

app="openjdk-7-jre"

version=`dpkg -s ${app} | grep 'Version: ' | cut -d' ' -f 2`

cat > /dgr/builder/attributes/version.yml <<EOF
default:
  version: ${version}
EOF

exit 0
