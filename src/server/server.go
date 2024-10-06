package server

import (
	"fmt"
	"net"
	"os"
	"time"
)

const discoveryPort = 9999

func StartServer(serverPort int) error {
	clients := make(map[string]*net.UDPAddr) // Store discovered clients
	metricsMap := make(map[string]string)    // Store metrics from each client

	// Setup UDP listener for the main server
	addr := net.UDPAddr{
		Port: serverPort,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return fmt.Errorf("error starting server: %s", err)
	}
	defer conn.Close()

	// Start discovery server in a separate goroutine
	go startDiscoveryServer(clients)

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

// startDiscoveryServer handles client registration via UDP discovery messages
func startDiscoveryServer(clients map[string]*net.UDPAddr) {
	addr := net.UDPAddr{
		Port: discoveryPort,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("Error starting discovery server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	for {
		buf := make([]byte, 2048)
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}

		clientInfo := string(buf[:n])
		fmt.Printf("Received registration from: %s, Info: %s\n", remoteAddr.String(), clientInfo)

		// Register the client address
		clients[remoteAddr.String()] = remoteAddr
	}
}
