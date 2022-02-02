package forwarding_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/cmars/oniongrok/config"
	"github.com/cmars/oniongrok/forwarding"
	"github.com/cmars/oniongrok/tor"
	qt "github.com/frankban/quicktest"
	"golang.org/x/net/context/ctxhttp"
)

const skipForwardingTests = "SKIP_FORWARDING_TESTS"

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("hello world"))
	if err != nil {
		panic(err)
	}
})

const testTimeout = 300 * time.Second
const forwardTimeout = 60 * time.Second
const clientTimeout = 60 * time.Second
const closeTimeout = 30 * time.Second

// Tor Browser User Manual
const importTestHost = "dsbqrprgkqqifztta6h3w7i2htjhnq7d3qkh3c7gvc35e66rrcv66did.onion"

func TestIntegration(t *testing.T) {
	c := qt.New(t)
	if os.Getenv(skipForwardingTests) != "" {
		c.Skip()
	}

	// Setup
	wd := c.Mkdir()
	c.Assert(os.Chdir(wd), qt.IsNil)

	// A local server we're going to export
	srv := httptest.NewServer(testHandler)
	c.Cleanup(srv.Close)

	unixSrv := httptest.NewUnstartedServer(testHandler)
	unixListener, err := net.Listen("unix", wd+"/test.sock")
	c.Assert(err, qt.IsNil)
	c.Cleanup(func() { unixListener.Close() })
	unixSrv.Listener = unixListener
	unixSrv.Start()
	c.Cleanup(srv.Close)

	// Find a likely open port for importing a remote server
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	c.Assert(err, qt.IsNil)
	importPort := l.Addr().(*net.TCPAddr).Port
	c.Assert(l.Close(), qt.IsNil)
	c.Assert(importPort, qt.Not(qt.Equals), 0)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	c.Cleanup(cancel)

	torSvc, err := tor.Start(ctx, tor.Debug(os.Stderr))
	c.Assert(err, qt.IsNil)
	c.Cleanup(func() {
		// TODO: improve bine to support a context on Close
		ch := make(chan struct{})
		go func() {
			err := torSvc.Close()
			c.Assert(err, qt.IsNil)
			close(ch)
		}()
		select {
		case <-ch:
		case <-time.After(closeTimeout):
			c.Log("failed to shut down Tor -- possible bug in bine")
		}
	})

	srvURL, err := url.Parse(srv.URL)
	c.Assert(err, qt.IsNil)
	exportHost, exportPort, err := net.SplitHostPort(srvURL.Host)
	c.Assert(err, qt.IsNil)
	exportPortNum, err := strconv.Atoi(exportPort)
	c.Assert(err, qt.IsNil)
	exportDoc := config.ForwardDoc{
		Src: config.EndpointDoc{
			Host:  exportHost,
			Ports: []int{exportPortNum},
		},
		Dest: config.EndpointDoc{
			Ports: []int{80},
		},
	}
	exportFwd, err := exportDoc.Forward()
	c.Assert(err, qt.IsNil)
	exportUnixDoc := config.ForwardDoc{
		Src: config.EndpointDoc{
			Path: wd + "/test.sock",
		},
		Dest: config.EndpointDoc{
			Ports: []int{81},
		},
	}
	exportUnixFwd, err := exportUnixDoc.Forward()
	c.Assert(err, qt.IsNil)

	importDoc := config.ForwardDoc{
		Src: config.EndpointDoc{
			// Tor Browser User Manual
			Host:  importTestHost,
			Ports: []int{80},
		},
		Dest: config.EndpointDoc{
			Ports: []int{importPort},
		},
	}
	importFwd, err := importDoc.Forward()
	c.Assert(err, qt.IsNil)

	fwdSvc := forwarding.New(torSvc, exportFwd, exportUnixFwd, importFwd)

	fwdCtx, cancel := context.WithTimeout(ctx, forwardTimeout)
	c.Cleanup(cancel)
	onionID, err := fwdSvc.Start(fwdCtx)
	c.Assert(err, qt.IsNil)

	// Request the exported, remote onion server
	c.Run("request exported service as onion", func(c *qt.C) {
		clientCtx, cancel := context.WithTimeout(ctx, clientTimeout)
		c.Cleanup(cancel)
		clientDialer, err := torSvc.Dialer(clientCtx, nil)
		c.Assert(err, qt.IsNil)
		client := &http.Client{Transport: &http.Transport{DialContext: clientDialer.DialContext}}

		resp, err := ctxhttp.Get(clientCtx, client, "http://"+onionID+".onion")
		c.Assert(err, qt.IsNil)
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		c.Assert(err, qt.IsNil)

		c.Assert(string(respBody), qt.Equals, "hello world")
	})

	// Request an exported unix socket through remote onion server
	c.Run("request exported unix socket as onion", func(c *qt.C) {
		clientCtx, cancel := context.WithTimeout(ctx, clientTimeout)
		c.Cleanup(cancel)
		clientDialer, err := torSvc.Dialer(clientCtx, nil)
		c.Assert(err, qt.IsNil)
		client := &http.Client{Transport: &http.Transport{DialContext: clientDialer.DialContext}}

		resp, err := ctxhttp.Get(clientCtx, client, "http://"+onionID+".onion:81")
		c.Assert(err, qt.IsNil)
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		c.Assert(err, qt.IsNil)

		c.Assert(string(respBody), qt.Equals, "hello world")
	})

	c.Run("request imported onion as local service", func(c *qt.C) {
		clientCtx, cancel := context.WithTimeout(ctx, clientTimeout)
		c.Cleanup(cancel)
		req, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/index.html", importPort), nil)
		c.Assert(err, qt.IsNil)
		resp, err := ctxhttp.Do(clientCtx, http.DefaultClient, req)
		c.Assert(err, qt.IsNil)
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		c.Assert(err, qt.IsNil)

		c.Assert(string(respBody), qt.Contains, "<HTML>")
		c.Assert(string(respBody), qt.Contains, "Tor Project")
	})
}
