package app

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/cmars/oniongrok/config"
	"github.com/cmars/oniongrok/forwarding"
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

const startTorTimeout = time.Minute + 3

func Forward(ctx *cli.Context) (cmdErr error) {
	var fwds []*config.Forward
	for i := 0; i < ctx.Args().Len(); i++ {
		fwd, err := config.ParseForward(ctx.Args().Get(i))
		if err != nil {
			return err
		}
		fwds = append(fwds, fwd)
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
	t, err := tor.Start(nil, torOptions...)
	if err != nil {
		return fmt.Errorf("failed to start tor: %v", err)
	}
	defer func() {
		if !stopped {
			if err := t.Close(); err != nil {
				log.Println(err)
			}
		}
	}()

	svc := forwarding.New(t, fwds...)
	onionID, err := svc.Start(fwdCtx, fwdOptions...)
	if err != nil {
		return err
	}

	for _, fwd := range fwds {
		fmt.Println(fwd.Description(onionID))
	}

	fmt.Println()
	fmt.Println("press Ctrl-C to exit")
	select {
	case <-fwdCtx.Done():
		log.Println("shutting down tor...")
		if err := t.Close(); err != nil {
			log.Println(err)
		}
		stopped = true
	}
	log.Println("shutdown complete")
	return nil
}
