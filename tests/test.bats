#!/bin/bats -x

@test "should fail with no name" {
  run $CNT_PATH -W empty build
  echo -e "$output"
  [ "$status" -eq 1 ]
  echo "$output" | grep "name is mandatory in manifest"
}

@test "should fail if not exists" {
  run $CNT_PATH -W DOES_NOT_EXISTS build
  echo -e "$output"
  [ "$status" -eq 1 ]
  echo "$output" | grep "Cannot construct aci or pod"
}
