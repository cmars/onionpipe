package forwarding

import (
	"context"
	"fmt"
	"io"

	"berty.tech/go-libtor"
	"github.com/cretz/bine/tor"
)

// TorOption is an option that configures Tor.
type TorOption func(*tor.StartConf)

// TorDebug configures Tor to write debug log messages to the given writer.
func TorDebug(w io.Writer) TorOption {
	return func(c *tor.StartConf) {
		c.DebugWriter = w
	}
}

// StartTor starts a new Tor process.
func StartTor(ctx context.Context, options ...TorOption) (*tor.Tor, error) {
	torConf := &tor.StartConf{
		ProcessCreator: libtor.Creator,
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
