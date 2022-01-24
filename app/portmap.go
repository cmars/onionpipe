package app

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type ExportMap struct {
	LocalAddr   string
	LocalPort   int
	RemotePorts []int
}

type ImportMap struct {
	RemoteAddr string
	RemotePort int
	LocalAddr  string
	LocalPort  int
}

func ParsePortMap(s string) (*ExportMap, *ImportMap, error) {
	localRemote := strings.SplitN(s, "=", 2)
	if len(localRemote) == 1 {
		host, ports, err := parseEndpoint(localRemote[0])
		if err != nil {
			return nil, nil, err
		}
		if len(ports) > 1 {
			return nil, nil, fmt.Errorf("cannot forward multiple ports in a single expression %q", s)
		}
		if isOnion(host) {
			return nil, &ImportMap{
				RemoteAddr: host,
				RemotePort: ports[0],
				LocalAddr:  "127.0.0.1",
				LocalPort:  ports[0],
			}, nil
		}
		if host == "" {
			host = "127.0.0.1"
		} else {
			host, err = resolveHost(host)
			if err != nil {
				return nil, nil, err
			}
		}
		return &ExportMap{
			LocalAddr:   host,
			LocalPort:   ports[0],
			RemotePorts: []int{ports[0]},
		}, nil, nil
	}
	if len(localRemote) == 2 {
		fromHost, fromPorts, err := parseEndpoint(localRemote[0])
		if err != nil {
			return nil, nil, err
		}
		if len(fromPorts) > 1 {
			return nil, nil, fmt.Errorf("cannot forward multiple ports in a single expression %q", s)
		}
		toHost, toPorts, err := parseEndpoint(localRemote[1])
		if err != nil {
			return nil, nil, err
		}
		if isOnion(fromHost) && isOnion(toHost) {
			return nil, nil, fmt.Errorf("forwarding from onion to onion not supported")
		}
		if isOnion(fromHost) {
			if toHost == "" {
				toHost = "127.0.0.1"
			}
			if len(toPorts) > 1 {
				return nil, nil, fmt.Errorf("cannot forward multiple ports in a single expression %q", s)
			}
			return nil, &ImportMap{
				RemoteAddr: fromHost,
				RemotePort: fromPorts[0],
				LocalAddr:  toHost,
				LocalPort:  toPorts[0],
			}, nil
		} else if toHost != "" {
			return nil, nil, fmt.Errorf("invalid remote address %q", s)
		}
		if fromHost == "" {
			fromHost = "127.0.0.1"
		} else {
			fromHost, err = resolveHost(fromHost)
			if err != nil {
				return nil, nil, err
			}
		}
		if len(fromPorts) > 1 {
			return nil, nil, fmt.Errorf("cannot forward multiple ports in single expression %q", s)
		}
		return &ExportMap{
			LocalAddr:   fromHost,
			LocalPort:   fromPorts[0],
			RemotePorts: toPorts,
		}, nil, nil
	}
	return nil, nil, fmt.Errorf("invalid port map expression %q", s)
}

func resolveHost(host string) (string, error) {
	addrs, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", host)
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", fmt.Errorf("could not resolve %q", host)
	}
	return addrs[0].String(), nil
}

func isOnion(s string) bool {
	return strings.HasSuffix(s, ".onion")
}

func parseEndpoint(s string) (string, []int, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 1 {
		ports, err := parsePorts(parts[0])
		if err != nil {
			return "", nil, err
		}
		return "", ports, nil
	}
	if len(parts) == 2 {
		ports, err := parsePorts(parts[1])
		if err != nil {
			return "", nil, err
		}
		return parts[0], ports, nil
	}
	return "", nil, fmt.Errorf("invalid endpoint %q", s)
}

func parsePorts(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid port(s) %q", s)
	}
	var ports []int
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		port, err := strconv.Atoi(parts[i])
		if err != nil {
			return nil, fmt.Errorf("invalid port number %q", parts[i])
		}
		ports = append(ports, port)
	}
	if len(ports) == 0 {
		return nil, fmt.Errorf("missing port number")
	}
	return ports, nil
}
