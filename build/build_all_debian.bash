#!/usr/bin/env bash
set -eux
cd $(dirname $0)

docker build -t onionpipe-embed-linux:latest -f Dockerfile .
if [ ! -e ../tor/lib/libtor.a ]; then
    docker run --rm -v $(pwd)/..:/go/src/onionpipe onionpipe-embed-linux:latest bash -eux ./build/build_tor_debian.bash
fi
docker run --rm -v $(pwd)/..:/go/src/onionpipe onionpipe-embed-linux:latest bash -eux ./build/build_onionpipe.bash
