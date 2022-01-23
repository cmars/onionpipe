package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/ipsn/go-libtor"
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

type PortMap struct {
	Local  int
	Remote []int
}

func NewPortMap(s string) (*PortMap, error) {
	localRemote := strings.SplitN(s, ":", 2)
	if len(localRemote) == 0 {
		return nil, fmt.Errorf("invalid port mapping %q", s)
	}

	localPort, err := strconv.Atoi(localRemote[0])
	if err != nil {
		return nil, fmt.Errorf("invalid local port %q", s)
	}
	if len(localRemote) == 1 {
		return &PortMap{
			Local:  localPort,
			Remote: []int{localPort},
		}, nil
	}

	portMap := &PortMap{
		Local: localPort,
	}
	remotePortArgs := strings.Split(localRemote[1], ",")
	for _, remotePortArg := range remotePortArgs {
		remotePort, err := strconv.Atoi(remotePortArg)
		if err != nil {
			return nil, fmt.Errorf("invalid remote port %q", remotePortArg)
		}
		portMap.Remote = append(portMap.Remote, remotePort)
	}
	if len(portMap.Remote) == 0 {
		portMap.Remote = []int{portMap.Local}
	}
	return portMap, nil
}

func Forward(ctx *cli.Context) error {
	portMaps := map[int][]int{}
	for i := 0; i < ctx.Args().Len(); i++ {
		portMap, err := NewPortMap(ctx.Args().Get(i))
		if err != nil {
			return err
		}
		portMaps[portMap.Local] = portMap.Remote
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

	// Create an onion service to listen on any port but show as 80
	fwd, err := t.Forward(publishCtx, &tor.ForwardConf{
		LocalAddr: "127.0.0.1",
		PortMap:   portMaps,
		Version3:  true,
	})
	if err != nil {
		return fmt.Errorf("Failed to create onion service: %v", err)
	}
	defer fwd.Close()

	fmt.Println("Forwarding local services:")
	for localPort, remotePorts := range portMaps {
		for _, remotePort := range remotePorts {
			fmt.Printf("%s:%d => %v.onion:%d", fwd.LocalAddr, localPort, fwd.ID, remotePort)
			fmt.Println()
		}
	}

	fmt.Println()
	waitForInterrupt()
	return nil
}

func waitForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
