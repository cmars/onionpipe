package app

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/cmars/oniongrok/config"
	"github.com/cmars/oniongrok/secrets"
	"github.com/cmars/oniongrok/tor"
)

func TestServiceCommands(t *testing.T) {
	c := qt.New(t)
	c.Patch(&startTor, func(_ context.Context, options ...tor.Option) (*tor.Tor, error) {
		return &tor.Tor{}, nil
	})
	c.Patch(&newForwardingService, func(_ *tor.Tor, fwds ...*config.Forward) forwardingService {
		return &mockForwardingService{}
	})
	home := c.Mkdir()
	c.Setenv("HOME", home)
	err := App().Run([]string{"oniongrok", "8080@test"})
	c.Assert(err, qt.IsNil)

	c.Run("list services", func(c *qt.C) {
		in, out, err := os.Pipe()
		c.Assert(err, qt.IsNil)
		c.Patch(&os.Stdout, out)
		go func() {
			defer out.Close()
			err := App().Run([]string{"oniongrok", "service"})
			c.Assert(err, qt.IsNil)
		}()
		var services secrets.ServicesPublic
		err = json.NewDecoder(in).Decode(&services)
		c.Assert(err, qt.IsNil)
		c.Assert(services, qt.HasLen, 1)
		c.Assert(services["test"].Address, qt.Not(qt.Equals), "")
	})
	c.Run("add/rm services", func(c *qt.C) {
		err := App().Run([]string{"oniongrok", "service", "add", "test"})
		c.Assert(err, qt.ErrorMatches, `service "test" already exists`)
		err = App().Run([]string{"oniongrok", "service", "add", "test2"})
		c.Assert(err, qt.IsNil)
		err = App().Run([]string{"oniongrok", "service", "add", "test3"})
		c.Assert(err, qt.IsNil)
		err = App().Run([]string{"oniongrok", "service", "rm", "test"})
		c.Assert(err, qt.IsNil)
		in, out, err := os.Pipe()
		c.Assert(err, qt.IsNil)
		c.Patch(&os.Stdout, out)
		go func() {
			defer out.Close()
			err := App().Run([]string{"oniongrok", "service"})
			c.Assert(err, qt.IsNil)
		}()
		var services secrets.ServicesPublic
		err = json.NewDecoder(in).Decode(&services)
		c.Assert(err, qt.IsNil)
		c.Assert(services, qt.HasLen, 2)
		c.Assert(services["test"].Address, qt.Equals, "")
		c.Assert(services["test2"].Address, qt.Not(qt.Equals), "")
		c.Assert(services["test3"].Address, qt.Not(qt.Equals), "")
	})
}
