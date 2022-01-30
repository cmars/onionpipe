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

	"github.com/cmars/oniongrok/config"
)

// Service implements a forwarding service. It sets up and operates forwarding
// as defined by configuration.
type Service struct {
	tor     *tor.Tor
	imports []*config.Forward
	exports []*config.Forward
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
	}
}

// Start starts forwarding.
func (s *Service) Start(ctx context.Context) (string, error) {
	// Start import forwarding
	for _, importFwd := range s.imports {
		err := s.startImporter(ctx, s.tor, importFwd)
		if err != nil {
			return "", err
		}
	}
	// Start export forwarding
	if len(s.exports) > 0 {
		onionFwd, err := s.startExporter(ctx)
		if err != nil {
			return "", err
		}
		return onionFwd.ID, nil
	}
	return "", nil
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

func (s *Service) startExporter(ctx context.Context) (*tor.OnionForward, error) {
	// Wait at most a few minutes to publish the service
	exportCtx, cancel := context.WithTimeout(ctx, exportTimeout)
	defer cancel()

	// Build a port map for remote onion forwards
	exportForwards := map[string][]int{}
	for _, export := range s.exports {
		srcAddr, err := export.Source().SingleAddr()
		if err != nil {
			return nil, err
		}
		exportForwards[srcAddr] = export.Destination().Ports()
	}

	// Forward onion services
	fwd, err := s.tor.Forward(exportCtx, &tor.ForwardConf{
		PortForwards: exportForwards,
		Version3:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create onion forward: %v", err)
	}

	// Shut down forward w/context
	go func() {
		<-ctx.Done()
		fwd.Close()
	}()

	return fwd, nil
}
