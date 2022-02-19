package app

import (
	"encoding/json"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/cmars/onionpipe/secrets"
)

func TestClientCommands(t *testing.T) {
	c := qt.New(t)
	home := c.Mkdir()
	c.Setenv("HOME", home)

	c.Run("list clients", func(c *qt.C) {
		in, out, err := os.Pipe()
		c.Assert(err, qt.IsNil)
		c.Patch(&os.Stdout, out)
		go func() {
			defer out.Close()
			err := App().Run([]string{"onionpipe", "client"})
			c.Assert(err, qt.IsNil)
		}()
		var clients secrets.ClientsPublic
		err = json.NewDecoder(in).Decode(&clients)
		c.Assert(err, qt.IsNil)
		c.Assert(clients, qt.HasLen, 0)
	})
	c.Run("add/rm clients", func(c *qt.C) {
		err := App().Run([]string{"onionpipe", "client", "new", "test"})
		c.Assert(err, qt.IsNil)
		err = App().Run([]string{"onionpipe", "client", "new", "test"})
		c.Assert(err, qt.ErrorMatches, `client "test" already exists`)
		err = App().Run([]string{"onionpipe", "client", "new", "test2"})
		c.Assert(err, qt.IsNil)
		err = App().Run([]string{"onionpipe", "client", "new", "test3"})
		c.Assert(err, qt.IsNil)
		err = App().Run([]string{"onionpipe", "client", "rm", "test"})
		c.Assert(err, qt.IsNil)
		in, out, err := os.Pipe()
		c.Assert(err, qt.IsNil)
		c.Patch(&os.Stdout, out)
		go func() {
			defer out.Close()
			err := App().Run([]string{"onionpipe", "client"})
			c.Assert(err, qt.IsNil)
		}()
		var clients secrets.ClientsPublic
		err = json.NewDecoder(in).Decode(&clients)
		c.Assert(err, qt.IsNil)
		c.Assert(clients, qt.HasLen, 2)
		c.Assert(clients["test"].Identity, qt.Equals, "")
		c.Assert(clients["test2"].Identity, qt.Not(qt.Equals), "")
		c.Assert(clients["test3"].Identity, qt.Not(qt.Equals), "")
	})
}
