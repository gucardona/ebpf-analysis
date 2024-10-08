package server

import (
	"fmt"
	"github.com/gucardona/ga-redes-udp/src/vars"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	mu                      sync.Mutex
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

		if strings.HasPrefix(message, "new-client-") {
			handleNewClient(message)
			continue
		}

		if strings.Contains(message, "@") {
			handleClientMessage(conn, remoteAddr, message)
		}
	}
}

func handleNewClient(message string) {
	mu.Lock()
	defer mu.Unlock()

	port, ok := strings.CutPrefix(message, "new-client-")
	if !ok {
		fmt.Println("Prefix not found to cut:", message)
		return
	}

	portCnv, err := strconv.Atoi(port)
	if err != nil {
		fmt.Println("Error converting port:", err)
		return
	}

	if !ArrayContains(serverRegisteredClients, portCnv) {
		serverRegisteredClients = append(serverRegisteredClients, portCnv)
	}
}

func handleClientMessage(conn *net.UDPConn, remoteAddr *net.UDPAddr, message string) {
	mu.Lock()
	defer mu.Unlock()

	clientKey := remoteAddr.String()
	clientMessages[clientKey] = message

	fmt.Print("\033[H\033[2J")
	//fmt.Printf(string([]byte{0x1b, '[', '3', 'J'}))
	fmt.Println(strings.Repeat("=", 150))
	fmt.Printf("Last update: %s\n\n", time.Now().Format(time.RFC3339))
	displayAllMetrics()
	fmt.Println(strings.Repeat("=", 150))

	if !strings.Contains(message, ":RESEND") {
		forwardMessageToClients(conn, message, remoteAddr.Port)
	}
}

func forwardMessageToClients(conn *net.UDPConn, message string, senderPort int) {
	for _, serverRegisteredClientPort := range serverRegisteredClients {
		if serverRegisteredClientPort != vars.ServerPort && serverRegisteredClientPort != vars.ClientPort {
			forwardAddr := &net.UDPAddr{
				Port: serverRegisteredClientPort,
				IP:   net.ParseIP("127.0.0.1"),
			}

			resend := []byte(":RESEND")
			messageBytes := append([]byte(message), resend...)

			_, err := conn.WriteToUDP(messageBytes, forwardAddr)
			if err != nil {
				fmt.Printf("Error sending data to client: %s\n", err)
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
		currentTypeMessage, formattedMessage := formatMetricsForClient(message)
		fmt.Println(currentTypeMessage)
		if clientKey == fmt.Sprintf("127.0.0.1:%d", vars.ClientPort) {
			fmt.Printf("%-30s %s: %s\n", fmt.Sprintf("127.0.0.1%d (this machine)", vars.ClientPort), currentTypeMessage, formattedMessage)
		} else {
			fmt.Printf("%-30s %s: %s\n", clientKey, currentTypeMessage, formattedMessage)
		}
		fmt.Println()
	}
}

func formatMetricsForClient(metricsData string) (string, string) {
	var currentTypeMessage string
	trim := strings.TrimSpace(metricsData)

	lines := strings.Split(trim, "\n")

	if len(lines) > 1 {
		lines = lines[1:]
	}

	var formattedMetrics strings.Builder
	for _, line := range lines {
		if strings.HasPrefix(line, ":T:") {
			currentType := line
			fmt.Println(currentType)
			if strings.Contains(currentType, ":RESEND") {
				currentType = strings.Replace(line, ":RESEND", "", 1)
				fmt.Println(currentType)
			}
			currentType = strings.Replace(line, ":T:", "", 1)
			fmt.Println(currentType)

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
	return currentTypeMessage, strings.TrimSuffix(formattedMetrics.String(), ", ")
}
