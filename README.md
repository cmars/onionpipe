# oniongrok

Onion addresses for anything.

`oniongrok` forwards ports on the local host to remote Onion addresses as Tor
hidden services and vice-versa.

### Why would I want to use this?

oniongrok is a decentralized way to create virtually unstoppable global network
tunnels.

For example, you might want to securely publish and access a personal service
from anywhere in the world, across all sorts of network obstructions -- your
ISP doesn't allow ingress traffic to your home lab, your clients might be in
heavily firewalled environments (public WiFi, mobile tether), etc.

With oniongrok, that service doesn't need a public IPv4 or IPv6 ingress. You'll
eventually be able to restrict access with auth tokens. And you don't need to
rely on, and share your personal data with for-profit services (like Tailscale,
ZeroTier, etc.) to get to it.

### How do I build it?

#### Local build

Should work on: Linux, Darwin, Android (gomobile) according to the
[berty.tech/go-libtor](https://github.com/berty/go-libtor) README.

In a local clone of this project,

    make

This will take a long time the first time you build, because it compiles CGO
wrappers for Tor and its dependencies.

You'll need to have C library dependencies installed for the build to work:

- tor
- openssl
- libevent
- zlib

If you're on NixOS, you can run `nix-shell` in this directory to get these
dependencies installed into your shell context.

#### Docker

The provided `Dockerfile` builds a minimal image that can run oniongrok in a
container. Build is Debian-based, runtime is distroless.

### What can I do with it right now?

Currently, you can publish local services as ephemeral Tor hidden services.

For example:

```
# Forward localhost port 8000 to remote onion port 8000
oniongrok 8000

# Forward localhost port 8000 to remote onion port 80
oniongrok 8000=80

# Forward local interface 8000 to remote onion ports 80, 8080
# and forward local port 9090 to remote port 9090.
oniongrok 192.168.1.100:8000=80,8080,9000

# Forward remote onion port 80 to localhost port 80
oniongrok xxx.onion:80

# Forward remote onion port 80 to local port 80 on all interfaces
oniongrok xxx.onion:80=0.0.0.0:80

```

Running with Docker is similar, though you'll need to provide local interface
IP addresses that route inside the container. Currently these are not DNS
resolved, but if you know the address of your host on the container,

    docker run --rm ghcr.io/cmars/oniongrok:main 192.168.65.1:8000

If you're using Podman, exposing the local host network is another option.

    podman run --network=host --rm ghcr.io/cmars/oniongrok:main 8000 

### What features are planned?

* Client authentication tokens
* Configurable hops policy (trade anonymity for performance)
* Persistent addresses.
* Resolve names for local bind addresses, useful for Docker/Compose/k8s use
  cases.
* Option to define forwards in a JSON or YAML config file

For example:

```
# Forward auth-protected remote onion port 22 to localhost port 2222.
oniongrok --auth hunter2 xxx.onion:22=2222

# Forward local port 22, requiring auth to connect (token will be displayed)
oniongrok --auth-generate 22

# Persistent key stored as "myhttpserver" to $XDG_DATA_HOME/oniongrok/myhttpserver
oniongrok 8000=80@myhttpserver

# Forward port 80 on Docker host
docker run --rm ghcr.io/cmars/oniongrok:main host.docker.internal:80

# Operate from a yaml file.
oniongrok --config config.yaml

```

I'd also like to support more platforms, and eventually some package managers
(brew, NixOS, choco if we get Windows).

### How can I contribute?

Pull requests are welcome in implementing the above wishlist / planned
functionality.

Otherwise, donate to the Tor project with your dollar, or by hosting honest
proxies and exit nodes. If you like and use this project, it only further
proves the point that Tor is essential public infrastructure that benefits us
all and needs our support.
