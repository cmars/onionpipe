package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/urfave/cli/v2"

	"github.com/cmars/oniongrok/config"
	"github.com/cmars/oniongrok/forwarding"
	"github.com/cmars/oniongrok/secrets"
	"github.com/cmars/oniongrok/tor"
)

var startTor = func(ctx context.Context, options ...tor.Option) (*tor.Tor, error) {
	return tor.Start(ctx, options...)
}

var newForwardingService = func(t *tor.Tor, fwds ...*config.Forward) forwardingService {
	return forwarding.New(t, fwds...)
}

type forwardingService interface {
	Done() <-chan struct{}
	Start(ctx context.Context, options ...forwarding.Option) (map[string]string, error)
}

// Forward sets up and operates oniongrok forwards.
func Forward(ctx *cli.Context) (cmdErr error) {
	var fwds []*config.Forward
	var sec *secrets.Secrets
	for i := 0; i < ctx.Args().Len(); i++ {
		fwd, err := config.ParseForward(ctx.Args().Get(i))
		if err != nil {
			return err
		}
		if fwd.Destination().Alias() != "" {
			if sec == nil {
				sec, err = openSecrets(ctx)
				if err != nil {
					return err
				}
			}
			privkey, err := sec.EnsureServiceKey(fwd.Destination().Alias())
			if err != nil {
				return err
			}
			fwd.Destination().SetServiceKey(privkey)
		}
		fwds = append(fwds, fwd)
	}
	// If we added any service keys, persist them now.
	if sec != nil {
		if err := sec.WriteFile(); err != nil {
			return err
		}
	}

	fwdCtx, cancel := signal.NotifyContext(ctx.Context, os.Interrupt)
	defer cancel()

	var torOptions []tor.Option
	var fwdOptions []forwarding.Option
	if ctx.Bool("debug") {
		torOptions = append(torOptions, tor.Debug(os.Stderr))
	}
	if !ctx.Bool("anonymous") {
		torOptions = append(torOptions, tor.NonAnonymous)
		fwdOptions = append(fwdOptions, forwarding.NonAnonymous)
	}

	var stopped bool
	log.Println("starting tor...")
	t, err := startTor(nil, torOptions...)
	if err != nil {
		return fmt.Errorf("failed to start tor: %v", err)
	}
	svc := newForwardingService(t, fwds...)
	defer func() {
		<-svc.Done()
		if !stopped {
			if err := t.Close(); err != nil {
				log.Println(err)
			}
		}
	}()

	onionIDs, err := svc.Start(fwdCtx, fwdOptions...)
	if err != nil {
		return err
	}

	for _, fwd := range fwds {
		fmt.Println(fwd.Description(onionIDs))
	}

	fmt.Println()
	fmt.Println("press Ctrl-C to exit")
	select {
	case <-svc.Done():
		log.Println("shutting down tor...")
		if err := t.Close(); err != nil {
			log.Println(err)
		}
		stopped = true
	}
	log.Println("shutdown complete")
	return nil
}
