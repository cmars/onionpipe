# onionpipe

Onion addresses for anything.

`onionpipe` forwards ports on the local host to remote Onion addresses as Tor
hidden services and vice-versa.

### Why would I want to use this?

onionpipe is a decentralized way to create virtually unstoppable global network
tunnels.

For example, you might want to securely publish and access a personal service
from anywhere in the world, across all sorts of network obstructions -- your
ISP doesn't allow ingress traffic to your home lab, your clients might be in
heavily firewalled environments (public WiFi, mobile tether), etc.

With onionpipe, that service doesn't need a public IPv4 or IPv6 ingress. You
can publish services with a globally-unique persistent onion address, and share
access securely and privately to your own allowlist of authorized keys.

You don't need to rely on, and share your personal data with for-profit
services (like Tailscale, ZeroTier, etc.) to get to it.

### What can I do with it right now?

onionpipe sets up socket forwarding tunnels. It's like `socat(1)`, for onions.

#### Export services on local networks to onion addresses

Export localhost port 8000 to a temporary, one-time remote onion address.
```
onionpipe 8000
```

Export localhost port 8000 to temporary remote onion port 80. `~` is shorthand
for the forward between source~destination.
```
onionpipe 8000~80
```

Export localhost port 8000 to a persistent remote onion address nicknamed
'my-app'.
```
onionpipe 8000~80@my-app
```

Nicknames can be re-used in multiple forwarding expressions to reference the
same onion address. Let's set up a little web forum for our Minecraft server.
```
onionpipe 8000~80@minecraft 25565@minecraft
```

All the forwards without nicknames use the same temporary address.
```
onionpipe 192.168.1.100:8000~80,8080,9000 9090
```

Export a UNIX socket to an onion address.
```
onionpipe /run/server.sock~80
```

Export to a non-anonymous remote onion service, trading network privacy for
possibly reduced latency.
```
onionpipe --anonymous=false 8000
```

#### Import onion services to local network interfaces.

Import a remote onion's port 80 to localhost port 80.
```
onionpipe xxx.onion:80
```

Import remote onion port 80 to local port 80 on all interfaces. This can be
used for creating an ingress to the onion on public networks.
```
onionpipe xxx.onion:80~0.0.0.0:80
```

Running with Docker is simple and easy, the only caveat is that its the
container forwarding, so adjust local addresses accordingly.

Forward port 80 on Docker host.
```
docker run --rm ghcr.io/cmars/onionpipe:main host.docker.internal:80
```

If you're using Podman, exposing the local host network is another option.
```
podman run --network=host --rm ghcr.io/cmars/onionpipe:main 8000 
```

Because local forwarding addresses are DNS resolved, it's very easy to publish
hidden services from within Docker Compose or K8s. Check out this
[nextcloud](examples/nextcloud/docker-compose.yml) example (watch the log for
the onion address)!

#### Client auth
[Client auth](https://community.torproject.org/onion-services/advanced/client-auth/)
is great for securing personal services over Tor. How to use it:

Alice creates a new client auth public key pair.

```
onionpipe client new alice
```

```
{
  "alice": {
    "identity": "p2pof7vumwsrqqavtovfwqqaw6cqzvtqqe7cjvxt754k6j7blufa"
  }
}
```

Alice shares this public key with Bob, who forwards an onion service that only
Alice can use.

```
onionpipe --require-auth p2pof7vumwsrqqavtovfwqqaw6cqzvtqqe7cjvxt754k6j7blufa 8000~80@test
```

```
2022/02/13 21:25:46 starting tor...
127.0.0.1:8000 => sd6aq2r6jvuoeisrudq7jbqufjh6nck5buuzjmgalicgwrobgfj4lkqd.onion:80
```

Alice can use her client private key to connect to this onion and forward to a
local port.

```
onionpipe --auth alice sd6aq2r6jvuoeisrudq7jbqufjh6nck5buuzjmgalicgwrobgfj4lkqd.onion:80~7000
```

```
2022/02/13 21:29:17 starting tor...
sd6aq2r6jvuoeisrudq7jbqufjh6nck5buuzjmgalicgwrobgfj4lkqd.onion:80 => 127.0.0.1:7000
```

### How do I install it?

Each commit into main triggers an automated release, which publishes a Homebrew
tap and Docker image.

#### Homebrew

    brew tap cmars/onionpipe
    brew install onionpipe

#### Docker

The provided `Dockerfile` builds a minimal image that can run onionpipe in a
container with the latest Tor release from the Tor Project. Build and runtime
is Debian-based.

#### Local build

In a local clone of this project,

    make onionpipe

The built binary `onionpipe` will require a `tor` daemon executable to be in
your `$PATH`.

#### Static standalone binary with libtor

See the [latest release](https://github.com/cmars/onionpipe/releases/latest)
for Linux and Darwin binaries.

Should theoretically work on: Linux, Darwin, Android (gomobile) according to
the [berty.tech/go-libtor](https://github.com/berty/go-libtor) README. There
are some quirks; see comments in `tor/init_libtor.go` for details.

In a local clone of this project,

    make onionpipe_libtor

This will take a long time the first time you build, because it compiles CGO
wrappers for Tor and its dependencies.

You'll need to have C library dependencies installed for the build to work:

- tor
- openssl
- libevent
- zlib

If you're on NixOS, you can run `nix-shell` in this directory to get these
dependencies installed into your shell context.

### What features are planned?

Declare forwards and operate from a yaml file rather than CLI arguments.

```
onionpipe --config config.yaml
```

Considering a fancy TUI.

Considering a control plane for onionpipe SDN orchestration.

Stay tuned.

### How can I contribute?

Donate to the Tor project with your dollar, or by hosting honest proxies and
exit nodes. If you like and use this project, support the public infrastructure
that benefits us all and makes this wonderful magic possible.
