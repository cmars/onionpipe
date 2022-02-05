package config

import (
	"fmt"
	"net"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
)

func TestEndpoint(t *testing.T) {
	c := qt.New(t)

	// Set up a local UNIX socket for tests
	socketDir := c.Mkdir()
	socketPath := filepath.Join(socketDir, "server.sock")
	ln, err := net.Listen("unix", socketPath)
	c.Assert(err, qt.IsNil)
	c.Cleanup(func() { c.Assert(ln.Close(), qt.IsNil) })

	tests := []struct {
		name          string
		in            string
		dest          bool
		parsed        *Endpoint
		parseErr      string
		asOnion       bool
		resolved      *Endpoint
		resolveErr    string
		singleAddr    string
		singleAddrErr string
	}{{
		/* Happy-path test cases; semantically valid endpoint definitions */
		name:          "implicit onion dest, single port",
		in:            "8000",
		dest:          true,
		parsed:        &Endpoint{ports: []int{8000}, dest: true},
		asOnion:       true,
		resolved:      &Endpoint{ports: []int{8000}, dest: true, onion: true, resolved: true},
		singleAddrErr: "onion destination",
	}, {
		name:          "implicit onion dest, multi port",
		in:            "80,8000,8080,8888",
		dest:          true,
		parsed:        &Endpoint{ports: []int{80, 8000, 8080, 8888}, dest: true},
		asOnion:       true,
		resolved:      &Endpoint{ports: []int{80, 8000, 8080, 8888}, dest: true, onion: true, resolved: true},
		singleAddrErr: "onion destination",
	}, {
		name:       "implicit local addr dest, single port",
		in:         "8000",
		dest:       true,
		parsed:     &Endpoint{ports: []int{8000}, dest: true},
		asOnion:    false,
		resolved:   &Endpoint{host: "127.0.0.1", ports: []int{8000}, dest: true, resolved: true},
		singleAddr: "127.0.0.1:8000",
	}, {
		name:          "implicit local addr dest, multi port",
		in:            "80,8000,8080,8888",
		dest:          true,
		parsed:        &Endpoint{ports: []int{80, 8000, 8080, 8888}, dest: true},
		asOnion:       false,
		resolveErr:    "local network address may only specify a single port",
		singleAddrErr: "endpoint does not represent a single address",
	}, {
		name:       "onion src, single port",
		in:         "xxx.onion:31337",
		dest:       false,
		parsed:     &Endpoint{host: "xxx.onion", ports: []int{31337}, dest: false},
		asOnion:    false,
		resolved:   &Endpoint{host: "xxx.onion", ports: []int{31337}, dest: false, onion: true, resolved: true},
		singleAddr: "xxx.onion:31337",
	}, {
		name:       "explicit local src, single port",
		in:         "10.1.1.1:22",
		dest:       false,
		parsed:     &Endpoint{host: "10.1.1.1", ports: []int{22}, dest: false},
		asOnion:    false,
		resolved:   &Endpoint{host: "10.1.1.1", ports: []int{22}, dest: false, resolved: true},
		singleAddr: "10.1.1.1:22",
	}, {
		name:       "explicit local src, resolve name, single port",
		in:         "localhost:25565",
		dest:       false,
		parsed:     &Endpoint{host: "localhost", ports: []int{25565}, dest: false},
		asOnion:    false,
		resolved:   &Endpoint{host: "127.0.0.1", ports: []int{25565}, dest: false, resolved: true},
		singleAddr: "127.0.0.1:25565",
	}, {
		name:       "unix dest",
		in:         socketPath,
		dest:       true,
		parsed:     &Endpoint{path: socketPath, dest: true},
		asOnion:    false,
		resolved:   &Endpoint{path: socketPath, dest: true, resolved: true},
		singleAddr: socketPath,
	}, {
		name:       "unix src",
		in:         socketPath,
		dest:       false,
		parsed:     &Endpoint{path: socketPath},
		asOnion:    false,
		resolved:   &Endpoint{path: socketPath, resolved: true},
		singleAddr: socketPath,
	}, {
		name:          "aliased onion dest, single port",
		in:            "8000@wallabag",
		dest:          true,
		parsed:        &Endpoint{ports: []int{8000}, dest: true, alias: "wallabag"},
		asOnion:       true,
		resolved:      &Endpoint{ports: []int{8000}, dest: true, alias: "wallabag", onion: true, resolved: true},
		singleAddrErr: "onion destination",
	}, {
		name:          "aliased onion dest, multi port",
		in:            "80,8000,8080@discourse",
		dest:          true,
		parsed:        &Endpoint{ports: []int{80, 8000, 8080}, dest: true, alias: "discourse"},
		asOnion:       true,
		resolved:      &Endpoint{ports: []int{80, 8000, 8080}, dest: true, alias: "discourse", onion: true, resolved: true},
		singleAddrErr: "onion destination",
	}, {
		/* Semantically invalid endpoints */
		name:       "onion dest w/host, single port",
		in:         "xxx.onion:8000",
		dest:       true,
		parsed:     &Endpoint{host: "xxx.onion", ports: []int{8000}, dest: true},
		asOnion:    true,
		resolveErr: `onion addresses may only be specified as source`,
	}, {
		name:       "implicit onion src, single port",
		in:         "31337",
		dest:       false,
		parsed:     &Endpoint{ports: []int{31337}},
		asOnion:    true,
		resolveErr: `onion source requires an onion address`,
	}, {
		name:     "local src, no port",
		in:       "localhost",
		dest:     false,
		parseErr: `invalid endpoint "localhost"`,
	}, {
		name:     "local dest, no port",
		in:       "localhost",
		dest:     true,
		parseErr: `invalid endpoint "localhost"`,
	}, {
		name:     "unix dest w/port",
		in:       socketPath + ":8000",
		dest:     true,
		parseErr: `UNIX socket does not exist: .*`,
	}, {
		name:     "unix src w/port",
		in:       socketPath + ":8000",
		dest:     false,
		parseErr: `UNIX socket does not exist: .*`,
	}, {
		name:     "aliased non-onion dest",
		in:       "1.2.3.4:5432@postgres",
		dest:     true,
		parseErr: "only remote onions can be aliased",
	}, {
		/* Syntactically invalid */
		name:     "empty w/alias",
		in:       "@what",
		dest:     true,
		parseErr: "only remote onions can be aliased",
	}, {
		name:     "empty w/out alias",
		in:       "",
		dest:     true,
		parseErr: "missing value",
	}, {
		name:     ":",
		in:       ":",
		dest:     true,
		parseErr: `invalid port.*`,
	}, {
		name:     "::",
		in:       "::",
		dest:     true,
		parseErr: `invalid port.*`,
	}}
	for i, test := range tests {
		c.Run(fmt.Sprintf("%d %s", i, test.name), func(c *qt.C) {
			endp, err := ParseEndpoint(test.in, test.dest)
			if test.parseErr != "" {
				c.Check(err, qt.ErrorMatches, test.parseErr)
				return
			} else {
				c.Check(err, qt.IsNil)
			}

			c.Assert(endp, qt.CmpEquals(cmp.AllowUnexported(Endpoint{})), test.parsed)
			err = endp.Resolve(test.asOnion)
			if test.resolveErr != "" {
				c.Check(err, qt.ErrorMatches, test.resolveErr)
				return
			} else {
				c.Check(err, qt.IsNil)
			}

			c.Assert(endp, qt.CmpEquals(cmp.AllowUnexported(Endpoint{})), test.resolved)
			singleAddr, err := endp.SingleAddr()
			if test.singleAddrErr != "" {
				c.Check(err, qt.ErrorMatches, test.singleAddrErr)
			} else {
				c.Check(err, qt.IsNil)
				c.Check(singleAddr, qt.Equals, test.singleAddr)
			}
		})
	}
}
