# oniongrok

Onion addresses for anything.

`oniongrok` forwards ports on the local host to remote Onion addresses as Tor
hidden services and vice-versa.

> How do I install it?

Currently supported on: Linux, Darwin, Android (gomobile)

    go install github.com/cmars/oniongrok/cmd/oniongrok@latest

This will take a long time because it is compiling tor and its dependencies
into your binary with CGO.

> What can I do with it?

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
```

> What features are planned?

* Client authentication tokens
* Reverse-forwarding (proxy Tor hidden services into the local network)
* Configurable hops policy (trade anonymity for performance)
* Persistent addresses.

For example:

```
# Forward auth-protected remote onion port 22 to localhost port 2222.
oniongrok --auth hunter2 xxx.onion:22=2222

# Forward local port 22, requiring auth to connect (token will be displayed)
oniongrok --auth-generate 22

# Forward remote onion port 80 to localhost port 80
oniongrok xxx.onion:80

# Forward remote onion port 80 to local port 80 on all interfaces
oniongrok xxx.onion:80=0.0.0.0:80

# Persistent key stored as "myhttpserver" to $XDG_DATA_HOME/oniongrok/myhttpserver
oniongrok 8000=80@myhttpserver
```

I'd also like to support more platforms, and eventually some package managers
(brew, NixOS, choco).

> Why would I want to use this?

oniongrok is a purely decentralized way to create virtually unstoppable global
network tunnels.

For example, you might want to securely publish and access a personal service
from anywhere in the world, across all sorts of network obstructions -- your
ISP doesn't allow ingress traffic to your home lab, your clients might be in
heavily firewalled environments (public WiFi, mobile tether), etc.

With oniongrok, that service doesn't need a public IPv4 or IPv6 ingress. You'll
eventually be able to restrict access with auth tokens. And you don't need to
rely on 3rd-party services (like Tailscale, ZeroTier, etc.) to get to it.

> How can I contribute?

Pull requests are welcome in implementing the above wishlist / planned
functionality.

Otherwise, I'd only ask that you donate to the Tor project, with your dollar,
or by hosting honest proxies and exit nodes. If you like and use this project,
it only further proves the point that Tor is legitimate, essential, useful
public infrastructure that benefits us all and needs our support.
