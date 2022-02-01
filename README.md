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

### What can I do with it right now?

Currently, you can publish local services as ephemeral Tor hidden services.

For example:

```
# Forward localhost port 8000 to remote onion port 8000
oniongrok 8000

# Forward localhost port 8000 to remote onion port 80.
# ~ is shorthand for the forward between source~destination.
oniongrok 8000~80

# Forward local interface 8000 to remote onion ports 80, 8080
# and forward local port 9090 to remote port 9090.
oniongrok 192.168.1.100:8000~80,8080,9000 9090

# Forward remote onion port 80 to localhost port 80
oniongrok xxx.onion:80

# Forward remote onion port 80 to local port 80 on all interfaces
oniongrok xxx.onion:80~0.0.0.0:80
```

Running with Docker is simple and easy, the only caveat is that its the
container forwarding, so adjust local addresses accordingly. For example:

```
# Forward port 80 on Docker host
docker run --rm ghcr.io/cmars/oniongrok:main host.docker.internal:80
```

If you're using Podman, exposing the local host network is another option.

    podman run --network=host --rm ghcr.io/cmars/oniongrok:main 8000 

Because local forwarding addresses are DNS resolved, it's very easy to publish
hidden services from within Docker Compose or K8s. Check out this
[nextcloud](examples/nextcloud/docker-compose.yml) example (watch the log for
the onion address)!

### How do I build it?

#### Docker

The provided `Dockerfile` builds a minimal image that can run oniongrok in a
container with the latest Tor release from the Tor Project. Build and runtime
is Debian-based.

#### Local build

In a local clone of this project,

    make oniongrok

The built binary `oniongrok` will require a `tor` daemon executable to be in
your `$PATH`.

#### Static standalone binary with libtor

Should theoretically work on: Linux, Darwin, Android (gomobile) according to
the [berty.tech/go-libtor](https://github.com/berty/go-libtor) README. There
are some quirks; see comments in `tor/init_libtor.go` for details.

In a local clone of this project,

    make oniongrok_libtor

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

* UNIX socket support
* Client authentication tokens
* Configurable hops policy (trade anonymity for performance)
* Persistent addresses.
* Option to define forwards in a JSON or YAML config file

For example:

```
# Forward local UNIX socket to remote onion port.
oniongrok /run/server.sock~80

# Forward auth-protected remote onion port 22 to localhost port 2222.
oniongrok --auth hunter2 xxx.onion:22~2222

# Forward local port 22, requiring auth to connect (token will be displayed)
oniongrok --auth-generate 22

# Persistent key stored as "myhttpserver" to $XDG_DATA_HOME/oniongrok/myhttpserver
oniongrok 8000~80@myhttpserver

# Operate from a yaml file.
oniongrok --config config.yaml
```

Considering support for distributions: NixOS, brew & choco

### How can I contribute?

Pull requests are welcome in implementing the above wishlist / planned
functionality.

Otherwise, donate to the Tor project with your dollar, or by hosting honest
proxies and exit nodes. If you like and use this project, support the public
infrastructure that benefits us all.
