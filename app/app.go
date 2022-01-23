package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"berty.tech/go-libtor"
	"github.com/cretz/bine/tor"
	"github.com/urfave/cli/v2"
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
	var exports []*ExportMap
	var imports []*ImportMap
	for i := 0; i < ctx.Args().Len(); i++ {
		exportMap, importMap, err := ParsePortMap(ctx.Args().Get(i))
		if err != nil {
			return err
		}
		if importMap != nil {
			imports = append(imports, importMap)
		}
		if exportMap != nil {
			exports = append(exports, exportMap)
		}
	}
	if len(imports) > 0 {
		return fmt.Errorf("import forwarding onions to local not supported yet")
	}

	exportForwards := map[string]map[int][]int{}
	for _, exportMap := range exports {
		portMap, ok := exportForwards[exportMap.LocalAddr]
		if !ok {
			portMap = map[int][]int{}
			exportForwards[exportMap.LocalAddr] = portMap
		}
		portMap[exportMap.LocalPort] = exportMap.RemotePorts
	}

	torConf := &tor.StartConf{ProcessCreator: libtor.Creator}
	if ctx.Bool("debug") {
		torConf.DebugWriter = os.Stderr
	}
	fmt.Println("Starting tor...")
	t, err := tor.Start(nil, torConf)
	if err != nil {
		return fmt.Errorf("failed to start tor: %v", err)
	}
	defer t.Close()

	// Wait at most a few minutes to publish the service
	publishCtx, cancel := context.WithTimeout(ctx.Context, 3*time.Minute)
	defer cancel()

	for localAddr, portMap := range exportForwards {
		// Create an onion service to listen on any port but show as 80
		fwd, err := t.Forward(publishCtx, &tor.ForwardConf{
			LocalAddr: localAddr,
			PortMap:   portMap,
			Version3:  true,
		})
		if err != nil {
			return fmt.Errorf("Failed to create onion service: %v", err)
		}
		defer fwd.Close()

		for localPort, remotePorts := range portMap {
			for _, remotePort := range remotePorts {
				fmt.Printf("%s:%d => %v.onion:%d", localAddr, localPort, fwd.ID, remotePort)
			}
			fmt.Println()
		}
	}

	fmt.Println()
	waitForInterrupt()
	fmt.Println("Interrupt received, shutting down")
	return nil
}

func waitForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
