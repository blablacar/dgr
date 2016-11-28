#!/bin/bash
set -e
start=$(date +%s)
: ${dir:="$( dirname $0 )"}

go_files=`find . -name '*.go' 2> /dev/null | grep -v dist/ | grep -v vendor/ | grep -v .git`

echo -e "\033[0;32mFormat\033[0m"
gofmt -w -s ${go_files}

echo -e "\033[0;32mFix\033[0m"
go tool fix ${go_files}

echo -e "\033[0;32mErr check\033[0m"

command -v errcheck >/dev/null || go get -u github.com/kisielk/errcheck
errcheck ./... | grep -v vendor

echo -e "\033[0;32mLint\033[0m"
command -v golint >/dev/null || go get -u github.com/golang/lint/golint
for i in ${go_files}; do
    golint ${i}
done

echo -e "\033[0;32mVet\033[0m"
go tool vet ${go_files} || true

echo -e "\033[0;32mMisspell\033[0m"
command -v misspell >/dev/null || go get -u github.com/client9/misspell/cmd/misspell
misspell -source=text ${go_files}

echo -e "\033[0;32mIneffassign\033[0m"
command -v ineffassign >/dev/null || go get -u github.com/gordonklaus/ineffassign
for i in ${go_files}; do
    ineffassign -n ${i} || true
done

echo -e "\033[0;32mGocyclo\033[0m"
command -v gocyclo || go get -u github.com/fzipp/gocyclo
gocyclo -over 15 ${go_files} || true

echo "Quality duration : $((`date +%s`-start))s"
