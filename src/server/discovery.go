package server

import (
	"fmt"
	"github.com/gucardona/ga-redes-udp/src/vars"
	"net"
	"strconv"
	"strings"
)

var discoveryClients []int

func StartDiscoveryServer() error {
	addr := net.UDPAddr{
		Port: 9999,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		if strings.Contains(err.Error(), "bind: address already in use") {
			return nil
		}

		return fmt.Errorf("error starting discovery server: %s", err)
	}
	defer conn.Close()

	buf := make([]byte, 2048)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving discovery message:", err)
			continue
		}

		message := string(buf[:n])

		if strings.Contains(message, "register-") {
			serverPort, ok := strings.CutPrefix(message, "register-")
			if !ok {
				fmt.Println("Prefix not found to cut: ", err)
				continue
			}

			port, err := strconv.Atoi(serverPort)
			if err != nil {
				fmt.Println("Error converting port: ", err)
				continue
			}

			if !ArrayContains(discoveryClients, port) {
				discoveryClients = append(discoveryClients, port)
			}

			for _, clientPort := range discoveryClients {
				if clientPort != port {
					_, err := conn.WriteToUDP([]byte(fmt.Sprintf("new-client-%d", port)), &net.UDPAddr{
						Port: clientPort,
						IP:   remoteAddr.IP,
					})
					if err != nil {
						fmt.Println("Error sending discovery message to client:", err)
					}

					_, err = conn.WriteToUDP([]byte(fmt.Sprintf("new-interval-%d", vars.MessageInterval)), &net.UDPAddr{
						Port: clientPort,
						IP:   remoteAddr.IP,
					})
					if err != nil {
						fmt.Println("Error sending new interval message to client:", err)
					}
				}
			}
		}
	}
}
