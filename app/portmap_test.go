package app_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/cmars/oniongrok/app"
)

func TestPortMap(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		in        string
		importMap *app.ImportMap
		exportMap *app.ExportMap
		err       string
	}{{
		in:        "8000",
		exportMap: &app.ExportMap{LocalAddr: "127.0.0.1", LocalPort: 8000, RemotePorts: []int{8000}},
	}, {
		in:        "8000=80",
		exportMap: &app.ExportMap{LocalAddr: "127.0.0.1", LocalPort: 8000, RemotePorts: []int{80}},
	}, {
		in:        "8000=80,8000,8080,8888",
		exportMap: &app.ExportMap{LocalAddr: "127.0.0.1", LocalPort: 8000, RemotePorts: []int{80, 8000, 8080, 8888}},
	}, {
		in:        "10.0.0.100:8000",
		exportMap: &app.ExportMap{LocalAddr: "10.0.0.100", LocalPort: 8000, RemotePorts: []int{8000}},
	}, {
		in:        "10.0.0.100:8000=80,81,82",
		exportMap: &app.ExportMap{LocalAddr: "10.0.0.100", LocalPort: 8000, RemotePorts: []int{80, 81, 82}},
	}, {
		in:        "foo.onion:8000",
		importMap: &app.ImportMap{RemoteAddr: "foo.onion", RemotePort: 8000, LocalAddr: "127.0.0.1", LocalPort: 8000},
	}, {
		in:        "foo.onion:8000=8001",
		importMap: &app.ImportMap{RemoteAddr: "foo.onion", RemotePort: 8000, LocalAddr: "127.0.0.1", LocalPort: 8001},
	}, {
		in:        "foo.onion:8000=0.0.0.0:8001",
		importMap: &app.ImportMap{RemoteAddr: "foo.onion", RemotePort: 8000, LocalAddr: "0.0.0.0", LocalPort: 8001},
	}, {
		in:  "",
		err: `missing port number`,
	}, {
		in:  ":",
		err: `missing port number`,
	}, {
		in:  "foo",
		err: `invalid port number "foo"`,
	}, {
		in:  "80:foo",
		err: `invalid port number "foo"`,
	}}
	for _, test := range tests {
		exportMap, importMap, err := app.ParsePortMap(test.in)
		if test.err != "" {
			c.Check(err, qt.ErrorMatches, test.err)
		} else if test.exportMap != nil {
			c.Check(err, qt.IsNil)
			c.Check(exportMap, qt.DeepEquals, test.exportMap)
		} else if test.importMap != nil {
			c.Check(err, qt.IsNil)
			c.Check(importMap, qt.DeepEquals, test.importMap)
		}
	}
}
