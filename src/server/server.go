package server

import (
	"fmt"
	"net"
	"time"
)

var Clients []int

func StartServer(serverPort int) error {
	Clients = append(Clients, serverPort)
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
		fmt.Printf("Last update: %s\n", time.Now().Format(time.RFC3339))
		fmt.Println("Metrics from all machines:")

		for addrStr, metricsData := range metricsMap {
			fmt.Printf("Metrics from %s:", addrStr)
			fmt.Println(metricsData)
		}

		fmt.Println("Clients:", Clients)
		for _, port := range Clients {
			if port != serverPort {
				fmt.Println(port)
				_, err := conn.WriteToUDP([]byte(metrics), &net.UDPAddr{
					Port: port,
					IP:   net.ParseIP("127.0.0.1"),
				})
				if err != nil {
					fmt.Printf("Error sending data to client %d: %s\n", port, err)
				}
			}
		}
	}
}
