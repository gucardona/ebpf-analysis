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
		return fmt.Errorf("error starting discovery server: %s", err)
	}
	defer conn.Close()

	clients := make(map[string]*net.UDPAddr)
	fmt.Println("clients: ", clients)
	buf := make([]byte, 2048)

	n, remoteAddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("Error receiving discovery message:", err)
	}

	message := string(buf[:n])
	fmt.Println(message)
	if strings.Contains(message, "register") {
		clients[remoteAddr.String()] = remoteAddr
		fmt.Printf("New client registered: %s\n", remoteAddr.String())

		for _, clientAddr := range clients {
			_, err := conn.WriteToUDP([]byte(remoteAddr.String()), clientAddr)
			if err != nil {
				fmt.Println("Error sending discovery message to client:", err)
			}
		}
	}
	return nil
}
