package main

import (
	"fmt"
	"log"

	"github.com/1995parham/tiles/server"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("18.20 at Sep 07 2016 7:20 IR721")

	// loads configuration
	cfg := config()

	// sets log level
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// creates a server
	s := server.New(cfg.Threads, cfg.Host, cfg.Port)

	// loads server extra configuration
	s.Config.KeepAlive = cfg.KeepAlive

	// setup the shards!
	shd := shards()
	for hash := range shd {
		opts := shd[hash]
		s.AddNode(hash, &opts)
	}

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
