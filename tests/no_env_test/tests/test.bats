#!/dgr/bin/bats -x

@test "Check value of TEST_VAR*" {
  [  -z ${TEST_VAR} ]
  [  -z ${TEST_VAR_APP} ]
  [  -z ${TEST_VAR_TEST} ]
}
