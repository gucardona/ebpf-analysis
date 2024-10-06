package server

import (
	"fmt"
	"net"
	"strings"
	"time"
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

	buf := make([]byte, 2048) // Increase buffer size to accommodate larger messages

	metricsMap := make(map[string]map[string]int) // map[MachineID]map[MetricName]Count

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}

		// Parse the incoming metrics (MachineID: metrics)
		received := string(buf[:n])
		parts := strings.SplitN(received, ": ", 2)
		if len(parts) != 2 {
			continue // Invalid format
		}
		machineID := parts[0]
		metricsData := parts[1]

		// Parse metrics data into a structured format
		metrics := parseMetrics(metricsData)

		// Update the metrics map
		if _, exists := metricsMap[machineID]; !exists {
			metricsMap[machineID] = make(map[string]int)
		}
		for name, count := range metrics {
			metricsMap[machineID][name] += count
		}

		// Clear the previous output using ANSI escape codes
		fmt.Print("\033[H\033[2J") // Clear terminal

		// Print the last update timestamp and metrics
		fmt.Printf("Last update: %s\n", time.Now().Format(time.RFC3339))
		fmt.Println("Metrics:")
		for id, metrics := range metricsMap {
			fmt.Printf("Machine: %s\n", id)
			for metric, count := range metrics {
				fmt.Printf(" - %s : %d\n", metric, count)
			}
		}
	}
}

// Function to parse metrics data into a map
func parseMetrics(data string) map[string]int {
	metrics := make(map[string]int)
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		metricName := strings.TrimSpace(parts[0])
		count := 0
		fmt.Sscanf(parts[1], "%d", &count)
		metrics[metricName] = count
	}
	return metrics
}
