#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

version=${ACI_VERSION%-*}

curl -o ${ROOTFS}/jmx_prometheus_httpserver.jar https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_httpserver/${version}/jmx_prometheus_httpserver-${version}-jar-with-dependencies.jar
