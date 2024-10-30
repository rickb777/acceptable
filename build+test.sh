#!/bin/bash -e
cd "$(dirname $0)"
PATH=$HOME/go/bin:$PATH
unset GOPATH
export GO111MODULE=on
export GOARCH=${1}

function v
{
  echo
  echo $@
  $@
}

if ! type -p goveralls; then
  v go get     github.com/mattn/goveralls
  v go install github.com/mattn/goveralls
fi

if ! type -p shadow; then
  v go get     golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
  v go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
fi

if ! type -p goreturns; then
  v go get     github.com/sqs/goreturns
  v go install github.com/sqs/goreturns
fi

go test ./...

echo acceptable...
go test -v -covermode=count -coverprofile=test.out .
go tool cover -func=test.out
[ -z "$COVERALLS_TOKEN" ] || goveralls -coverprofile=test.out -service=travis-ci -repotoken $COVERALLS_TOKEN

for d in *; do
  if [ -f $d/doc.go ]; then
    echo $d...
    go test -v -covermode=count -coverprofile=$d/test.out ./$d
    go tool cover -func=$d/test.out
    [ -z "$COVERALLS_TOKEN" ] || goveralls -coverprofile=$d/test.out -service=travis-ci -repotoken $COVERALLS_TOKEN
  fi
done

v goreturns -l -w *.go */*.go

v go vet ./...

v shadow ./...
