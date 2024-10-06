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

	// Channel to signal when the server is ready
	serverReady := make(chan struct{})

	go func() {
		if err := server.StartServer(serverPort); err != nil {
			log.Fatalf("Failed to start server: %s", err)
		}
		// Signal that the server is ready
		close(serverReady)
	}()

	// Wait for the server to be ready
	<-serverReady
	log.Println("Server is ready, starting the client...")

	if err := client.StartClient(serverPort, messageInterval); err != nil {
		log.Fatalf("Failed to start client: %s", err)
	}
}
