#!/cnt/bin/bats -x

@test "this will fail" {
  run echo "hello!"
  [ "$status" -eq 1 ]
}
