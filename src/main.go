package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/gucardona/ga-redes-udp/src/client"
	"github.com/gucardona/ga-redes-udp/src/server"
	"github.com/gucardona/ga-redes-udp/src/vars"
	"log"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	serverPortRangeMin := 1001
	serverPortRangeMax := 9999
	clientPortRangeMin := 1001
	clientPortRangeMax := 9999

	vars.ServerPort = rand.Intn(serverPortRangeMax-serverPortRangeMin+1) + serverPortRangeMin
	vars.ClientPort = rand.Intn(clientPortRangeMax-clientPortRangeMin+1) + clientPortRangeMin

	flag.IntVar(&vars.ServerPort, "server-port", vars.ServerPort, "UDP server port")
	flag.IntVar(&vars.ClientPort, "client-port", vars.ClientPort, "UDP client port")

	flag.Parse()

	go func() {
		reader := bufio.NewReader(os.Stdin)

		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				continue
			}
			input = strings.TrimSpace(input)
			r := regexp.MustCompile(`:\d+`)
			newIntervalStr := r.FindString(input)

			newIntervalStr, found := strings.CutPrefix(newIntervalStr, ":")
			if !found {
				continue
			}

			newIntervalSec, err := strconv.Atoi(newIntervalStr)
			if err != nil {
				continue
			}

			newInterval := time.Duration(newIntervalSec) * time.Second

			vars.MessageInterval = newInterval
		}
	}()

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

	if err := client.StartClient(vars.ServerPort, vars.ClientPort, vars.MessageInterval); err != nil {
		log.Fatalf("Failed to start client: %s", err)
	}
}

func waitForServer(port int) {
	for {
		_, err := net.Dial("udp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
}
