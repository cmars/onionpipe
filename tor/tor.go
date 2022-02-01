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
