package main

import (
	"flag"
	"github.com/gucardona/ga-redes-udp/src/client"
	"github.com/gucardona/ga-redes-udp/src/server"
	"log"
	"time"
)

const (
	messageInterval = 5 * time.Second
)

func main() {
	var serverPort int
	flag.IntVar(&serverPort, "port", 8443, "UDP server port")
	flag.Parse()

	go func() {
		if err := server.StartDiscoveryServer(); err != nil {
			log.Fatalf("Failed to start discovery server: %s", err)
		}
	}()

	mainServer := server.NewServer(serverPort)
	go func() {
		if err := mainServer.Start(); err != nil {
			log.Fatalf("Failed to start main server: %s", err)
		}
	}()

	// Wait for servers to start
	time.Sleep(2 * time.Second)

	if err := client.StartClient(serverPort, messageInterval); err != nil {
		log.Fatalf("Failed to start client: %s", err)
	}
}
