#!/dgr/bin/bats -x

@test "Check value of TEST_VAR*" {
  [ ! -z ${TEST_VAR_APP} ]
  [ ! -z ${TEST_VAR} ]
  [ ${TEST_VAR_APP} == "true" ]
  [ ${TEST_VAR} == "false" ]
}
