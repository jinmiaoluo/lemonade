package client

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"os"

	log "github.com/inconshreveable/log15"

	"github.com/lemonade-command/lemonade/lemon"
	"github.com/lemonade-command/lemonade/param"
	"github.com/lemonade-command/lemonade/server"
)

// client struct
type client struct {
	host               string
	port               int
	lineEnding         string
	noFallbackMessages bool
	logger             log.Logger
}

// parse CLI struct instance and logger struct instance comes from log15 package
// return a client struct instance
// use the value comes from  CLI struct as client struct's value
func New(c *lemon.CLI, logger log.Logger) *client {
	return &client{
		host:               c.Host,
		port:               c.Port,
		lineEnding:         c.LineEnding,
		noFallbackMessages: c.NoFallbackMessages,
		logger:             logger,
	}
}

var dummy = &struct{}{}

// if fname is a file name
func fileExists(fname string) bool {
	_, err := os.Stat(fname)
	return err == nil
}

// parse file name
// return server address, channel receiver, and error
func serveFile(fname string) (string, <-chan struct{}, error) {
	// create a TCP listener
	// this listener is announces with two basic info
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", nil, err
	}
	finished := make(chan struct{})

	go func() {
		// send data with http
		http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// convert file to []byte
			b, err := ioutil.ReadFile(fname)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(b)

			w.(http.Flusher).Flush()
			//
			finished <- struct{}{}
		}))
	}()

	return fmt.Sprintf("http://127.0.0.1:%d/%s", l.Addr().(*net.TCPAddr).Port, fname), finished, nil
}

//
func (c *client) Open(uri string, transLocalfile, transLoopback bool) error {
	// create a variable to due with channel receive operation
	// receive a struct from send operation
	var finished <-chan struct{}

	// if transLocalfile is true and CLI.DataSource is file
	// serve file
	if transLocalfile && fileExists(uri) {
		var err error
		uri, finished, err = serveFile(uri)
		if err != nil {
			return err
		}
	}

	// if CLI.DataSource not file
	c.logger.Info("Opening " + uri)
	err := c.withRPCClient(func(rc *rpc.Client) error {
		p := &param.OpenParam{
			URI:           uri,
			TransLoopback: transLoopback || transLocalfile,
		}

		return rc.Call("URI.Open", p, dummy)
	})
	if err != nil {
		return err
	}

	if finished != nil {
		<-finished
	}
	return nil
}

func (c *client) Paste() (string, error) {
	var resp string

	err := c.withRPCClient(func(rc *rpc.Client) error {
		return rc.Call("Clipboard.Paste", dummy, &resp)
	})
	if err != nil {
		return "", err
	}

	return lemon.ConvertLineEnding(resp, c.lineEnding), nil
}

func (c *client) Copy(text string) error {
	c.logger.Debug("Sending: " + text)
	return c.withRPCClient(func(rc *rpc.Client) error {
		return rc.Call("Clipboard.Copy", text, dummy)
	})
}

func (c *client) withRPCClient(f func(*rpc.Client) error) error {
	rc, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port))
	if err != nil {
		if !c.noFallbackMessages {
			c.logger.Error(err.Error())
			c.logger.Error("Falling back to localhost")
		}
		rc, err = c.fallbackLocal()
		if err != nil {
			return err
		}
	}

	err = f(rc)
	if err != nil {
		return err
	}
	return nil
}

//
func (c *client) fallbackLocal() (*rpc.Client, error) {
	port, err := server.ServeLocal(c.logger)
	server.LineEndingOpt = c.lineEnding
	if err != nil {
		return nil, err
	}
	return rpc.Dial("tcp", fmt.Sprintf("localhost:%d", port))
}
