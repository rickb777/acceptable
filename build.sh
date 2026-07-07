#!/bin/bash -ex
cd "$(dirname "$0")"
go install tool
mage build coverage
cat report.out

go run ./examples/stdlibExample/stdlibExample.go &
sleep 0.5
./examples/curl-commands.sh
