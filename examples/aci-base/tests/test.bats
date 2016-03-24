#!/dgr/bin/bats -x

@test "should simple start" {
  run ls
  [ "$status" -eq 0 ]
}
