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
	clientsMutex sync.Mutex
)

// Inicia o servidor
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

		// Registrar novos clientes a partir das mensagens de descoberta
		if strings.HasPrefix(message, "new-client-") {
			portStr := strings.TrimPrefix(message, "new-client-")
			port, err := strconv.Atoi(portStr)
			if err == nil {
				addClient(port, remoteAddr) // Armazena o endereço do cliente
				fmt.Printf("Registered new client: %s\n", remoteAddr.String())
			}
		} else if strings.HasPrefix(message, "client-list:") {
			portStrs := strings.TrimPrefix(message, "client-list:")
			portList := strings.Split(portStrs, ",")

			for _, portStr := range portList {
				port, err := strconv.Atoi(strings.TrimSpace(portStr))
				if err == nil {
					addClient(port, remoteAddr) // Armazena o endereço do cliente
					fmt.Printf("Added client from client-list message: %s\n", remoteAddr.String())
				} else {
					fmt.Println("Error converting port:", err)
				}
			}
		}

		// Processar métricas
		if strings.Contains(message, "@") {
			metricsMap[remoteAddr.String()] = message
			displayMetrics(metricsMap)

			// Enviar métricas para outros clientes
			sendMetricsToClients(conn, message, serverPort)
		}
	}
}

// Adiciona um cliente ao mapa de clientes
func addClient(port int, addr net.Addr) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	Clients[port] = *addr.(*net.UDPAddr) // Armazena o endereço do cliente no map
}

// Exibe as métricas recebidas
func displayMetrics(metricsMap map[string]string) {
	fmt.Println(strings.Repeat("=", 40))
	fmt.Printf("Last update: %s\n\n", time.Now().Format(time.RFC3339))

	for _, metricsData := range metricsMap {
		formatAndPrintMetrics(metricsData)
	}

	fmt.Println(strings.Repeat("=", 40))
}

// Formata e imprime as métricas
func formatAndPrintMetrics(metricsData string) {
	fmt.Printf("%-30s %s\n", "Metric", "Count")
	fmt.Println(strings.Repeat("-", 40))

	trim := strings.TrimSpace(metricsData)
	lines := strings.Split(trim, "\n")

	if len(lines) > 1 {
		lines = lines[1:] // Ignorar o primeiro item se for uma linha de cabeçalho
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

// Envia métricas para todos os clientes registrados
func sendMetricsToClients(conn *net.UDPConn, metrics string, serverPort int) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for clientPort, addr := range Clients {
		if clientPort != serverPort {
			go func(addr net.UDPAddr, metrics string) {
				_, err := conn.WriteToUDP([]byte(metrics), &addr)
				if err != nil {
					fmt.Printf("Error sending data to client %d: %s\n", clientPort, err)
				}
			}(addr, metrics)
		}
	}
}
