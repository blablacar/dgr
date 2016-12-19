#!/bin/bash
set -x
set -e

if [ "$(id -u)" != "0" ]; then
    echo "Sorry, you are not root."
    exit 1
fi

execute_tests() {
  fdir=$1
  [ -d "$fdir" ] || return 0

  for file in $fdir/*; do
    filename=$(basename $file)
    [ "$filename" == "wait.sh" ] && continue
    [ -d "$file" ] && continue
    [ "$filename" == "test.sh" ] && continue

    echo -e "\e[1m\e[32mRunning test file -> $filename\e[0m"
    bats -p $file
  done
}

#dgr
command -v rkt >/dev/null 2>&1 || { echo >&2 "rkt not found in path"; exit 1; }
command -v bats >/dev/null 2>&1 || { echo >&2 "bats not found in path"; exit 1; }
dir=$( dirname $0 )
cd ${dir}
execute_tests "."
