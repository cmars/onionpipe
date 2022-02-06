package secrets

import (
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var secretsCmp = qt.CmpEquals(cmpopts.IgnoreUnexported(Secrets{}))

func TestNoSecrets(t *testing.T) {
	c := qt.New(t)
	dir := c.Mkdir()
	path := dir + "/sec.json"
	sec, err := ReadFile(path)
	c.Assert(err, qt.IsNil)
	c.Assert(sec, secretsCmp, &Secrets{
		Version: "1",
		path:    path,
	})
	err = sec.WriteFile()
	c.Assert(err, qt.IsNil)
	_, err = os.Stat(path)
	c.Assert(os.IsNotExist(err), qt.IsTrue)
}

func TestSecrets(t *testing.T) {
	c := qt.New(t)
	dir := c.Mkdir()
	path := dir + "/sec.json"
	sec, err := ReadFile(path)
	c.Assert(err, qt.IsNil)
	c.Assert(sec, secretsCmp, &Secrets{
		Version: "1",
		path:    path,
	})
	fooKey, err := sec.EnsureServiceKey("foo")
	c.Assert(err, qt.IsNil)
	barKey, err := sec.EnsureServiceKey("bar")
	c.Assert(err, qt.IsNil)

	err = sec.WriteFile()
	c.Assert(err, qt.IsNil)
	_, err = os.Stat(path)
	c.Assert(err, qt.IsNil)

	sec2, err := ReadFile(path)
	c.Assert(err, qt.IsNil)
	fooKey2, err := sec2.EnsureServiceKey("foo")
	c.Assert(err, qt.IsNil)
	barKey2, err := sec2.EnsureServiceKey("bar")
	c.Assert(err, qt.IsNil)
	c.Assert(fooKey, qt.DeepEquals, fooKey2)
	c.Assert(barKey, qt.DeepEquals, barKey2)
}
