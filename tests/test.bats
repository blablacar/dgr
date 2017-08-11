#!/bin/bats -x

@test "should fail with no name" {
  run $DGR_PATH -W without_name build
  echo -e "$output"
  [ "$status" -eq 1 ]
  echo "$output" | grep "name is mandatory in manifest"
}

@test "should fail if not exists" {
  run $DGR_PATH -W DOES_NOT_EXISTS build
  echo -e "$output"
  [ "$status" -eq 1 ]
  echo "$output" | grep "Cannot construct aci or pod"
}

@test "should be runnable with only name" {
  run $DGR_PATH -W only_name build
  echo -e "$output"
  [ "$status" -eq 0 ]
}

@test "should be globally working" {
  run $DGR_PATH -W filled_up test
  echo -e "$output"
  [ "$status" -eq 0 ]

  grep 500m filled_up/target/manifest.json
  grep umount2 filled_up/target/manifest.json
}

@test "dgr init should create working aci" {
  rm -Rf /tmp/aci-init
  mkdir -p /tmp/aci-init
  run $DGR_PATH -W /tmp/aci-init init
  echo -e "$output"
  [ "$status" -eq 0 ]
  run $DGR_PATH -W /tmp/aci-init test
  echo -e "$output"
  rm -Rf /tmp/aci-init
  [ "$status" -eq 0 ]
}

@test "should see when a test fail" {
  run $DGR_PATH -W with_failed_test test
  echo -e "$output"
  echo "$output" | grep "Failed test"
  [ "$status" -eq 1 ]
}

@test "should see when a test suceed" {
  run $DGR_PATH -W with_successful_test test
  echo -e "$output"
  echo "$output" | grep "succeed"
  [ "$status" -eq 0 ]
}
@test "tester should work when there is no env var set" {
  run $DGR_PATH -W no_env_test test
  echo -e "$output"
  echo "$output" | grep "ok 1 Check value of TEST_VAR"
  [ "$status" -eq 0 ]
}
@test "tester should work when there is only apps env var set" {
  run $DGR_PATH -W app_only_env_test test
  echo -e "$output"
  echo "$output" | grep "ok 1 Check value of TEST_VAR"
  [ "$status" -eq 0 ]
}

@test "tester should work when there is 2 conflicting env var for the tester" {
  run $DGR_PATH -W double_env_test test
  echo -e "$output"
  echo "$output" | grep "ok 1 Check value of TEST_VAR"
  [ "$status" -eq 0 ]
}
