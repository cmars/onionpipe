package forwarding

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/cretz/bine/tor"
	"golang.org/x/crypto/ed25519"

	"github.com/cmars/oniongrok/config"
)

// Service implements a forwarding service. It sets up and operates forwarding
// as defined by configuration.
type Service struct {
	tor     *tor.Tor
	imports []*config.Forward
	exports []*config.Forward

	nonAnonymous bool
	done         chan struct{}
}

// New returns a new forwarding service.
func New(t *tor.Tor, fwds ...*config.Forward) *Service {
	var imports, exports []*config.Forward
	for _, fwd := range fwds {
		if fwd.IsImport() {
			imports = append(imports, fwd)
		} else {
			exports = append(exports, fwd)
		}
	}
	return &Service{
		tor:     t,
		imports: imports,
		exports: exports,
		done:    make(chan struct{}),
	}
}

// Option is an option that configures Tor.
type Option func(s *Service)

// NonAnonymous configures this service to forward as a non-anonymous service.
// The use of this option requires tor.Start to have been configured with
// tor.NonAnonymous. Import forwards are also not allowed with this option,
// because Tor will not accept Socks proxy connections in this mode.
func NonAnonymous(s *Service) {
	s.nonAnonymous = true
}

// Start starts forwarding.
func (s *Service) Start(ctx context.Context, options ...Option) (map[string]string, error) {
	for i := range options {
		options[i](s)
	}
	// Start import forwarding
	for _, importFwd := range s.imports {
		if s.nonAnonymous {
			return nil, fmt.Errorf("import forwards not supported in non-anonymous single-hop mode")
		}
		err := s.startImporter(ctx, s.tor, importFwd)
		if err != nil {
			return nil, err
		}
	}
	// Start export forwarding
	if len(s.exports) > 0 {
		aliasOnions := map[string]string{}
		onionFwds, err := s.startExporter(ctx)
		if err != nil {
			return nil, err
		}
		for alias, fwd := range onionFwds {
			aliasOnions[alias] = fwd.ID
		}
		return aliasOnions, nil
	}
	return nil, nil
}

// Done returns a channel that closes when the forwarding service is shut down.
// This may be used to wait for all the forwards to close before shutting down
// tor.
func (s *Service) Done() <-chan struct{} {
	return s.done
}

func (s *Service) startImporter(ctx context.Context, t *tor.Tor, fwd *config.Forward) error {
	srcAddr, err := fwd.Source().SingleAddr()
	if err != nil {
		return fmt.Errorf("source: %w", err)
	}
	destAddr, err := fwd.Destination().SingleAddr()
	if err != nil {
		return fmt.Errorf("destination: %w", err)
	}

	l, err := net.Listen("tcp", destAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on local address %q", destAddr)
	}
	remoteDialer, err := t.Dialer(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create tor network dialer")
	}

	go func() {
		for {
			localConn, err := l.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					log.Printf("failed to accept on local address %q", destAddr)
				}
				return
			}
			go func() {
				defer localConn.Close()
				remoteConn, err := remoteDialer.DialContext(ctx, "tcp", srcAddr)
				if err != nil {
					log.Printf("failed to connect to onion address %q", srcAddr)
					return
				}
				defer remoteConn.Close()

				recvDone := make(chan struct{})
				go func() {
					io.Copy(localConn, remoteConn)
					close(recvDone)
				}()
				sendDone := make(chan struct{})
				go func() {
					io.Copy(remoteConn, localConn)
					close(sendDone)
				}()
				select {
				case <-recvDone:
				case <-sendDone:
				}
			}()
		}
	}()

	go func() {
		<-ctx.Done()
		l.Close()
	}()

	return nil
}

const exportTimeout = 3 * time.Minute

func (s *Service) startExporter(ctx context.Context) (map[string]*tor.OnionForward, error) {
	// Wait at most a few minutes to publish the service
	exportCtx, cancel := context.WithTimeout(ctx, exportTimeout)
	defer cancel()

	// Build a port map for remote onion forwards, per service alias
	serviceFwds := map[string]map[string][]int{}
	serviceKeys := map[string][]byte{}
	for _, export := range s.exports {
		srcAddr, err := export.Source().SingleAddr()
		if err != nil {
			return nil, err
		}
		if export.Source().IsUnix() {
			srcAddr = "unix:" + srcAddr
		}
		exportFwds, ok := serviceFwds[export.Destination().Alias()]
		if !ok {
			exportFwds = map[string][]int{}
			exportFwds[srcAddr] = export.Destination().Ports()
			serviceFwds[export.Destination().Alias()] = exportFwds
			if key := export.Destination().ServiceKey(); len(key) > 0 {
				serviceKeys[export.Destination().Alias()] = key
				// TODO: really should use memguard for this
				// TODO: but what does bine and Tor do about keys in memory?
				// TODO: and what about all the copies of the keys on disk?
				defer zeroize(key)
			}
		}
		exportFwds[srcAddr] = export.Destination().Ports()
	}

	// Forward onion services
	fwds := map[string]*tor.OnionForward{}
	for alias, exportFwds := range serviceFwds {
		var key interface{}
		if aliasKey, ok := serviceKeys[alias]; ok {
			key = ed25519.PrivateKey(aliasKey)
		} else {
			key = nil
		}
		fwd, err := s.tor.Forward(exportCtx, &tor.ForwardConf{
			PortForwards: exportFwds,
			Key:          key,
			Version3:     true,
			NonAnonymous: s.nonAnonymous,
		})
		if err != nil {
			return nil, fmt.Errorf("Failed to create onion forward: %v", err)
		}
		fwds[alias] = fwd
	}

	// Shut down forward w/context
	go func() {
		<-ctx.Done()
		for _, fwd := range fwds {
			fwd.Close()
		}
		close(s.done)
	}()

	return fwds, nil
}

var zeroKey ed25519.PrivateKey

func zeroize(b []byte) {
	copy(b, zeroKey)
}
