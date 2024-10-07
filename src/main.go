package main

import (
	"flag"
	"fmt"
	"github.com/gucardona/ga-redes-udp/src/client"
	"github.com/gucardona/ga-redes-udp/src/server"
	"github.com/gucardona/ga-redes-udp/src/vars"
	"log"
	"math/rand"
	"net"
	"time"
)

const (
	messageInterval = 5 * time.Second
)

func main() {
	serverPortRangeMin := 1001
	serverPortRangeMax := 9999
	clientPortRangeMin := 1001
	clientPortRangeMax := 9999

	randomServerPort := rand.Intn(serverPortRangeMax-serverPortRangeMin+1) + serverPortRangeMin
	randomClientPort := rand.Intn(clientPortRangeMax-clientPortRangeMin+1) + clientPortRangeMin

	flag.IntVar(&vars.ServerPort, "server-port", randomServerPort, "UDP server port")
	flag.IntVar(&vars.ClientPort, "client-port", randomClientPort, "UDP client port")

	flag.Parse()

	go func() {
		if err := server.StartDiscoveryServer(); err != nil {
			log.Fatalf("Failed to start discovery server: %s", err)
		}
	}()

	go func() {
		if err := server.StartServer(vars.ServerPort); err != nil {
			log.Fatalf("Failed to start server: %s", err)
		}
	}()

	waitForServer(9999)

	if err := client.StartClient(vars.ServerPort, vars.ClientPort, messageInterval); err != nil {
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
		time.Sleep(1 * time.Second)
	}
}
