package server

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var serverRegisteredClients []int

func StartServer(serverPort int) error {
	metricsMap := make(map[string]string)

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
		if strings.Contains(message, "new-client-") {
			port, ok := strings.CutPrefix(message, "new-client-")
			if !ok {
				fmt.Println("Prefix not found to cut: ", err)
			}
			portCnv, err := strconv.Atoi(port)
			if err != nil {
				fmt.Println("Error converting port: ", err)
				continue
			}
			serverRegisteredClients = append(serverRegisteredClients, portCnv)
		} else {
			if strings.Contains(message, "probes...") {
				metricsMap[remoteAddr.String()] = message
				fmt.Println(strings.Repeat("=", 40))
				fmt.Println()

				fmt.Print("\033[H\033[2J")
				fmt.Printf("Last update: %s\n\n", time.Now().Format(time.RFC3339))

				fmt.Println(metricsMap)
				for _, metricsData := range metricsMap {
					formatAndPrintMetrics(metricsData)
				}

				fmt.Println()
				fmt.Println(strings.Repeat("=", 40))
			}

			if !ArrayContains(serverRegisteredClients, serverPort) {
				serverRegisteredClients = append(serverRegisteredClients, serverPort)
			}

			for _, clientPort := range serverRegisteredClients {
				if clientPort != serverPort {
					fmt.Println("Forwarding to client:", clientPort)
					_, err := conn.WriteToUDP([]byte(message), &net.UDPAddr{
						Port: clientPort,
						IP:   net.ParseIP("127.0.0.1"),
					})
					if err != nil {
						fmt.Printf("Error sending data to client %d: %s", clientPort, err)
					}
				}
			}
		}
	}
}

func formatAndPrintMetrics(metricsData string) {
	fmt.Printf("%-30s %s\n", "Metric", "Count")
	fmt.Println(strings.Repeat("-", 40))

	trim := strings.TrimSpace(metricsData)

	lines := strings.Split(trim, "\n")

	if len(lines) > 1 {
		lines = lines[1:]
	}

	for _, line := range lines {
		nameIndex := strings.Index(line, "@[")
		quantIndex := strings.Index(line, "]: ")

		if nameIndex != -1 && quantIndex != -1 {
			name := line[nameIndex+2 : quantIndex]
			quant := line[quantIndex+3:]

			fmt.Printf("%-30s %s\n", name, quant)
		} else {
			fmt.Printf("Invalid line format: %s\n", line)
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
