package server

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	discoveryPort         = 9999
	discoveryErrorMessage = "discovery server is already running on port"
)

func StartServer(serverPort int) error {
	clients := make(map[string]*net.UDPAddr)
	metricsMap := make(map[string]string)

	if err := startDiscoveryServer(clients); err != nil {
		if !strings.Contains(err.Error(), discoveryErrorMessage) {
			return fmt.Errorf("error starting discovery server: %s", err)
		}
	}

	addr := net.UDPAddr{
		Port: serverPort,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return fmt.Errorf("error starting server: %s", err)
	}
	defer conn.Close()

	buf := make([]byte, 2048)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}

		metrics := string(buf[:n])
		clients[remoteAddr.String()] = remoteAddr
		metricsMap[remoteAddr.String()] = metrics

		fmt.Print("\033[H\033[2J")
		fmt.Printf("Last update: %s\n", time.Now().Format(time.RFC3339))
		fmt.Println("Metrics from all machines:")

		for addrStr, metricsData := range metricsMap {
			fmt.Printf("Metrics from %s: %s\n", addrStr, metricsData)
		}

		for addrStr, addrPtr := range clients {
			if addrStr != remoteAddr.String() {
				_, err := conn.WriteToUDP([]byte(metrics), addrPtr)
				if err != nil {
					fmt.Println("Error sending data to client:", err)
				}
			}
		}
	}
}

func startDiscoveryServer(clients map[string]*net.UDPAddr) error {
	addr := net.UDPAddr{
		Port: discoveryPort,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return fmt.Errorf("error starting discovery server: %s", err)
	}
	defer conn.Close()

	fmt.Printf("Discovery server is running on port %d\n", discoveryPort)

	go func() {
		for {
			buf := make([]byte, 2048)
			n, remoteAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println("Error receiving data:", err)
				continue
			}

			clientInfo := string(buf[:n])
			fmt.Printf("Received registration from: %s, Info: %s\n", remoteAddr.String(), clientInfo)
			clients[remoteAddr.String()] = remoteAddr
		}
	}()

	return nil
}
