package server

import (
	"fmt"
	"net"
	"strconv"
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

	Clients = []int{}
	fmt.Println("Discovery server started. Listening for messages...")

	buf := make([]byte, 2048)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving discovery message:", err)
			continue
		}

		message := string(buf[:n])
		fmt.Println("Received message:", message)
		if strings.Contains(message, "register-") {
			serverPort, _ := strings.CutPrefix(message, "register-")
			port, _ := strconv.Atoi(serverPort)

			Clients = append(Clients, port)
			fmt.Printf("New client registered: %s\n", remoteAddr.String())

			for _, port := range Clients {
				_, err := conn.WriteToUDP([]byte(remoteAddr.String()), &net.UDPAddr{
					Port: port,
					IP:   net.ParseIP("127.0.0.1"),
				})
				if err != nil {
					fmt.Println("Error sending discovery message to client:", err)
				}
			}
		}
	}
}
