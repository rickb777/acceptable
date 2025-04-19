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

go install tool

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

v gofmt -l -w *.go */*.go

v go vet ./...
