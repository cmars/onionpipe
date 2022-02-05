package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"

	"github.com/cmars/oniongrok/config"
	"github.com/cmars/oniongrok/forwarding"
	"github.com/cmars/oniongrok/secrets"
	"github.com/cmars/oniongrok/tor"
)

var forwardFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:  "debug",
		Usage: "enable debug log output",
	},
	&cli.BoolFlag{
		Name:  "anonymous",
		Usage: "publish anonymous hidden services",
		Value: true,
	},
	&cli.PathFlag{
		Name:  "secrets",
		Usage: "path where service and client secrets are stored",
	},
}

func defaultSecretsPath() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Printf("failed to locate home directory: %v", err)
		return ""
	}
	return filepath.Join(home, ".local", "share", "oniongrok", "secrets.json")
}

func App() *cli.App {
	return &cli.App{
		Name:   "oniongrok",
		Usage:  "forward services through Tor; .onion addresses for anything",
		Flags:  forwardFlags,
		Action: Forward,
		Commands: []*cli.Command{{
			Name:    "forward",
			Aliases: []string{"fwd"},
			Usage:   "forward socket address through Tor network",
			Flags:   forwardFlags,
			Action:  Forward,
		}},
	}
}

const startTorTimeout = time.Minute * 3

func secretsPath(path string, anonymous bool) string {
	if !anonymous {
		return strings.TrimSuffix(path, filepath.Ext(path)) + ".not-anonymous" + filepath.Ext(path)
	}
	return path
}

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
				secPath := ctx.Path("secrets")
				if secPath == "" {
					secPath = defaultSecretsPath()
				}
				sec, err = secrets.ReadFile(secretsPath(secPath, ctx.Bool("anonymous")))
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
