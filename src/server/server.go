package server

import (
	"fmt"
	"net"
	"time"
)

func StartServer(serverPort int) error {
	clients := make(map[string]*net.UDPAddr)
	metricsMap := make(map[string]string) // To store metrics from each client

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

		// Store received metrics in metricsMap
		metricsMap[remoteAddr.String()] = metrics

		// Clear the terminal and print the last update timestamp
		fmt.Print("\033[H\033[2J")
		fmt.Printf("Last update: %s\n", time.Now().Format(time.RFC3339))
		fmt.Println("Metrics from all machines:")

		// Print metrics from all clients
		for addrStr, metricsData := range metricsMap {
			fmt.Printf("Metrics from %s: %s\n", addrStr, metricsData)
		}

		// Broadcast metrics to all clients
		for addrStr, addrPtr := range clients {
			if addrStr != remoteAddr.String() { // Avoid sending metrics back to the sender
				_, err := conn.WriteToUDP([]byte(metrics), addrPtr)
				if err != nil {
					fmt.Println("Error sending data to client:", err)
				}
			}
		}
	}
}
