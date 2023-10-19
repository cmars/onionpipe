#!/usr/bin/env bash
set -eux

TOR_RELEASE=0.4.8

cd $(dirname $0)/..
project_root=$(pwd)

tmp_prefix=$(mktemp -d)
#trap "rm -rf $tmp_prefix" EXIT
ln -s /usr/include $tmp_prefix/include
ln -s /usr/lib/x86_64-linux-gnu $tmp_prefix/lib

# Clone Tor target release
mkdir -p $tmp_prefix/src
cd $tmp_prefix/src
git clone --depth 1 https://git.torproject.org/tor.git -b release-${TOR_RELEASE} libtor
cd $tmp_prefix/src/libtor

# Configure and build Tor
./autogen.sh
# Avoid symbol conflicts with openssl (not sure why autoconf doesn't pick this up)
export CFLAGS="-DHAVE_SSL_SESSION_GET_MASTER_KEY -DHAVE_SSL_GET_SERVER_RANDOM -DHAVE_SSL_GET_CLIENT_RANDOM -DHAVE_SSL_GET_CLIENT_CIPHERS"
# Configure and build Tor
./configure --disable-asciidoc \
    --enable-static-libevent \
    --enable-static-openssl \
    --enable-static-zlib \
    --with-libevent-dir=/usr/local/opt/libevent \
    --with-openssl-dir=/usr/local/opt/openssl@1.1 \
    --with-zlib-dir=/usr/local/opt/zlib
make

mkdir -p $project_root/tor/lib $project_root/tor/include $project_root/tor/bin
cp ./libtor.a $project_root/tor/lib
cp ./src/feature/api/tor_api.h $project_root/tor/include
cp /usr/local/opt/libevent/lib/*.a $project_root/tor/lib
cp /usr/local/opt/openssl@1.1/lib/*.a $project_root/tor/lib
cp /usr/local/opt/zlib/lib/*.a $project_root/tor/lib
