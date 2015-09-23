#!/cnt/bin/busybox sh
set -x
set -e
export TARGET=$( dirname $0 )
export ROOTFS=%%ROOTFS%%
export TERM=xterm

export PATH=/cnt/bin:$PATH

busybox --install

execute_tests() {
  fdir=$1
  [ -d "$fdir" ] || return 0

  for file in $fdir/*; do
    echo -e "\e[1m\e[32mRunning test file -> $file\e[0m"
    /cnt/bin/bats $file
  done
}

execute_tests "/tests"
