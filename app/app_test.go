package app

import (
	"context"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/mitchellh/go-homedir"

	"github.com/cmars/oniongrok/config"
	"github.com/cmars/oniongrok/forwarding"
	"github.com/cmars/oniongrok/tor"
)

func init() {
	homedir.DisableCache = true
}

func TestSecretsPath(t *testing.T) {
	c := qt.New(t)
	path := "/path/to/secrets.json"
	c.Assert(secretsPath(path, true), qt.Equals, path)
	c.Assert(secretsPath(path, false), qt.Equals, "/path/to/secrets.not-anonymous.json")
}

func TestApp(t *testing.T) {
	c := qt.New(t)
	c.Patch(&startTor, func(_ context.Context, options ...tor.Option) (*tor.Tor, error) {
		return &tor.Tor{}, nil
	})
	fwdSvc := &mockForwardingService{
		onions: map[string]string{
			"":     "xyz",
			"test": "abc",
		},
	}
	zeroKey := [64]byte{0}
	c.Patch(&newForwardingService, func(_ *tor.Tor, fwds ...*config.Forward) forwardingService {
		fwdSvc.fwds = fwds
		// simulate zeroization
		for _, fwd := range fwds {
			copy(fwd.Destination().ServiceKey(), zeroKey[:])
		}
		return fwdSvc
	})

	c.Run("anonymous ephemeral", func(c *qt.C) {
		home := c.Mkdir()
		c.Setenv("HOME", home)
		fwdSvc.fwds = nil
		err := App().Run([]string{"oniongrok", "8080"})
		c.Assert(err, qt.IsNil)

		c.Assert(fwdSvc.fwds, qt.HasLen, 1)
		c.Assert(fwdSvc.fwds[0].Description(fwdSvc.onions), qt.Equals, "127.0.0.1:8080 => xyz.onion:8080")
		_, err = os.Stat(defaultSecretsPath())
		c.Assert(os.IsNotExist(err), qt.IsTrue)
	})
	c.Run("anonymous persistent", func(c *qt.C) {
		home := c.Mkdir()
		c.Setenv("HOME", home)
		fwdSvc.fwds = nil
		err := App().Run([]string{"oniongrok", "8080@test"})
		c.Assert(err, qt.IsNil)

		c.Assert(fwdSvc.fwds, qt.HasLen, 1)
		c.Assert(fwdSvc.fwds[0].Description(fwdSvc.onions), qt.Equals, "127.0.0.1:8080 => abc.onion:8080")
		_, err = os.Stat(defaultSecretsPath())
		c.Assert(err, qt.IsNil)

		c.Assert(fwdSvc.fwds[0].Destination().ServiceKey(), qt.DeepEquals, zeroKey[:])
	})
	c.Run("non-anonymous persistent", func(c *qt.C) {
		home := c.Mkdir()
		c.Setenv("HOME", home)
		fwdSvc.fwds = nil
		err := App().Run([]string{"oniongrok", "--anonymous=false", "8080@test"})
		c.Assert(err, qt.IsNil)

		c.Assert(fwdSvc.fwds, qt.HasLen, 1)
		c.Assert(fwdSvc.fwds[0].Description(fwdSvc.onions), qt.Equals, "127.0.0.1:8080 => abc.onion:8080")
		_, err = os.Stat(defaultSecretsPath())
		c.Assert(os.IsNotExist(err), qt.IsTrue)
		_, err = os.Stat(home + "/.local/share/oniongrok/secrets.not-anonymous.json")
		c.Assert(err, qt.IsNil)
	})
}

type mockForwardingService struct {
	fwds   []*config.Forward
	onions map[string]string
}

func (*mockForwardingService) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func (m *mockForwardingService) Start(ctx context.Context, options ...forwarding.Option) (map[string]string, error) {
	return m.onions, nil
}
