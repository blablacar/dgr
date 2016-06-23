#!/bin/bats

setup() {
    rm -Rf ./target
    mkdir -p ./target
}

@test "should template" {
  run ../../dist/templater -t ./target 1/
  echo -e "$output"
  [ "$status" -eq 0 ]
  ls ./target/fff/yopla2.yml
  ls ./target/fff/yopla.yml
  ! ls ./target/fff/yopla.yml.cfg
#  echo "$output" | grep ""
}

@test "should fail on no-value" {
  run ../../dist/templater -t ./target 2/
  echo -e "$output"
  [ "$status" -eq 1 ]
}

@test "should template non string" {
  run ../../dist/templater -t ./target 3/
  echo -e "$output"
  [ "$status" -eq 0 ]
}

