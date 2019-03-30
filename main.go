package main

import (
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
	"gitlab.snapp.ir/golangify/tiles/server"
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
	for hash := range cfg.Tiles {
		opts := cfg.Tiles[hash]
		s.AddNode(hash, &opts)
	}

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
