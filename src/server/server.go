package server

import (
	"fmt"
	"net"
	"strings"
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

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}

		timestamp := time.Now().Format(time.RFC3339)
		metrics := string(buf[:n])

		// Remove the "Attaching" line if it exists
		if strings.Contains(metrics, "Attaching") {
			lines := strings.Split(metrics, "\n")
			var filteredMetrics []string
			for _, line := range lines {
				if !strings.Contains(line, "Attaching") {
					filteredMetrics = append(filteredMetrics, line)
				}
			}
			metrics = strings.Join(filteredMetrics, "\n")
		}

		// Clear the previous output using ANSI escape codes
		fmt.Print("\033[H\033[2J") // Clear terminal

		// Print the last update timestamp and metrics
		fmt.Printf("Last update: %s\n", timestamp)
		fmt.Println(metrics)
	}
}
