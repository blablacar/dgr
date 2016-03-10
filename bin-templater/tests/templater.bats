#!/bin/bats

setup() {
    rm -Rf ./target
    mkdir -p ./target
}

@test "should template" {
  run ../../dist/templater -t ./target 1/
  echo -e "$output"
  [ "$status" -eq 0 ]
#  echo "$output" | grep ""
}
