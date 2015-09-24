#!/cnt/bin/busybox sh
set -x
export TARGET=$( dirname $0 )
export ROOTFS=%%ROOTFS%%
export TERM=xterm

export PATH=/cnt/bin:$PATH

busybox --install &> /dev/null

execute_tests() {
  fdir=$1
  [ -d "$fdir" ] || return 0

  for file in $fdir/*; do
    filename=$(basename $file)
    echo -e "\e[1m\e[32mRunning test file -> $filename\e[0m"
    /cnt/bin/bats $file
    echo $? > /result/$filename
  done
}

execute_tests "/tests"
