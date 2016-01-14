@test "should be globally working" {
  run $CNT_PATH -W filled_up test
  echo -e "$output"
  [ "$status" -eq 0 ]
}

@test "cnt init should create working aci" {
  mkdir /tmp/aci-init
  run $CNT_PATH -W /tmp/aci-init init
  echo -e "$output"
  [ "$status" -eq 0 ]
  run $CNT_PATH -W /tmp/aci-init test
  echo -e "$output"
  rm -Rf /tmp/aci-init
  [ "$status" -eq 0 ]
}

@test "should see when a test fail" {
  run $CNT_PATH -W with_failed_test test
  echo -e "$output"
  echo "$output" | grep "Failed test"
  [ "$status" -eq 2 ]
}
