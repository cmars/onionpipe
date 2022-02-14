package tor

import (
	"context"
	"encoding/base32"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cretz/bine/tor"
)

var processOption Option

// Tor represents a controller over a local Tor node.
type Tor = tor.Tor

// StartConf configures the managed Tor daemon.
type StartConf struct {
	tor.StartConf

	ClientAuths []ClientAuth
}

// ClientAuth represents client authorization needed to connect to
// auth-protected onion services.
type ClientAuth struct {
	// OnionID is the base32-encoded onion ID (without the .onion suffix) which
	// the client is authenticating to.
	OnionID string
	// PrivateKey is a 32-byte x25519 private key, which only the client
	// should know.
	PrivateKey []byte
}

// Option is an option that configures Tor.
type Option func(*StartConf)

// Debug configures Tor to write debug log messages to the given writer.
func Debug(w io.Writer) Option {
	return func(c *StartConf) {
		c.DebugWriter = w
	}
}

// NonAnonymous configures Tor to publish non-anonymous services. This allows
// trading anonymity for a possible performance increase as less hops are used
// with this option.
func NonAnonymous(c *StartConf) {
	c.NoAutoSocksPort = true
	c.ExtraArgs = append(c.ExtraArgs,
		"--HiddenServiceSingleHopMode", "1",
		"--HiddenServiceNonAnonymousMode", "1",
		"--SocksPort", "0",
	)
}

// ClientAuths configures Tor with client authorizations needed in order to
// connect to protected onion services.
func ClientAuths(clientAuths ...ClientAuth) Option {
	return func(c *StartConf) {
		c.ClientAuths = clientAuths
	}
}

// Start starts a new Tor process.
func Start(ctx context.Context, options ...Option) (*tor.Tor, error) {
	torConf := &StartConf{}
	if processOption != nil {
		processOption(torConf)
	}
	for i := range options {
		options[i](torConf)
	}
	t, err := tor.Start(ctx, &torConf.StartConf)
	if err != nil {
		return nil, fmt.Errorf("failed to start tor: %w", err)
	}
	err = configureClientAuth(t, torConf)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func configureClientAuth(t *tor.Tor, conf *StartConf) error {
	if len(conf.ClientAuths) == 0 {
		return nil
	}
	clientsDir := filepath.Join(t.DataDir, "clients")
	err := os.MkdirAll(clientsDir, 0700)
	if err != nil {
		return err
	}
	for _, clientAuth := range conf.ClientAuths {
		f, err := os.Create(filepath.Join(clientsDir, clientAuth.OnionID+".auth_private"))
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = fmt.Fprintf(f, "%s:descriptor:x25519:%s",
			clientAuth.OnionID,
			base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(clientAuth.PrivateKey),
		)
		if err != nil {
			return err
		}
	}
	_, err = t.Control.SendRequest(fmt.Sprintf("SETCONF ClientOnionAuthDir=%s", clientsDir))
	if err != nil {
		return fmt.Errorf("failed to configure client auth: %w", err)
	}
	return err
}
