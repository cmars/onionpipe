// +build libtor

package tor

import (
	"berty.tech/go-libtor"
	"github.com/cretz/bine/tor"
)

// init uses libtor directly in the same process, rather than by controlling a
// tor executable. This configuration allows onionpipe to run as a single executable
// with no unpacking required, but has some quirks:
//
// libtor is not as up-to-date with the latest Tor release. It may be missing
// security and bug fixes as a result. Use with caution. There are also some quirks:
// 1. Signal handling seems to be a little strange. Have to send two in order
// for signal.Notify to see it.
// 2. Shutdown sometimes seems to hang. Could be a race condition?
//
func init() {
	processOption = func(c *tor.StartConf) {
		c.ProcessCreator = libtor.Creator
	}
}
