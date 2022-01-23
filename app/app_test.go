package app_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/cmars/oniongrok/app"
)

func TestPortMap(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		in  string
		out *app.PortMap
		err string
	}{{
		in:  "8000",
		out: &app.PortMap{Local: 8000, Remote: []int{8000}},
	}, {
		in:  "8000:80",
		out: &app.PortMap{Local: 8000, Remote: []int{80}},
	}, {
		in:  "8000:80,8000,8080,8888",
		out: &app.PortMap{Local: 8000, Remote: []int{80, 8000, 8080, 8888}},
	}, {
		in:  "",
		err: `invalid local port ""`,
	}, {
		in:  ":",
		err: `invalid local port ":"`,
	}, {
		in:  "foo",
		err: `invalid local port "foo"`,
	}, {
		in:  "80:foo",
		err: `invalid remote port "foo"`,
	}}
	for _, test := range tests {
		pm, err := app.NewPortMap(test.in)
		if test.err != "" {
			c.Check(err, qt.ErrorMatches, test.err)
		} else {
			c.Check(err, qt.IsNil)
			c.Check(pm, qt.DeepEquals, test.out)
		}
	}
}
