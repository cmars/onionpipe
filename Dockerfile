# Build image
FROM debian:12 AS tor
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update -qq && apt-get install -y apt-transport-https wget gnupg
RUN echo "deb [signed-by=/usr/share/keyrings/tor-archive-keyring.gpg] https://deb.torproject.org/torproject.org bookworm main" >/etc/apt/sources.list.d/tor.list
RUN wget -qO- https://deb.torproject.org/torproject.org/A3C4F0F979CAA22CDBA8F512EE8CBC9E886DDD89.asc | gpg --dearmor | tee /usr/share/keyrings/tor-archive-keyring.gpg >/dev/null
RUN apt-get update -qq && apt-get install -y tor deb.torproject.org-keyring

FROM golang:1.22-bookworm as build
WORKDIR /src
COPY go.* /src/
RUN go mod download
COPY . /src/
ENV SKIP_FORWARDING_TESTS=1
RUN make all test

# Deploy image
FROM tor
RUN useradd --create-home -d /data -s /bin/bash onionpipe
COPY --from=build /src/onionpipe /onionpipe
VOLUME [ "/data" ]
WORKDIR /data
USER onionpipe
ENTRYPOINT [ "/onionpipe" ]
