package server

import (
	"fmt"
	"net"
	"strings"
)

const (
	DiscoveryPort = 9999
)

func StartDiscoveryServer() error {
	addr := net.UDPAddr{
		Port: DiscoveryPort,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		if strings.Contains(err.Error(), "bind: address already in use") {
			return nil
		}

		return fmt.Errorf("error starting discovery server: %s", err)
	}
	defer conn.Close()

	Clients = make(map[string]*net.UDPAddr)
	fmt.Println("Discovery server started. Listening for messages...")

	buf := make([]byte, 2048)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		fmt.Println(">>> ", string(buf[:n]))
		if err != nil {
			fmt.Println("Error receiving discovery message:", err)
			continue // Go to the next iteration of the loop
		}

		message := string(buf[:n])
		fmt.Println("Received message:", message)
		if strings.Contains(message, "register") {
			Clients[remoteAddr.String()] = remoteAddr
			fmt.Printf("New client registered: %s\n", remoteAddr.String())

			// Notify all clients about the new client
			for _, clientAddr := range Clients {
				_, err := conn.WriteToUDP([]byte(remoteAddr.String()), clientAddr)
				if err != nil {
					fmt.Println("Error sending discovery message to client:", err)
				}
			}
		}
	}
}
