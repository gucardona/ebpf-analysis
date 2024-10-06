package main

import (
	"github.com/gucardona/ga-redes-udp/src/client"
	"github.com/gucardona/ga-redes-udp/src/server"
	"log"
	"time"
)

const (
	serverPort      = 8443
	messageInterval = 5 * time.Second
)

func main() {
	go func() {
		if err := server.StartServer(serverPort); err != nil {
			log.Fatalf("Failed to start server: %s", err)
		}
	}()

	if err := client.StartClient(serverPort, messageInterval); err != nil {
		log.Fatalf("Failed to start client: %s", err)
	}
}
