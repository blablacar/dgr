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

@test "should understand base64 override" {
  export TEMPLATER_OVERRIDE_BASE64=base64,eyJ5b2xvIjoidGVzdCIsInRlc3QiOiJ5b2xvIn0=
  run ../../dist/templater  -o "TEMPLATER_OVERRIDE_BASE64" -t ./target 4/
  echo -e "$output"
  [ "$status" -eq 0 ]
}

@test "should understand Json override" {
  export TEMPLATER_OVERRIDE='{"yolo":"test","test":"yolo"}'
  run ../../dist/templater  -o "TEMPLATER_OVERRIDE" -t ./target 4/
  echo -e "$output"
  [ "$status" -eq 0 ]
}

@test "should template with a symlink as attributes" {
  run ../../dist/templater -t ./target 5/
  echo -e "$output"
  [ "$status" -eq 0 ]
}
