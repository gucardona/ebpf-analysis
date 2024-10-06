package server

import (
	"fmt"
	"net"
	"time"
)

var Clients = make(map[string]*net.UDPAddr)

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
		Clients[remoteAddr.String()] = remoteAddr

		fmt.Println(Clients)

		metricsMap[remoteAddr.String()] = metrics

		fmt.Print("\033[H\033[2J")
		fmt.Printf("Last update: %s\n", time.Now().Format(time.RFC3339))
		fmt.Println("Metrics from all machines:")

		for addrStr, metricsData := range metricsMap {
			fmt.Printf("Metrics from %s: %s\n", addrStr, metricsData)
		}

		for addrStr, addrPtr := range Clients {
			if addrStr != remoteAddr.String() {
				_, err := conn.WriteToUDP([]byte(metrics), addrPtr)
				if err != nil {
					fmt.Println("Error sending data to client:", err)
				}
			}
		}
	}
}
