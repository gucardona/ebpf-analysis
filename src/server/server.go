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

	fmt.Println("Server is listening on port:", serverPort)
	fmt.Printf("%-25s %-50s\n", "Timestamp", "Metrics") // Header for the dashboard

	i := 0
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}

		timestamp := time.Now().Format(time.RFC3339)
		metrics := string(buf[:n])

		// Print the received metrics in a structured format
		if i == 1 {
			fmt.Printf("%-25s %-50s\n", "Timestamp", "Metrics") // Header for the dashboard
		}
		fmt.Printf("%-25s %-50s\n", timestamp, metrics)

		i++
	}
}
