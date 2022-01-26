package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
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

	exportForwards := map[string][]int{}
	for _, exportMap := range exports {
		exportForwards[fmt.Sprintf("%s:%d", exportMap.LocalAddr, exportMap.LocalPort)] = exportMap.RemotePorts
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

	importCtx, cancel := context.WithCancel(ctx.Context)
	defer cancel()

	for _, importMap := range imports {
		l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", importMap.LocalAddr, importMap.LocalPort))
		if err != nil {
			return fmt.Errorf("failed to listen on local address %s:%d", importMap.LocalAddr, importMap.LocalPort)
		}

		remoteDialer, err := t.Dialer(importCtx, nil)
		if err != nil {
			return fmt.Errorf("failed to create tor network dialer")
		}

		go func() {
			for {
				localConn, err := l.Accept()
				if err != nil {
					if !errors.Is(err, net.ErrClosed) {
						log.Printf("failed to accept on local address %s:%d", importMap.LocalAddr, importMap.LocalPort)
					}
					return
				}
				go func() {
					defer localConn.Close()
					remoteConn, err := remoteDialer.DialContext(importCtx, "tcp", fmt.Sprintf("%s:%d", importMap.RemoteAddr, importMap.RemotePort))
					if err != nil {
						log.Printf("failed to connect to onion address %s:%d", importMap.RemoteAddr, importMap.RemotePort)
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
			<-importCtx.Done()
			l.Close()
		}()

		fmt.Printf("%s:%d => %s:%d", importMap.RemoteAddr, importMap.RemotePort, importMap.LocalAddr, importMap.LocalPort)
	}

	// Wait at most a few minutes to publish the service
	publishCtx, cancel := context.WithTimeout(ctx.Context, 3*time.Minute)
	defer cancel()

	// Create an onion service to listen on any port but show as 80
	fwd, err := t.Forward(publishCtx, &tor.ForwardConf{
		PortForwards: tor.PortForwardMap(exportForwards),
		Version3:     true,
	})
	if err != nil {
		return fmt.Errorf("Failed to create onion service: %v", err)
	}
	defer fwd.Close()

	for localAddr, remotePorts := range exportForwards {
		for _, remotePort := range remotePorts {
			fmt.Printf("%s => %v.onion:%d", localAddr, fwd.ID, remotePort)
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
