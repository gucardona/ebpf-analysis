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
	serverRegisteredClients []int
	clientMessages          = make(map[string]string)
	mu                      sync.Mutex // Mutex for synchronizing access to client messages
)

// StartServer starts the UDP server on the specified port.
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

		// Register new clients
		if isNewClient(message) {
			port, _ := extractPort(message)
			registerClient(port)
			continue
		}

		// Process and print metrics
		if isMetricMessage(message) {
			clientMessages[clientKey] = message
			printMetrics()
			sendMetricsToClients(conn, message, serverPort)
		}
	}
}

// isNewClient checks if the message indicates a new client.
func isNewClient(message string) bool {
	return strings.Contains(message, "new-client-")
}

// extractPort extracts the client port from the message.
func extractPort(message string) (int, bool) {
	portStr, ok := strings.CutPrefix(message, "new-client-")
	if !ok {
		return 0, false
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Println("Error converting port:", err)
		return 0, false
	}
	return port, true
}

// registerClient registers a new client port.
func registerClient(port int) {
	mu.Lock()
	defer mu.Unlock()

	if !arrayContains(serverRegisteredClients, port) {
		serverRegisteredClients = append(serverRegisteredClients, port)
		fmt.Printf("Client registered: %d\n", port)
	}
}

// isMetricMessage checks if the message contains metric data.
func isMetricMessage(message string) bool {
	return strings.Contains(message, "@")
}

// printMetrics formats and displays the collected metrics.
func printMetrics() {
	fmt.Println(strings.Repeat("=", 150))
	fmt.Println()
	fmt.Printf("Last update: %s\n\n", time.Now().Format(time.RFC3339))
	displayAllMetrics()
	fmt.Println()
	fmt.Println(strings.Repeat("=", 150))
}

// sendMetricsToClients sends the received metrics to all registered clients.
func sendMetricsToClients(conn *net.UDPConn, message string, serverPort int) {
	mu.Lock()
	defer mu.Unlock()

	for _, registeredServerPort := range serverRegisteredClients {
		if registeredServerPort != serverPort {
			forwardAddr := &net.UDPAddr{
				Port: registeredServerPort,
				IP:   net.ParseIP("127.0.0.1"),
			}
			if _, err := conn.WriteToUDP([]byte(message), forwardAddr); err != nil {
				fmt.Printf("Error sending data to client %d: %s\n", registeredServerPort, err)
			}
		}
	}
}

// arrayContains checks if a slice contains a specific integer.
func arrayContains(slice []int, item int) bool {
	for _, element := range slice {
		if element == item {
			return true
		}
	}
	return false
}

// displayAllMetrics displays all collected metrics.
func displayAllMetrics() {
	fmt.Printf("%-30s %s\n", "Metric Type", "Metric Data")
	fmt.Println(strings.Repeat("-", 150))

	for clientKey, message := range clientMessages {
		currentTypeMessage, formattedMessage := formatMetricsForClient(message)
		if clientKey == fmt.Sprintf("127.0.0.1:%d", vars.ClientPort) {
			fmt.Printf("%-30s %s: %s\n", fmt.Sprintf("127.0.0.1:%d (this machine)", vars.ClientPort), currentTypeMessage, formattedMessage)
		} else {
			fmt.Printf("%-30s %s: %s\n", clientKey, currentTypeMessage, formattedMessage)
		}
		fmt.Println()
	}
}

// formatMetricsForClient formats the metrics data for a specific client.
func formatMetricsForClient(metricsData string) (string, string) {
	var currentTypeMessage string
	lines := strings.Split(strings.TrimSpace(metricsData), "\n")

	if len(lines) > 1 {
		lines = lines[1:] // Skip the first line if needed
	}

	var formattedMetrics strings.Builder
	for _, line := range lines {
		if strings.HasPrefix(line, ":T:") {
			currentTypeMessage = getMetricType(line)
		}

		if name, quant := extractNameAndQuantity(line); name != "" && quant != "" {
			formattedMetrics.WriteString(fmt.Sprintf("%s: %s, ", name, quant))
		}
	}

	return currentTypeMessage, strings.TrimSuffix(formattedMetrics.String(), ", ")
}

// getMetricType maps the metric type from the line.
func getMetricType(line string) string {
	currentType := strings.TrimSpace(strings.Replace(line, ":T:", "", 1))
	switch currentType {
	case "SCHEDULE_METRIC":
		return "Kernel Schedule Times"
	case "PACKET_METRIC":
		return "Kernel Sent/Received Packets"
	case "DATA_METRIC":
		return "Kernel Transmitted Data"
	case "RTIME_METRIC":
		return "Kernel Runtime Process (ns)"
	case "READ_METRIC":
		return "Kernel Read Times"
	case "WRITE_METRIC":
		return "Kernel Write Times"
	default:
		return ""
	}
}

// extractNameAndQuantity extracts the name and quantity from a line.
func extractNameAndQuantity(line string) (string, string) {
	nameIndex := strings.Index(line, "@[")
	quantIndex := strings.Index(line, "]: ")
	if nameIndex != -1 && quantIndex != -1 {
		name := line[nameIndex+2 : quantIndex]
		quant := line[quantIndex+3:]
		return name, quant
	}
	return "", ""
}
