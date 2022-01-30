package app

import (
	"fmt"
	"os"
	"os/signal"

	"berty.tech/go-libtor"
	"github.com/cretz/bine/tor"
	"github.com/urfave/cli/v2"

	"github.com/cmars/oniongrok/config"
	"github.com/cmars/oniongrok/forwarding"
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

func Forward(ctx *cli.Context) error {
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

	torConf := &tor.StartConf{ProcessCreator: libtor.Creator}
	if ctx.Bool("debug") {
		torConf.DebugWriter = os.Stderr
	}
	fmt.Println("Starting tor...")
	t, err := tor.Start(fwdCtx, torConf)
	if err != nil {
		return fmt.Errorf("failed to start tor: %v", err)
	}
	defer t.Close()

	svc := forwarding.New(t, fwds...)
	onionID, err := svc.Start(fwdCtx)
	if err != nil {
		return err
	}

	for _, fwd := range fwds {
		fmt.Println(fwd.Description(onionID))
	}

	fmt.Println()

	<-fwdCtx.Done()
	fmt.Println("Interrupt received, shutting down")
	return nil
}
