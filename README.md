# oniongrok

Onion addresses for anything.

`oniongrok` forwards local ports to remote Onion addresses as Tor hidden
services.

Usage is simple:

```
# Forward local port 8000 to remote onion port 80
oniongrok 8000:80

# Forward local port 8000 to remote onion ports 80, 8080
# and forward local port 9090 to remote port 9090.
oniongrok 8000:80,8080 9090

# Forward remote onion port 22 to local port 2222
oniongrok --client-auth hunter2 xxx.onion:22:2222

# Forward local port 22, requiring auth (token will be displayed)
oniongrok --require-auth 127.0.0.1:22

# Forward remote onion port 80 to all interfaces port 80
oniongrok xxx.onion:80:0.0.0.0:2222

```

Forwards syntax:

(localaddrport | localport)(:remoteport) | (remoteaddrport):(localaddrport | localport)

