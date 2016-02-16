#!/dgr/bin/busybox sh
set -x
export TARGET=$( dirname $0 )
export ROOTFS=%%ROOTFS%%
export TERM=xterm

echo_red() {
  echo -e "\033[0;31m${1}\033[0m"
}

echo_green() {
  echo -e "\033[0;32m${1}\033[0m"
}

export PATH=/dgr/bin:$PATH
busybox --install &> /dev/null

execute_tests() {
  fdir=$1
  [ -d "$fdir" ] || return 0

  for file in $fdir/*; do
    filename=$(basename $file)
    [ "$filename" == "wait.sh" ] && continue

    echo -e "\e[1m\e[32mRunning test file -> $filename\e[0m"
    res=$(/dgr/bin/bats -t $file)
    res_code=$?
    echo "$res" > /result/${filename}
    echo "$res_code" > /result/${filename}_status

    if [ "$res_code" == "0" ]; then
      echo_green "$res"
    else
      echo_red "$res"
    fi
  done
}

execute_tests "/tests"
