package server

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	serverRegisteredClients []int
	clientMessages          = make(map[string]string)
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

	buf := make([]byte, 2048)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}

		message := string(buf[:n])
		clientKey := remoteAddr.String()

		if strings.Contains(message, "new-client-") {
			port, ok := strings.CutPrefix(message, "new-client-")
			if !ok {
				fmt.Println("Prefix not found to cut:", err)
				continue
			}
			portCnv, err := strconv.Atoi(port)
			if err != nil {
				fmt.Println("Error converting port:", err)
				continue
			}
			if !ArrayContains(serverRegisteredClients, portCnv) {
				serverRegisteredClients = append(serverRegisteredClients, portCnv)
				fmt.Printf("New client registered: %d\n", portCnv)
			}
			continue
		}

		if strings.Contains(message, "@") {
			if previousMsg, exists := clientMessages[clientKey]; exists && previousMsg == message {
				continue
			}

			clientMessages[clientKey] = message

			fmt.Println(strings.Repeat("=", 40))
			fmt.Println()

			fmt.Print("\033[H\033[2J")
			fmt.Printf("Last update: %s\n\n", time.Now().Format(time.RFC3339))

			fmt.Println(clientMessages)

			fmt.Println()
			fmt.Println(strings.Repeat("=", 40))

			for _, clientPort := range serverRegisteredClients {
				if clientPort != serverPort && remoteAddr.Port != clientPort {
					fmt.Printf("Forwarding to client: %d\n", clientPort)
					_, err := conn.WriteToUDP([]byte(message), &net.UDPAddr{
						Port: clientPort,
						IP:   net.ParseIP("127.0.0.1"),
					})
					if err != nil {
						fmt.Printf("Error sending data to client %d: %s\n", clientPort, err)
					}
				}
			}
		}
	}
}

func ArrayContains(slice []int, item int) bool {
	for _, element := range slice {
		if element == item {
			return true
		}
	}
	return false
}
