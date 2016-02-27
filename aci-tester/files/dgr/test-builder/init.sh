#!/bin/bash
set -x

%%COMMAND%% &
cd /
waitsuccess="0"
if [ -f "/tests/wait.sh" ]; then
	chmod +x /tests/wait.sh
	i="0"
	while [ $i -lt 60 ]; do
	  /tests/wait.sh
	  if [ $? == 0 ]; then
	  	waitsuccess="1"
	  	break;
	  fi
	  i=$[$i+1]
	  sleep 1
	done
fi

if [ wait_success == "0" ]; then
	echo "1\n" > /result/wait.sh
	echo "WAIT FAILED"
	exit 1
fi

/test.sh
