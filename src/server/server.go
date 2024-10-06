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

	buf := make([]byte, 1024)
	fmt.Printf("%-20s %-20s\n", "Timestamp", "Metrics")

	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}

		timestamp := time.Now().Format(time.RFC3339)
		metrics := string(buf[:n])

		fmt.Printf("%-20s %-20s\n", timestamp, metrics)
		fmt.Printf("Received from %s: %s\n", clientAddr, string(buf[:n]))
	}
}
