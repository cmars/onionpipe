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

func App() *cli.App {
	return &cli.App{
		Name:  "oniongrok",
		Usage: "forward services through Tor; .onion addresses for anything",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "enable debug log output",
			},
		},
		Action: Forward,
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

	var options []tor.Option
	if ctx.Bool("debug") {
		options = append(options, tor.Debug(os.Stderr))
	}

	var stopped bool
	log.Println("starting tor...")
	t, err := tor.Start(nil, options...)
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
	onionID, err := svc.Start(fwdCtx)
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
