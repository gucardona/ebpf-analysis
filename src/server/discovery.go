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

var Clients = make(map[int]net.UDPAddr) // Mapeia a porta do cliente para seu endereço UDP

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

		if strings.HasPrefix(message, "register-") {
			serverPortStr := strings.TrimPrefix(message, "register-")
			port, err := strconv.Atoi(serverPortStr)
			if err != nil {
				fmt.Println("Error converting port:", err)
				continue
			}

			// Adiciona o cliente ao map
			if _, exists := Clients[port]; !exists {
				Clients[port] = *remoteAddr // Mapeia a porta para o endereço remoto
				fmt.Printf("New client registered: %s\n", remoteAddr.String())
			}

			// Notifica todos os clientes registrados sobre o novo cliente
			for clientPort, clientAddr := range Clients {
				if clientPort != port {
					_, err := conn.WriteToUDP([]byte(fmt.Sprintf("new-client-%d", port)), &clientAddr)
					if err != nil {
						fmt.Println("Error sending discovery message to client:", err)
					}
				}
			}

			// Envia a lista de clientes para o novo cliente
			clientListMessage := "client-list:"
			for clientPort := range Clients {
				clientListMessage += fmt.Sprintf("%d,", clientPort)
			}
			clientListMessage = strings.TrimSuffix(clientListMessage, ",") // Remove a última vírgula

			_, err = conn.WriteToUDP([]byte(clientListMessage), remoteAddr)
			if err != nil {
				fmt.Println("Error sending discovery message to the new client:", err)
			}
		}
	}
}
