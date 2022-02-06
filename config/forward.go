package config

import (
	"fmt"
	"strings"
)

// Forward defines a network forwarding relay from a source endpoint to a
// destination endpoint.
type Forward struct {
	src  *Endpoint
	dest *Endpoint
}

// IsImport returns whether the forward is importing an onion to a local
// network endpoint.
func (f *Forward) IsImport() bool {
	return f.src.onion
}

// Source returns the source endpoint in the forwarding relay.
func (f *Forward) Source() *Endpoint {
	return f.src
}

// Destination returns the destination endpoint in the forwarding relay.
func (f *Forward) Destination() *Endpoint {
	return f.dest
}

// ForwardDoc defines a JSON representation of a forward.
type ForwardDoc struct {
	Src  EndpointDoc `json:"src"`
	Dest EndpointDoc `json:"dest"`
}

// Forward returns a validated and resolved Forward from a JSON document object
// model.
func (d *ForwardDoc) Forward() (*Forward, error) {
	ig, err := d.Src.Endpoint(false, IsOnionHost(d.Src.Host))
	if err != nil {
		return nil, fmt.Errorf("forward source: %w", err)
	}
	eg, err := d.Dest.Endpoint(true, !IsOnionHost(d.Src.Host))
	if err != nil {
		return nil, fmt.Errorf("forward destination: %w", err)
	}
	f := &Forward{
		src:  ig,
		dest: eg,
	}
	err = f.Resolve()
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Description returns a string description of the forward. If the forward
// destination is an onion, a mapping of alias to remote onion ID may be
// provided, to render the assigned onion address.
func (f *Forward) Description(remoteOnions map[string]string) string {
	return fmt.Sprintf("%s => %s", f.src.Description(nil), f.dest.Description(remoteOnions))
}

// ParseForward returns a new Forward parsed from a string representation
func ParseForward(s string) (*Forward, error) {
	parts := strings.SplitN(s, "~", 2)
	var src, dest *Endpoint
	var err error
	switch len(parts) {
	case 1:
		src, err = ParseEndpoint(parts[0], false)
		if err != nil {
			return nil, fmt.Errorf("forward source: %w", err)
		}
		dest, err = ParseEndpoint(parts[0], true)
		if err != nil {
			return nil, fmt.Errorf("forward destination: %w", err)
		}
	case 2:
		src, err = ParseEndpoint(parts[0], false)
		if err != nil {
			return nil, fmt.Errorf("forward source: %w", err)
		}
		dest, err = ParseEndpoint(parts[1], true)
		if err != nil {
			return nil, fmt.Errorf("forward destination: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid forward %q", s)
	}

	f := &Forward{
		src:  src,
		dest: dest,
	}
	err = f.Resolve()
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Resolve validates the endpoints of the forward and whether together they
// constitute a valid, well-formed and supported forwarding arrangement.
func (f *Forward) Resolve() error {
	var srcOnion, destOnion bool
	if IsOnionHost(f.src.host) {
		srcOnion = true
	} else {
		destOnion = true
	}
	err := f.src.Resolve(srcOnion)
	if err != nil {
		return fmt.Errorf("forward source: %w", err)
	}
	err = f.dest.Resolve(destOnion)
	if err != nil {
		return fmt.Errorf("forward destination: %w", err)
	}
	if f.src.onion == f.dest.onion {
		return fmt.Errorf("source or destination must be an onion address")
	}
	return nil
}
