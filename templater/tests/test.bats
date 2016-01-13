#!/bin/bats

setup() {
    rm -Rf ./target
    mkdir -p ./target
}

@test "should template nothing" {
  run ../templater -t ./target 1/attributes 1/templates
  echo -e "$output"
  [ "$status" -eq 1 ]
  echo "$output" | grep "name is mandatory in manifest"
}
