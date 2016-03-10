#!/bin/bash
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

export PATH=/dgr/bin:$PATH

execute_tests() {
  fdir=$1
  [ -d "$fdir" ] || return 0

  for file in $fdir/*; do
    filename=$(basename $file)
    [ "$filename" == "wait.sh" ] && continue

    echo -e "\e[1m\e[32mRunning test file -> $filename\e[0m"
    res=$(/dgr/bin/bats -t $file)
    res_code=$?
    echo "$res" > /tests-result/${filename}
    echo "$res_code" > /tests-result/${filename}_status

    if [ "$res_code" == "0" ]; then
      echo_green "$res"
    else
      echo_red "$res"
    fi
  done
}

if [ "$(ls -A /path/to/directory)" ]; then
    execute_tests "/dgr/tests"
fi
touch /tests-result/end-of-tests