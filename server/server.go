package server

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/tidwall/evio"
)

// Server represents tiles (tile-shard) server.
type Server struct {
	Threads int
	Host    string
	Port    int

	Config struct {
		KeepAlive time.Duration
	}
}

// New creates new server instance
func New(threads int, host string, port int) *Server {
	return &Server{
		Threads: threads,
		Host:    host,
		Port:    port,
	}
}

// Run runs the server with event loop
func (s *Server) Run() error {
	return s.evioServe()
}

func (s *Server) evioServe() error {
	var events evio.Events
	if s.Threads == 0 {
		events.NumLoops = -1
	} else {
		events.NumLoops = s.Threads
	}
	events.LoadBalance = evio.LeastConnections

	// fires when the server is ready to accept new connections
	events.Serving = func(eserver evio.Server) (action evio.Action) {
		if eserver.NumLoops == 1 {
			log.Infof("Running single-threaded")
		} else {
			log.Infof("Running on %d threads", eserver.NumLoops)
		}
		for _, addr := range eserver.Addrs {
			log.Infof("Ready to accept connections at %s",
				addr)
		}
		return
	}

	// fires when a connection has closed.
	events.Closed = func(econn evio.Conn, err error) (action evio.Action) {
		// load the client
		client := econn.Context().(*Client)

		log.Debugf("Closed connection: %s", client.remoteAddr)
		return
	}

	// fires when a connection has opened
	events.Opened = func(econn evio.Conn) (out []byte, opts evio.Options, action evio.Action) {
		client := new(Client)
		client.opened = time.Now()
		client.remoteAddr = econn.RemoteAddr().String()

		// keep track of the client
		econn.SetContext(client)

		// set the client keep-alive, if needed
		if s.Config.KeepAlive > 0 {
			opts.TCPKeepAlive = time.Duration(s.Config.KeepAlive)
		}

		log.Debugf("Opened connection: %s", client.remoteAddr)

		return
	}

	return evio.Serve(events, fmt.Sprintf("%s:%d", s.Host, s.Port))
}
