package server

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	Clients      []int
	clientsMutex sync.Mutex
)

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
			portStr := strings.TrimPrefix(message, "new-client-")
			port, err := strconv.Atoi(portStr)
			if err == nil {
				addClient(port)
				fmt.Printf("Registered new client from message: %s\n", remoteAddr.String())
			}
		} else if strings.Contains(message, "client-list-") {
			portStr := strings.TrimPrefix(message, "client-list-")
			port, err := strconv.Atoi(portStr)
			if err == nil {
				addClient(port)
				fmt.Printf("Added client from client-list message: %s\n", remoteAddr.String())
			}
		}

		if strings.Contains(message, "@") {
			metricsMap[remoteAddr.String()] = message
			displayMetrics(metricsMap)

			sendMetricsToClients(conn, message, serverPort)
		}
	}
}

func addClient(port int) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	if !ArrayContains(Clients, port) {
		Clients = append(Clients, port)
	}
}

func displayMetrics(metricsMap map[string]string) {
	fmt.Println(strings.Repeat("=", 40))
	fmt.Printf("Last update: %s\n\n", time.Now().Format(time.RFC3339))

	for _, metricsData := range metricsMap {
		formatAndPrintMetrics(metricsData)
	}

	fmt.Println(strings.Repeat("=", 40))
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

func sendMetricsToClients(conn *net.UDPConn, metrics string, serverPort int) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for _, clientPort := range Clients {
		if clientPort != serverPort {
			go func(port int) {
				_, err := conn.WriteToUDP([]byte(metrics), &net.UDPAddr{
					Port: port,
					IP:   net.ParseIP("127.0.0.1"),
				})
				if err != nil {
					fmt.Printf("Error sending data to client %d: %s\n", port, err)
				}
			}(clientPort)
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
