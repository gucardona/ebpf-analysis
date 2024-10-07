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

var Clients []int

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
			serverPort, ok := strings.CutPrefix(message, "register-")
			if !ok {
				fmt.Println("Prefix not found to cut: ", err)
				continue
			}

			port, err := strconv.Atoi(serverPort)
			if err != nil {
				fmt.Println("Error converting port: ", err)
				continue
			}

			if !ArrayContains(Clients, port) {
				Clients = append(Clients, port)
				fmt.Printf("New client registered: %s\n", remoteAddr.String())
			}

			for _, clientPort := range Clients {
				if clientPort != port {
					_, err := conn.WriteToUDP([]byte(fmt.Sprintf("new-client-%d", port)), &net.UDPAddr{
						Port: clientPort,
						IP:   remoteAddr.IP,
					})
					if err != nil {
						fmt.Println("Error sending discovery message to client:", err)
					}
				}
			}

			for _, clientPort := range Clients {
				_, err := conn.WriteToUDP([]byte(fmt.Sprintf("client-list-%d", clientPort)), remoteAddr)
				if err != nil {
					fmt.Println("Error sending discovery message to the new client:", err)
				}
			}
		}
	}
}
