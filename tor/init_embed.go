// +build embed

package tor

import (
	"embed"
	"io/ioutil"
	"os"

	"github.com/cretz/bine/tor"
)

//go:embed bin/tor
var fs embed.FS

var torPath string

// init configures onionpipe to use an embedded tor static binary. The binary
// is extracted to a temporary directory and executed from there. This is
// currently a work in progress and not yet ready for distribution.
func init() {
	inf, err := fs.Open("tor")
	if err != nil {
		panic(err)
	}
	defer inf.Close()
	outf, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	defer outf.Close()
	torPath = outf.Name()

	_, err = io.Copy(outf, inf)
	if err != nil {
		panic(err)
	}
	err = os.Chmod(torPath, 0755)
	if err != nil {
		panic(err)
	}
	processOption = func(c *tor.StartConf) {
		c.ExePath = torPath
	}
}
