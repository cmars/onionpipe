// Package config provides the forwarding configuration data models used in
// this application.
package config

import (
	"context"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Endpoint represents a source or destination endpoint in a forwarding tunnel.
// The source may be regarded as the "backend" which provides a service being
// forwarded. The destination may be regarded as the "frontend" where the
// service is forwarded to, or published for consumption.
type Endpoint struct {
	host  string
	ports []int
	path  string

	dest     bool
	onion    bool
	resolved bool
}

// EndpointDoc defines a JSON representation of an endpoint.
type EndpointDoc struct {
	Host  string `json:"host"`
	Ports []int  `json:"ports"`
	Path  string `json:"unix"`
}

// Endpoint returns a validated and resolved Endpoint from a JSON document
// object model.
func (d *EndpointDoc) Endpoint(dest, asOnion bool) (*Endpoint, error) {
	e := &Endpoint{
		host:  d.Host,
		ports: d.Ports,
		path:  d.Path,
		dest:  dest,
	}
	err := e.Resolve(asOnion)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// IsOnionHost returns whether the host is a .onion address.
func IsOnionHost(s string) bool {
	return strings.HasSuffix(s, ".onion")
}

// IsDest returns whether a resolved endpoint is a destination. If it's not a
// destination, it's a source in a forward.
func (e *Endpoint) IsDest() bool {
	return e.dest
}

// IsOnion returns whether a resolved endpoint is an onion address.
func (e *Endpoint) IsOnion() bool {
	return e.onion
}

// Resolve validates the endpoint to ensure it is well-formed. For UNIX socket
// endpoints, the socket path is validated for existence. For local network
// TCP socket endpoints, the host is resolved to a network address according to
// the system's default name resolver.
//
// asOnion is provided to disambiguate an endpoint that does not yet have its
// host resolved.
//
// Such name resolution is provided here as a convenience -- if it doesn't
// resolve how you like, or there's a chance it will choose the wrong
// interface, configure the actual interface you want. The default resolver is
// generally Good Enough for container use cases, where the networking and
// routing is handled up the stack in something like compose or k8s. It might
// be a bit YOLO-networking on a bare metal server with several physical NICs
// and multiple DNS names. Caveat emp-tor.
func (e *Endpoint) Resolve(asOnion bool) error {
	if e.resolved {
		return nil
	}

	// Resolving onions
	if asOnion || IsOnionHost(e.host) {
		if len(e.ports) == 0 {
			e.ports = []int{80}
		}
		if len(e.path) > 0 {
			return fmt.Errorf("invalid onion endpoint")
		}
		if e.dest && len(e.host) > 0 {
			return fmt.Errorf("onion addresses may only be specified as source")
		}
		if !e.dest && len(e.host) == 0 {
			return fmt.Errorf("onion source requires an onion address")
		}
	}
	if IsOnionHost(e.host) {
		if e.dest {
			return fmt.Errorf("onion addresses may only be specified as a source")
		}
		if len(e.ports) != 1 {
			return fmt.Errorf("onion source address may only specify a single port")
		}
		e.resolved = true
		e.onion = true
		return nil
	} else if asOnion {
		if !e.dest {
			return fmt.Errorf("invalid onion address")
		}
		if e.host != "" {
			return fmt.Errorf("onion destination may only specify ports")
		}
		e.resolved = true
		e.onion = true
		return nil
	}

	// Resolving UNIX sockets
	if e.path != "" {
		if e.host != "" || len(e.ports) > 0 {
			return fmt.Errorf("ambiguous endpoint: must be either a UNIX socket or TCP address")
		}
		if st, err := os.Stat(e.path); err != nil {
			return err
		} else if st.Mode()&os.ModeSocket == 0 {
			return fmt.Errorf("not a UNIX socket: %s", e.path)
		}
		e.resolved = true
		return nil
	}

	// Resolving local TCP addresses
	if len(e.ports) != 1 {
		return fmt.Errorf("local network address may only specify a single port")
	}
	if e.host == "" {
		e.host = "127.0.0.1"
		e.resolved = true
		return nil
	}
	addrs, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", e.host)
	if err != nil {
		return err
	}
	if len(addrs) == 0 {
		return fmt.Errorf("could not resolve %q", e.host)
	}
	e.host = addrs[0].String()
	e.resolved = true
	return nil
}

// SingleAddr returns the string representation of the endpoint as a single
// address in a port or socket forward. Some endpoints do not have such a
// representation, in which case an error is returned.
func (e *Endpoint) SingleAddr() (string, error) {
	if e.onion && e.dest {
		return "", fmt.Errorf("onion destination")
	}
	switch len(e.ports) {
	case 0:
		if e.path != "" {
			return e.path, nil
		}
	case 1:
		if e.host != "" {
			return fmt.Sprintf("%s:%d", e.host, e.ports[0]), nil
		}
	default:
		return "", fmt.Errorf("endpoint does not represent a single address")
	}
	return "", fmt.Errorf("unresolved endpoint")
}

// Description returns a string description of the endpoint. If the endpoint is
// an onion destination, the remote onion ID may be provided to render its
// assigned hostname.
func (e *Endpoint) Description(remoteOnion string) string {
	if e.onion && e.dest {
		ports := make([]string, len(e.ports))
		for i := range e.ports {
			ports[i] = strconv.Itoa(e.ports[i])
		}
		return fmt.Sprintf("%s.onion:%s", remoteOnion, strings.Join(ports, ","))
	}
	addr, err := e.SingleAddr()
	if err != nil {
		return fmt.Sprintf("%%!(BADADDR %v)", err)
	}
	return addr
}

// Ports returns the port numbers assigned to this endpoint.
func (e *Endpoint) Ports() []int {
	return e.ports
}

var remotePortsRE = regexp.MustCompile(`^\d+(,\d+)*$`)

// ParseEndpoint returns an Endpoint from the given string representation and
// whether it is intended to be used as a destination or source in a forward.
func ParseEndpoint(s string, dest bool) (*Endpoint, error) {
	if s == "" {
		return nil, fmt.Errorf("missing value")
	}

	// Check for a remote port list to be exported
	if remotePortsRE.MatchString(s) {
		ports, err := parsePortList(s)
		if err != nil {
			return nil, err
		}
		return &Endpoint{ports: ports, dest: dest}, nil
	}

	// Check for a local UNIX socket
	if _, err := os.Stat(s); err == nil {
		return &Endpoint{
			path: s,
			dest: dest,
		}, nil
	} else {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if s[0] == '/' {
			return nil, fmt.Errorf("UNIX socket does not exist: %s", s)
		}
	}

	// Otherwise endpoint is some form of host:port then
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid endpoint %q", s)
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}
	endp := &Endpoint{
		host:  parts[0],
		ports: []int{port},
		dest:  dest,
	}
	return endp, nil
}

func parsePortList(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid port(s) %q", s)
	}
	var ports []int
	for i := range parts {
		port, err := strconv.Atoi(parts[i])
		if err != nil {
			return nil, fmt.Errorf("invalid port number %q", parts[i])
		}
		ports = append(ports, port)
	}
	return ports, nil
}
