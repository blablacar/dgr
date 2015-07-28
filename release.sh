#!/bin/bash
set -x
set -e
dir=$( dirname $0 )

if [ $# != 2 ]; then
    echo "Usage: release.sh VERSION GITHUB_KEY"
    exit 1
fi

version=$1
access_token=$2



require_clean_work_tree () {
    # Update the index
    git update-index -q --ignore-submodules --refresh
    err=0

    # Disallow unstaged changes in the working tree
    if ! git diff-files --quiet --ignore-submodules --
    then
        echo >&2 "cannot $1: you have unstaged changes."
        git diff-files --name-status -r --ignore-submodules -- >&2
        err=1
    fi

    # Disallow uncommitted changes in the index
    if ! git diff-index --cached --quiet HEAD --ignore-submodules --
    then
        echo >&2 "cannot $1: your index contains uncommitted changes."
        git diff-index --cached --name-status -r --ignore-submodules HEAD -- >&2
        err=1
    fi

    if [ $err = 1 ]
    then
        echo >&2 "Please commit or stash them."
        exit 1
    fi
}

${dir}/build.sh
require_clean_work_tree


for i in ${dir}/target/*/ ; do
    if [ -d "$i" ]; then
        cd $i
        platform=${PWD##*/}
        tar cvzf cnt-$platform-$version.tar.gz *
        cd -
    fi
done

git tag $version -a -m "Version $version"
git push --tags

sleep 5

posturl=$(curl --data "{\"tag_name\": \"$1\",\"target_commitish\": \"master\",\"name\": \"$1\",\"body\": \"Release of version $1\",\"draft\": false,\"prerelease\": true}" https://api.github.com/repos/blablacar/cnt/releases?access_token=${access_token} | grep "\"upload_url\"" | sed -ne 's/.*\(http[^"]*\).*/\1/p')

for i in ${dir}/target/*/ ; do
    if [ -d "$i" ]; then
        fullpath=$(ls ${i}/*.tar.gz)
        filename=${fullpath##*/}
        curl -i -X POST -H "Content-Type: application/x-gzip" --data-binary "@${fullpath}" ${posturl}?name=${filename}&access_token=${access_token}
    fi
done

