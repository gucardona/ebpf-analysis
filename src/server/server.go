package server

import (
	"fmt"
	"net"
	"time"
)

func StartServer(serverPort int) error {
	addr := net.UDPAddr{
		Port: serverPort,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return fmt.Errorf("error starting server: %s", err)
	}
	defer conn.Close()

	buf := make([]byte, 2048) // Increase buffer size to accommodate larger messages

	var metricsList []string // Store received metrics

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}

		timestamp := time.Now().Format(time.RFC3339)
		metrics := string(buf[:n])

		// Append received metrics to the list
		metricsList = append(metricsList, metrics)

		// Clear the previous output using ANSI escape codes
		fmt.Print("\033[H\033[2J") // Clear terminal

		// Print the last update timestamp and metrics
		fmt.Printf("Last update: %s\n", timestamp)
		fmt.Println("Metrics:")
		fmt.Println(metrics)
	}
}
