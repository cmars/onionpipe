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
	qt "github.com/frankban/quicktest"
	"golang.org/x/net/context/ctxhttp"
)

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("hello world"))
	if err != nil {
		panic(err)
	}
})

const testTimeout = 3 * time.Minute

func TestForwardExportTCP(t *testing.T) {
	c := qt.New(t)

	srv := httptest.NewServer(testHandler)
	c.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	c.Cleanup(cancel)

	tor, err := forwarding.StartTor(ctx, forwarding.TorDebug(os.Stderr))
	c.Assert(err, qt.IsNil)
	c.Cleanup(func() { tor.Close() })

	srvURL, err := url.Parse(srv.URL)
	c.Assert(err, qt.IsNil)
	host, port, err := net.SplitHostPort(srvURL.Host)
	c.Assert(err, qt.IsNil)
	portNum, err := strconv.Atoi(port)
	c.Assert(err, qt.IsNil)
	fwdDoc := config.ForwardDoc{
		Src: config.EndpointDoc{
			Host:  host,
			Ports: []int{portNum},
		},
		Dest: config.EndpointDoc{
			Ports: []int{80},
		},
	}
	fwd, err := fwdDoc.Forward()
	c.Assert(err, qt.IsNil)

	fwdSvc := forwarding.New(tor, fwd)

	fwdCtx, cancel := context.WithTimeout(ctx, time.Minute)
	c.Cleanup(cancel)
	onionID, err := fwdSvc.Start(fwdCtx)
	c.Assert(err, qt.IsNil)

	clientCtx, cancel := context.WithTimeout(ctx, time.Minute)
	c.Cleanup(cancel)
	clientDialer, err := tor.Dialer(clientCtx, nil)
	c.Assert(err, qt.IsNil)
	client := &http.Client{Transport: &http.Transport{DialContext: clientDialer.DialContext}}

	resp, err := ctxhttp.Get(clientCtx, client, "http://"+onionID+".onion")
	c.Assert(err, qt.IsNil)
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, qt.IsNil)

	c.Assert(string(respBody), qt.Equals, "hello world")
}

// Tor Browser User Manual
const importTestHost = "dsbqrprgkqqifztta6h3w7i2htjhnq7d3qkh3c7gvc35e66rrcv66did.onion"

func TestForwardImportTCP(t *testing.T) {
	c := qt.New(t)

	// Find a likely open port
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	c.Assert(err, qt.IsNil)
	port := l.Addr().(*net.TCPAddr).Port
	c.Assert(l.Close(), qt.IsNil)
	c.Assert(port, qt.Not(qt.Equals), 0)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	c.Cleanup(cancel)

	tor, err := forwarding.StartTor(ctx, forwarding.TorDebug(os.Stderr))
	c.Assert(err, qt.IsNil)
	c.Cleanup(func() { tor.Close() })

	fwdDoc := config.ForwardDoc{
		Src: config.EndpointDoc{
			// Tor Browser User Manual
			Host:  importTestHost,
			Ports: []int{80},
		},
		Dest: config.EndpointDoc{
			Ports: []int{port},
		},
	}
	fwd, err := fwdDoc.Forward()
	c.Assert(err, qt.IsNil)

	fwdSvc := forwarding.New(tor, fwd)

	fwdCtx, cancel := context.WithTimeout(ctx, time.Minute)
	c.Cleanup(cancel)
	onionID, err := fwdSvc.Start(fwdCtx)
	c.Assert(err, qt.IsNil)
	// We're not forwarding anything to Tor, so there's no service managed by
	// this instance.
	c.Assert(onionID, qt.Equals, "")

	clientCtx, cancel := context.WithTimeout(ctx, time.Minute)
	c.Cleanup(cancel)
	req, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/index.html", port), nil)
	c.Assert(err, qt.IsNil)
	resp, err := ctxhttp.Do(clientCtx, http.DefaultClient, req)
	c.Assert(err, qt.IsNil)
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, qt.IsNil)

	c.Assert(string(respBody), qt.Contains, "<HTML>")
	c.Assert(string(respBody), qt.Contains, "Tor Project")
}
