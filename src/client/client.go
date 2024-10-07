package client

import (
	"fmt"
	"net"
	"time"
)

func StartClient(port int) {
	addr := net.UDPAddr{
		Port: 9999,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		fmt.Println("Error starting client:", err)
		return
	}
	defer conn.Close()

	// Enviar registro inicial
	register(conn, port)

	// Loop para enviar métricas
	for {
		metrics := fmt.Sprintf("CPU@[utilization]: %d", port) // Exemplo de métrica
		_, err := conn.Write([]byte(metrics))
		if err != nil {
			fmt.Println("Error sending metrics:", err)
		}
		time.Sleep(5 * time.Second) // Envia a cada 5 segundos
	}
}

func register(conn *net.UDPConn, port int) {
	_, err := conn.Write([]byte(fmt.Sprintf("register-%d", port)))
	if err != nil {
		fmt.Println("Error sending register message:", err)
	}
}
