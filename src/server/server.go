package server

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	port       int
	clients    map[int]*net.UDPAddr
	metricsMap map[string]string
	mutex      sync.RWMutex
}

func NewServer(port int) *Server {
	return &Server{
		port:       port,
		clients:    make(map[int]*net.UDPAddr),
		metricsMap: make(map[string]string),
	}
}

func (s *Server) Start() error {
	addr := net.UDPAddr{
		Port: s.port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return fmt.Errorf("error starting server: %s", err)
	}
	defer conn.Close()

	fmt.Printf("Server started on port %d\n", s.port)

	go s.listenForDiscoveryUpdates()

	buf := make([]byte, 2048)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}

		var message struct {
			Type    string `json:"type"`
			Metrics string `json:"metrics"`
		}

		if err := json.Unmarshal(buf[:n], &message); err != nil {
			fmt.Println("Error parsing message:", err)
			continue
		}

		if message.Type == "metrics" {
			s.handleMetrics(remoteAddr, message.Metrics)
		}
	}
}

func (s *Server) handleMetrics(remoteAddr *net.UDPAddr, metrics string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.metricsMap[remoteAddr.String()] = metrics

	s.printMetrics()

	// Forward metrics to all other clients
	for port, addr := range s.clients {
		if addr.String() != remoteAddr.String() {
			s.sendMetrics(port, metrics)
		}
	}
}

func (s *Server) sendMetrics(port int, metrics string) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		fmt.Printf("Error connecting to client %d: %s\n", port, err)
		return
	}
	defer conn.Close()

	message := struct {
		Type    string `json:"type"`
		Metrics string `json:"metrics"`
	}{
		Type:    "metrics",
		Metrics: metrics,
	}

	data, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Error marshaling metrics for client %d: %s\n", port, err)
		return
	}

	_, err = conn.Write(data)
	if err != nil {
		fmt.Printf("Error sending metrics to client %d: %s\n", port, err)
	}
}

func (s *Server) printMetrics() {
	fmt.Print("\033[H\033[2J")
	fmt.Printf("Last update: %s\n\n", time.Now().Format(time.RFC3339))

	for _, metricsData := range s.metricsMap {
		formatAndPrintMetrics(metricsData)
	}

	fmt.Println(strings.Repeat("=", 40))
}

func (s *Server) listenForDiscoveryUpdates() {
	addr := net.UDPAddr{
		Port: DiscoveryPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("Error listening for discovery updates:", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 2048)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving discovery update:", err)
			continue
		}

		var message struct {
			Type    string `json:"type"`
			Clients []int  `json:"clients"`
		}

		if err := json.Unmarshal(buf[:n], &message); err != nil {
			fmt.Println("Error parsing discovery update:", err)
			continue
		}

		if message.Type == "client_list" {
			s.updateClientList(message.Clients)
		}
	}
}

func (s *Server) updateClientList(clients []int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	newClients := make(map[int]*net.UDPAddr)
	for _, port := range clients {
		if port != s.port {
			newClients[port] = &net.UDPAddr{
				Port: port,
				IP:   net.ParseIP("127.0.0.1"),
			}
		}
	}

	s.clients = newClients
	fmt.Printf("Updated client list: %v\n", clients)
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

func ArrayContains(slice []int, item int) bool {
	for _, element := range slice {
		if element == item {
			return true
		}
	}
	return false
}
