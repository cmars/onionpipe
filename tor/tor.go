package tor

import (
	"context"
	"fmt"
	"io"

	"github.com/cretz/bine/tor"
)

var processOption Option

// Option is an option that configures Tor.
type Option func(*tor.StartConf)

// Debug configures Tor to write debug log messages to the given writer.
func Debug(w io.Writer) Option {
	return func(c *tor.StartConf) {
		c.DebugWriter = w
	}
}

// NonAnonymous configures Tor to publish non-anonymous services. This allows
// trading anonymity for a possible performance increase as less hops are used
// with this option.
func NonAnonymous(c *tor.StartConf) {
	c.NoAutoSocksPort = true
	c.ExtraArgs = append(c.ExtraArgs,
		"--HiddenServiceSingleHopMode", "1",
		"--HiddenServiceNonAnonymousMode", "1",
		"--SocksPort", "0",
	)
}

// Start starts a new Tor process.
func Start(ctx context.Context, options ...Option) (*tor.Tor, error) {
	torConf := &tor.StartConf{}
	if processOption != nil {
		processOption(torConf)
	}
	for i := range options {
		options[i](torConf)
	}
	t, err := tor.Start(ctx, torConf)
	if err != nil {
		return nil, fmt.Errorf("failed to start tor: %w", err)
	}
	return t, nil
}
