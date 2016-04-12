#!/dgr/bin/bats -x

@test "Cassandra should be running" {
  nodetool status
  [ $? -eq 0 ]
}
