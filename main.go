package main

import (
	"log"

	"github.com/snapp-cab/tiles/config"
	"github.com/snapp-cab/tiles/server"
)

func main() {
	s := server.New(config.GetConfig().Threads, config.GetConfig().Host, config.GetConfig().Port)
	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
