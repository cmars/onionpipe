#!/usr/bin/env bash
set -eux

cd $(dirname $0)/..

ONIONPIPE_BIN=./dist/onionpipe-$(go env GOOS)-$(go env GOARCH)-static
go build -o $ONIONPIPE_BIN -v -x -tags "embed" .
if [ "$(go env GOOS)" == "linux" ]; then
    strip $ONIONPIPE_BIN
fi
