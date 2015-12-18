#!/bin/bats -x

@test "should fail if no name" {
  run $CNT_PATH build
  echo -e "$output"
  [ "$status" -eq 1 ]
  [ echo $output | grep "foo: no such file 'nonexistent_filename'" ]
}
