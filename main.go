package main

import (
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
	"gitlab.snapp.ir/golangify/tiles/config"
	"gitlab.snapp.ir/golangify/tiles/server"
)

func main() {
	fmt.Println("18.20 at Sep 07 2016 7:20 IR721")

	if config.GetConfig().Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// creates a server
	s := server.New(config.GetConfig().Threads, config.GetConfig().Host, config.GetConfig().Port)

	// loads server extra configuration
	s.Config.KeepAlive = config.GetConfig().KeepAlive

	// setup the shard!
	for hash, opts := range config.GetConfig().Tiles {
		s.AddNode(hash, &opts)
	}

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
