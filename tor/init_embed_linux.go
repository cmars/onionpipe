// +build embed

package tor

/*
#cgo CFLAGS: -Iinclude
#cgo LDFLAGS: -Llib -ltor -levent -lssl -lcrypto -lz -lm -ldl -static

#include <stdlib.h>
#include <tor_api.h>

static char** makeCharArray(int size) {
        return calloc(sizeof(char*), size);
}
static void setArrayString(char **a, char *s, int n) {
        a[n] = s;
}
static void freeCharArray(char **a, int size) {
        int i;
        for (i = 0; i < size; i++)
                free(a[i]);
        free(a);
}
*/
import "C"
import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/cretz/bine/process"
)

// init uses libtor directly in the same process, rather than by controlling a
// tor executable. This configuration allows onionpipe to run as a single executable
// with no unpacking required, but has some quirks:
//
// libtor is not as up-to-date with the latest Tor release. It may be missing
// security and bug fixes as a result. Use with caution. There are also some quirks:
// 1. Signal handling seems to be a little strange. Have to send two in order
// for signal.Notify to see it.
// 2. Shutdown sometimes seems to hang. Could be a race condition?
//
func init() {
	processOption = func(c *StartConf) {
		creator := &embeddedCreator{}
		c.ProcessCreator = creator
	}
}

// ProviderVersion returns the Tor provider name and version exposed from the
// Tor embedded API.
func ProviderVersion() string {
	return C.GoString(C.tor_api_get_provider_version())
}

// Creator implements the bine.process.Creator, permitting libtor to act as an API
// backend for the bine/tor Go interface.
var Creator process.Creator = new(embeddedCreator)

// embeddedCreator implements process.Creator, permitting libtor to act as an API
// backend for the bine/tor Go interface.
type embeddedCreator struct{}

// New implements process.Creator, creating a new embedded tor process.
func (embeddedCreator) New(ctx context.Context, args ...string) (process.Process, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return &embeddedProcess{
		ctx:  ctx,
		conf: C.tor_main_configuration_new(),
		args: args,
	}, nil
}

// embeddedProcess implements process.Process, permitting libtor to act as an API
// backend for the bine/tor Go interface.
type embeddedProcess struct {
	ctx  context.Context
	conf *C.struct_tor_main_configuration_t
	args []string
	done chan int
}

// Start implements process.Process, starting up the libtor embedded process.
func (e *embeddedProcess) Start() error {
	if e.done != nil {
		return errors.New("already started")
	}
	// Create the char array for the args
	args := append([]string{"tor"}, e.args...)

	charArray := C.makeCharArray(C.int(len(args)))
	for i, a := range args {
		C.setArrayString(charArray, C.CString(a), C.int(i))
	}
	// Build the tor configuration
	if code := C.tor_main_configuration_set_command_line(e.conf, C.int(len(args)), charArray); code != 0 {
		C.tor_main_configuration_free(e.conf)
		C.freeCharArray(charArray, C.int(len(args)))
		return fmt.Errorf("failed to set arguments: %v", int(code))
	}
	// Start tor and return
	e.done = make(chan int, 1)
	go func() {
		defer C.freeCharArray(charArray, C.int(len(args)))
		defer C.tor_main_configuration_free(e.conf)
		e.done <- int(C.tor_run_main(e.conf))
	}()
	return nil
}

// Wait implements process.Process, blocking until the embedded process terminates.
func (e *embeddedProcess) Wait() error {
	if e.done == nil {
		return errors.New("not started")
	}
	select {
	case <-e.ctx.Done():
		return e.ctx.Err()

	case code := <-e.done:
		if code == 0 {
			return nil
		}
		return fmt.Errorf("embedded tor failed: %v", code)
	}
}

// EmbeddedControlConn implements process.Process, connecting to the control port
// of the embedded Tor isntance.
func (e *embeddedProcess) EmbeddedControlConn() (net.Conn, error) {
	file := os.NewFile(uintptr(C.tor_main_configuration_setup_control_socket(e.conf)), "")
	conn, err := net.FileConn(file)
	if err != nil {
		return nil, fmt.Errorf("unable to create control socket: %v", err)
	}
	return conn, nil
}
