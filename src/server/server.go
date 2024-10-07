package server

import (
	"fmt"
	"github.com/gucardona/ga-redes-udp/src/vars"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	serverRegisteredClients []int
	clientMessages          = make(map[string]string)
	currentType             string
	currentTypeMessage      string
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

			fmt.Println(strings.Repeat("=", 150))
			fmt.Println()

			fmt.Print("\033[H\033[2J")
			fmt.Printf("Last update: %s\n\n", time.Now().Format(time.RFC3339))

			displayAllMetrics()

			fmt.Println()
			fmt.Println(strings.Repeat("=", 150))

			for _, registeredServerPort := range serverRegisteredClients {
				if registeredServerPort != serverPort && remoteAddr.Port != registeredServerPort {
					forwardAddr := &net.UDPAddr{
						Port: registeredServerPort,
						IP:   net.ParseIP("127.0.0.1"),
					}

					_, err := conn.WriteToUDP([]byte(message), forwardAddr)
					if err != nil {
						fmt.Printf("Error sending data to client %d: %s\n", registeredServerPort, err)
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
	fmt.Printf("%-30s %s\n", "Metric Type", "Metric Data")
	fmt.Println(strings.Repeat("-", 150))

	for clientKey, message := range clientMessages {
		if clientKey == fmt.Sprintf("127.0.0.1:%d", vars.ClientPort) {
			fmt.Printf("%-30s %s:%s\n", fmt.Sprintf("127.0.0.1%d (this machine)", vars.ClientPort), currentTypeMessage, formatMetricsForClient(message))
		} else {
			fmt.Printf("%-30s %s:%s\n", clientKey, currentTypeMessage, formatMetricsForClient(message))
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
		if strings.HasPrefix(line, ":T:") {
			currentType = strings.TrimSpace(line)
			currentType = strings.Replace(currentType, ":T:", "", 1)
			switch currentType {
			case "SCHEDULE_METRIC":
				currentTypeMessage = "Kernel Schedule Times"
				break
			case "PACKET_METRIC":
				currentTypeMessage = "Kernel Sent/Received Packets"
				break
			case "DATA_METRIC":
				currentTypeMessage = "Kernel Transmitted Data"
				break
			case "RTIME_METRIC":
				currentTypeMessage = "Kernel Runtime Process (ns)"
				break
			case "READ_METRIC":
				currentTypeMessage = "Kernel Read Times"
				break
			case "WRITE_METRIC":
				currentTypeMessage = "Kernel Write Times"
				break
			}
		}

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
