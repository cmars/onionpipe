FROM golang:1.17-buster AS build
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update -qq && apt-get install -y apt-transport-https wget gnupg
RUN echo "deb [signed-by=/usr/share/keyrings/tor-archive-keyring.gpg] https://deb.torproject.org/torproject.org buster main" >/etc/apt/sources.list.d/tor.list
RUN wget -qO- https://deb.torproject.org/torproject.org/A3C4F0F979CAA22CDBA8F512EE8CBC9E886DDD89.asc | gpg --dearmor | tee /usr/share/keyrings/tor-archive-keyring.gpg >/dev/null
RUN apt-get update -qq && apt-get install -y tor deb.torproject.org-keyring libssl-dev libevent-dev libz-dev
COPY . /src
WORKDIR /src
RUN make all
WORKDIR /data
RUN touch /data/.build

FROM gcr.io/distroless/base
COPY --from=build /src/oniongrok /oniongrok
COPY --from=build --chown=1000 /data /data
WORKDIR /data
USER 1000
ENTRYPOINT [ "/oniongrok" ]
