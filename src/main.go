package main

import (
	"flag"
	"github.com/gucardona/ga-redes-udp/src/server"
	"log"
	"sync"
)

func main() {
	var serverPort int
	flag.IntVar(&serverPort, "port", 8443, "UDP server port")
	flag.Parse()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		if err := server.StartServer(serverPort); err != nil {
			log.Fatalf("Failed to start server: %s", err)
		}
	}()
}
