package config

import (
	"fmt"
	"net"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
)

func TestForward(t *testing.T) {
	c := qt.New(t)

	// Set up a local UNIX socket for tests
	socketDir := c.Mkdir()
	socketPath := filepath.Join(socketDir, "server.sock")
	ln, err := net.Listen("unix", socketPath)
	c.Assert(err, qt.IsNil)
	c.Cleanup(func() { c.Assert(ln.Close(), qt.IsNil) })

	tests := []struct {
		name     string
		in       string
		parsed   *Forward
		parseErr string
	}{{
		/* Happy-path test cases; semantically valid forward definitions */
		name: "single port",
		in:   "8080",
		parsed: &Forward{
			src: &Endpoint{
				host:     "127.0.0.1",
				ports:    []int{8080},
				resolved: true,
			},
			dest: &Endpoint{
				ports:    []int{8080},
				onion:    true,
				dest:     true,
				resolved: true,
			},
		},
	}, {
		name: "single port, mapped",
		in:   "8080~80",
		parsed: &Forward{
			src: &Endpoint{
				host:     "127.0.0.1",
				ports:    []int{8080},
				resolved: true,
			},
			dest: &Endpoint{
				ports:    []int{80},
				onion:    true,
				dest:     true,
				resolved: true,
			},
		},
	}, {
		name: "multi port, mapped",
		in:   "8080~80,8080,8888",
		parsed: &Forward{
			src: &Endpoint{
				host:     "127.0.0.1",
				ports:    []int{8080},
				resolved: true,
			},
			dest: &Endpoint{
				ports:    []int{80, 8080, 8888},
				onion:    true,
				dest:     true,
				resolved: true,
			},
		},
	}, {
		name: "resolve localhost, multi port, mapped",
		in:   "localhost:8080~80,8080,8888",
		parsed: &Forward{
			src: &Endpoint{
				host:     "127.0.0.1",
				ports:    []int{8080},
				resolved: true,
			},
			dest: &Endpoint{
				ports:    []int{80, 8080, 8888},
				onion:    true,
				dest:     true,
				resolved: true,
			},
		},
	}, {
		name: "local net addr, multi port, mapped",
		in:   "10.0.0.1:8080~80,8080,8888",
		parsed: &Forward{
			src: &Endpoint{
				host:     "10.0.0.1",
				ports:    []int{8080},
				resolved: true,
			},
			dest: &Endpoint{
				ports:    []int{80, 8080, 8888},
				onion:    true,
				dest:     true,
				resolved: true,
			},
		},
	}, {
		name: "unix addr, multi port, mapped",
		in:   socketPath + "~80,8080,8888",
		parsed: &Forward{
			src: &Endpoint{
				path:     socketPath,
				resolved: true,
			},
			dest: &Endpoint{
				ports:    []int{80, 8080, 8888},
				onion:    true,
				dest:     true,
				resolved: true,
			},
		},
	}, {
		name: "onion to local net",
		in:   "xxx.onion:80~8000",
		parsed: &Forward{
			src: &Endpoint{
				host:     "xxx.onion",
				ports:    []int{80},
				onion:    true,
				resolved: true,
			},
			dest: &Endpoint{
				host:     "127.0.0.1",
				ports:    []int{8000},
				dest:     true,
				resolved: true,
			},
		},
	}, {
		name: "onion to local unix",
		in:   "xxx.onion:80~" + socketPath,
		parsed: &Forward{
			src: &Endpoint{
				host:     "xxx.onion",
				ports:    []int{80},
				onion:    true,
				resolved: true,
			},
			dest: &Endpoint{
				path:     socketPath,
				dest:     true,
				resolved: true,
			},
		},
	}, {
		/* Semantically invalid forwards */
		name:     "multi port",
		in:       "80,81,82",
		parseErr: `.*: local network address may only specify a single port`,
	}, {
		name:     "multi multi",
		in:       "80,81,82~80,8080,8888",
		parseErr: `.*: local network address may only specify a single port`,
	}, {
		name:     "chaining",
		in:       "80~81~82",
		parseErr: `.*: invalid endpoint "81~82"`,
	}, {
		name:     "empty",
		in:       "",
		parseErr: `.*: missing value`,
	}, {
		name:     "local to local",
		in:       "10.0.0.1:8080~192.168.1.1:8888",
		parseErr: `.*: onion addresses may only be specified as source`,
	}}
	for i, test := range tests {
		c.Run(fmt.Sprintf("%d %s", i, test.name), func(c *qt.C) {
			fwd, err := ParseForward(test.in)
			if test.parseErr != "" {
				c.Check(err, qt.ErrorMatches, test.parseErr)
				return
			} else {
				c.Check(err, qt.IsNil)
			}
			c.Assert(fwd, qt.CmpEquals(cmp.AllowUnexported(Forward{}, Endpoint{})), test.parsed)
		})
	}
}
