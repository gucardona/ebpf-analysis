package server

import (
	"fmt"
	"github.com/gucardona/ga-redes-udp/src/client"
	"github.com/gucardona/ga-redes-udp/src/vars"
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

			fmt.Println(strings.Repeat("=", 80))
			fmt.Println()

			fmt.Print("\033[H\033[2J")
			fmt.Printf("Last update: %s\n\n", time.Now().Format(time.RFC3339))

			displayAllMetrics()

			fmt.Println()
			fmt.Println(strings.Repeat("=", 80))

			for _, serverRegisteredPort := range serverRegisteredClients {
				if serverRegisteredPort != serverPort && remoteAddr.Port != serverRegisteredPort {
					serverAddr := net.UDPAddr{
						Port: serverRegisteredPort,
						IP:   net.ParseIP("127.0.0.1"),
					}

					conn, err := net.DialUDP("udp", client.ClientAddr, &serverAddr)
					if err != nil {
						return fmt.Errorf("error connecting to server: %s", err)
					}
					defer conn.Close()

					_, err = conn.Write([]byte(message))
					if err != nil {
						fmt.Printf("Error sending data to client %d: %s\n", serverRegisteredPort, err)
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

func displayAllMetrics() {
	fmt.Printf("%-30s %s\n", "Client", "Metric Data")
	fmt.Println(strings.Repeat("-", 80))

	for clientKey, message := range clientMessages {
		if clientKey == fmt.Sprintf("127.0.0.1:%d", vars.ClientPort) {
			fmt.Printf("%-30s %s\n", clientKey+" (this machine)", formatMetricsForClient(message))
		} else {
			fmt.Printf("%-30s %s\n", clientKey, formatMetricsForClient(message))
		}
		fmt.Println()
	}
}

func formatMetricsForClient(metricsData string) string {
	trim := strings.TrimSpace(metricsData)

	lines := strings.Split(trim, "\n")

	if len(lines) > 1 {
		lines = lines[1:]
	}

	var formattedMetrics strings.Builder
	for _, line := range lines {
		nameIndex := strings.Index(line, "@[")
		quantIndex := strings.Index(line, "]: ")

		if nameIndex != -1 && quantIndex != -1 {
			name := line[nameIndex+2 : quantIndex]
			quant := line[quantIndex+3:]
			formattedMetrics.WriteString(fmt.Sprintf("%s: %s, ", name, quant))
		}
	}
	return strings.TrimSuffix(formattedMetrics.String(), ", ")
}
