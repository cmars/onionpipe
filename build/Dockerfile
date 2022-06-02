FROM golang:1.17 as golang

FROM debian:11 as build
COPY --from=golang /usr/local/go /usr/local/go
ENV PATH=/usr/local/go/bin:$PATH
ENV GOPATH=${GOPATH:-/go}
ENV DEBIAN_FRONTEND=noninteractive

# Install Tor build dependencies & tools
RUN apt-get update -qq && \
    apt-get install -y libssl-dev libevent-dev zlib1g-dev \
        automake autoconf build-essential ca-certificates \
        git libtool

VOLUME /go/src/onionpipe
WORKDIR /go/src/onionpipe
