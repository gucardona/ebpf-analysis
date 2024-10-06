package main

import (
	"flag"
	"fmt"
	"github.com/gucardona/ga-redes-udp/src/client"
	"github.com/gucardona/ga-redes-udp/src/server"
	"log"
	"net"
	"sync"
	"time"
)

const (
	messageInterval = 5 * time.Second
)

func main() {
	var serverPort int
	flag.IntVar(&serverPort, "port", 8443, "UDP server port")
	flag.Parse()

	var wg sync.WaitGroup
	wg.Add(1)

	if err := server.StartDiscoveryServer(); err != nil {
		log.Fatalf("Failed to start discovery server: %s", err)
	}

	go func() {
		if err := server.StartServer(serverPort); err != nil {
			log.Fatalf("Failed to start server: %s", err)
		}
	}()

	waitForServer(9999)

	if err := client.StartClient(serverPort, messageInterval); err != nil {
		log.Fatalf("Failed to start client: %s", err)
	}
}

func waitForServer(port int) {
	for {
		_, err := net.Dial("udp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			log.Println("Servidor está pronto para aceitar conexões.")
			break
		}
		time.Sleep(1 * time.Second) // Espera um segundo antes de tentar novamente
	}
}
