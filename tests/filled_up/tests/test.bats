#!/dgr/bin/bats -x

@test "just an echo" {
  run echo "hello!"
  [ "$status" -eq 0 ]
}
