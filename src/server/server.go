package server

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var Clients []int

func StartServer(serverPort int) error {
	metricsMap := make(map[string][]string)

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

		metricsMap[remoteAddr.String()] = append(metricsMap[remoteAddr.String()], metrics)

		fmt.Print("\033[H\033[2J")
		fmt.Printf("Last update: %s", time.Now().Format(time.RFC3339))
		fmt.Println("Metrics from all machines:")

		for addrStr, metricsData := range metricsMap {
			fmt.Printf("Metrics from %s:\n", addrStr)
			formatAndPrintMetrics(metricsData)
		}

		fmt.Println("Clients:", Clients)
		for i := 0; i < len(Clients); i++ {
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

func formatAndPrintMetrics(metricsData []string) {
	fmt.Printf("%-30s %s\n", "Metric", "Count")
	fmt.Println(strings.Repeat("-", 40))

	metricsCount := make(map[string]int)

	for _, metric := range metricsData {
		parts := strings.Split(metric, "]:")
		if len(parts) == 2 {
			metricsCount[parts[0]] += atoi(parts[1])
		} else {
			fmt.Printf("%-30s %s\n", metric, "N/A")
		}
	}

	for metric, count := range metricsCount {
		fmt.Printf("%-30s %d\n", metric, count)
	}
}

func atoi(s string) int {
	if num, err := strconv.Atoi(s); err == nil {
		return num
	}
	return 0
}
