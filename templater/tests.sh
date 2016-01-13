#!/bin/bash
set -e

pushd `dirname $0` > /dev/null
dir=`pwd`
popd > /dev/null

if [ -z "$DEBUG" ]; then
    trap "rm -Rf ${dir}/tests/*/target/; exit" EXIT HUP INT QUIT PIPE TERM
else
    set -x
fi

execute_tests() {
  fdir=$1
  [ -d "$fdir" ] || return 0

  for file in $fdir/*; do
    filename=$(basename $file)
    [ "$filename" == "wait.sh" ] && continue
    [ -d "$file" ] && continue

    echo -e "\e[1m\e[32mRunning test file -> $filename\e[0m"
    bats -p $file
  done
}

cd "${dir}/tests"
execute_tests "."