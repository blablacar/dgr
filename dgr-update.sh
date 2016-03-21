#!/bin/bash
#set -x
set -e
dir=$( dirname $0 )
command -v jq >/dev/null 2>&1 || { echo >&2 "I require jq but it's not installed. Aborting."; exit 1; }


if [[  "$OSTYPE" =~ ^linux ]]; then
    platform="linux-amd64"
#elif [[ "$OSTYPE" == "darwin"* ]]; then
#    platform="darwin-amd64"
#elif [[ "$OSTYPE" == "cygwin" ]]; then
#    platform="windows-amd64"
#elif [[ "$OSTYPE" == "msys" ]]; then
#    platform="windows-amd64"
#elif [[ "$OSTYPE" == "win32" ]]; then
#    platform="windows-amd64"
else
    echo "Only linux is supported"
    exit 1
fi

url_data=$(curl -s https://api.github.com/repos/blablacar/dgr/releases\?access_token\=72a6176d6ebb890caee21ce453f76eae9c836699 )
url=$(echo $url_data | jq -r -c '.[0].assets[] | select(.name | contains("'$platform'")).browser_download_url') 
version=$(echo $url_data | jq -r -c '.[0].tag_name')

if [ -f "$dir/.last_dgr" ]; then
    last_dgr=`cat $dir/.last_dgr`
fi

if [ "$version" != "$last_dgr" ]; then
	wget -O ${dir}/dgr.tar.gz $url
	tar xvzf ${dir}/dgr.tar.gz -C ${dir}
	rm ${dir}/dgr.tar.gz
	echo $version > $dir/.last_dgr
fi
