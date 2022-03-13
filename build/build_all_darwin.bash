#!/usr/bin/env bash
set -eux
cd $(dirname $0)

brew install \
    pkg-config autoconf@2.69 automake \
    openssl@1.1 \
    libevent \
    zlib

if [ ! -e ../tor/lib/libtor.a ]; then
    ./build_tor_darwin.bash
fi

./build_onionpipe.bash
