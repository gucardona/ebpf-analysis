package server

import (
	"fmt"
	"net"
	"strings"
	"time"
)

var Clients []int

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

		metrics := string(buf[:n])

		metricsMap[remoteAddr.String()] = metrics

		fmt.Print("\033[H\033[2J")
		fmt.Printf("Last update: %s\n", time.Now().Format(time.RFC3339))
		fmt.Println("Metrics from all machines:")

		for addrStr, metricsData := range metricsMap {
			fmt.Printf("Metrics from %s:\n", addrStr)
			formatAndPrintMetrics(metricsData)
		}

		fmt.Println("Clients:", Clients)
		for i := 0; i < len(Clients); i++ {
			if !arrayContains(Clients, serverPort) {
				Clients = append(Clients, serverPort)
			}
			
			if Clients[i] != serverPort {
				fmt.Println(Clients[i])
				_, err := conn.WriteToUDP([]byte(metrics), &net.UDPAddr{
					Port: Clients[i],
					IP:   net.ParseIP("127.0.0.1"),
				})
				if err != nil {
					fmt.Printf("Error sending data to client %d: %s", Clients[i], err)
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

	lines = lines[1:]

	for _, line := range lines {
		nameIndex := strings.Index(line, "@[")
		quantIndex := strings.Index(line, "]: ")

		name := line[nameIndex+2 : quantIndex]
		quant := line[quantIndex+3:]

		fmt.Printf("%-30s %s\n", name, quant)
	}
}

func arrayContains(slice []int, item int) bool {
	for _, element := range slice {
		if element == item {
			return true
		}
	}
	return false
}
