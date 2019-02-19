package server

import (
	"fmt"

	"github.com/labstack/gommon/log"
	"github.com/tidwall/evio"
)

// Server represents tiles (tile-shard) server.
type Server struct {
	Threads int
	Host    string
	Port    int
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

	return evio.Serve(events, fmt.Sprintf("%s:%d", s.Host, s.Port))
}
